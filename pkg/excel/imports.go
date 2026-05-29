package excel

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"lfiber/internal/providers"

	"github.com/xuri/excelize/v2"
)

// ImportOption 定义导入的配置选项。
type ImportOption func(*importConfig)

type importConfig struct {
	sheetName string
}

// WithImportSheetName 允许在导入时指定 Sheet。
func WithImportSheetName(name string) ImportOption {
	return func(c *importConfig) {
		c.sheetName = name
	}
}

// Import 统一导入入口。根据传入的 importObj 实现的 Concern 接口，执行 Excel/CSV 导入解析。
func Import(ctx context.Context, importObj interface{}, r io.Reader, opts ...ImportOption) error {
	cfg := &importConfig{
		sheetName: "Sheet1",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// 优先使用 WithTitle 接口指定的 Sheet 名称
	if t, ok := importObj.(WithTitle); ok {
		cfg.sheetName = t.Title()
	}

	f, err := excelize.OpenReader(r)
	if err != nil {
		return fmt.Errorf("failed to open excel reader: %w", err)
	}
	defer func() { _ = f.Close() }()

	// 获取所有的行迭代器以支持流式极低内存读取
	rows, err := f.Rows(cfg.sheetName)
	if err != nil {
		return fmt.Errorf("failed to read sheet %s: %w", cfg.sheetName, err)
	}
	defer func() { _ = rows.Close() }()

	// 判定 Heading Row 索引位置
	headingRowIdx := 0
	if h, ok := importObj.(WithHeadingRow); ok {
		headingRowIdx = h.HeadingRow()
	}

	var headers []string
	currentRow := 0
	var dataRows [][]string

	// 用来收集 ToSlice 的数据
	var sliceElements []reflect.Value
	var elementType reflect.Type
	isToSlice := false

	if toSlice, ok := importObj.(ToSlice); ok {
		isToSlice = true
		destVal := reflect.ValueOf(toSlice.ToSlice())
		if destVal.Kind() != reflect.Pointer || destVal.Elem().Kind() != reflect.Slice {
			return fmt.Errorf("ToSlice() must return a pointer to a slice, got %T", toSlice.ToSlice())
		}
		elementType = destVal.Elem().Type().Elem()
		if elementType.Kind() == reflect.Pointer {
			elementType = elementType.Elem()
		}
	}

	// 批量插入上下文
	var batch []interface{}
	batchSize := 0
	isToModel := false
	if _, ok := importObj.(ToModel); ok {
		isToModel = true
		if b, ok := importObj.(WithBatchInserts); ok {
			batchSize = b.BatchSize()
		}
	}

	onRow, isOnRow := importObj.(OnRow)

	for rows.Next() {
		currentRow++
		cols, err := rows.Columns()
		if err != nil {
			return fmt.Errorf("failed to scan row %d: %w", currentRow, err)
		}

		// 跳过空行且在 Heading 行之前跳过
		if len(cols) == 0 {
			continue
		}

		if headingRowIdx > 0 && currentRow < headingRowIdx {
			continue
		}

		if headingRowIdx > 0 && currentRow == headingRowIdx {
			headers = cols
			continue
		}

		switch {
		case isOnRow:
			if err = onRow.OnRow(cols); err != nil {
				return fmt.Errorf("failed to process row %d in OnRow: %w", currentRow, err)
			}
		case isToModel:
			model, err := importObj.(ToModel).ToModel(cols)
			if err != nil {
				return fmt.Errorf("failed to convert row %d to model: %w", currentRow, err)
			}

			// 数据校验
			if v, ok := importObj.(WithValidation); ok {
				if err = v.Validate(model); err != nil {
					return fmt.Errorf("row %d validation failed: %w", currentRow, err)
				}
			}

			if batchSize > 0 {
				batch = append(batch, model)
				if len(batch) >= batchSize {
					if err = insertBatch(ctx, batch); err != nil {
						return fmt.Errorf("failed to insert batch of models: %w", err)
					}
					batch = nil
				}
			} else {
				// 单条即写入
				if err = insertBatch(ctx, []interface{}{model}); err != nil {
					return fmt.Errorf("failed to insert model at row %d: %w", currentRow, err)
				}
			}
		case isToSlice:
			// 暂时保留，行结束统一写回
			dataRows = append(dataRows, cols)
		}
	}

	// 最终冲刷可能剩余的 Batch 数据
	if isToModel && len(batch) > 0 {
		if err = insertBatch(ctx, batch); err != nil {
			return fmt.Errorf("failed to flush remaining batch of models: %w", err)
		}
	}

	// 执行 ToSlice 的序列化装填
	if isToSlice && len(dataRows) > 0 {
		destVal := reflect.ValueOf(importObj.(ToSlice).ToSlice()).Elem()
		isPtrSlice := destVal.Type().Elem().Kind() == reflect.Pointer

		for _, row := range dataRows {
			newElemPtr := reflect.New(elementType)
			newElem := newElemPtr.Elem()

			if err = mapRowToStruct(row, headers, newElem); err != nil {
				return err
			}

			// 字段校验
			if v, ok := importObj.(WithValidation); ok {
				if err = v.Validate(newElemPtr.Interface()); err != nil {
					return fmt.Errorf("data validation failed: %w", err)
				}
			}

			if isPtrSlice {
				sliceElements = append(sliceElements, newElemPtr)
			} else {
				sliceElements = append(sliceElements, newElem)
			}
		}

		// 一次性合并写入原 slice 中
		resultSlice := reflect.MakeSlice(destVal.Type(), len(sliceElements), len(sliceElements))
		for i, el := range sliceElements {
			resultSlice.Index(i).Set(el)
		}
		destVal.Set(resultSlice)
	}

	return nil
}

// insertBatch 针对 Bun Model 批量极速写入
func insertBatch(ctx context.Context, batch []interface{}) error {
	app := providers.App()
	if app == nil || app.Connection == nil {
		return fmt.Errorf("database connection is not available in providers.App()")
	}
	db, err := app.Connection.BunDB()
	if err != nil {
		return fmt.Errorf("failed to get bun db: %w", err)
	}

	// 通过 NewInsert 将批量 slice 进行 Bulk 写入
	_, err = db.NewInsert().Model(&batch).Exec(ctx)
	return err
}

// mapRowToStruct 将行数据列映射到具体 Struct 字段
func mapRowToStruct(cols []string, headers []string, dest reflect.Value) error {
	typ := dest.Type()

	headerMap := make(map[string]int)
	for i, h := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(h))] = i
	}

	for i := 0; i < dest.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" || field.Anonymous {
			continue // 忽略私有字段和匿名嵌套字段 (例如 bun.BaseModel)
		}

		colIdx := -1
		tag := field.Tag.Get("excel")
		if tag == "" {
			tag = field.Tag.Get("json")
		}

		if len(headers) > 0 && (tag != "" && tag != "-") {
			// 按 Tag 名字在 headers 寻找匹配列索引
			if idx, ok := headerMap[strings.ToLower(tag)]; ok {
				colIdx = idx
			}
		} else {
			// 如果没有表头或者没匹配到，按字段顺序分配
			if i < len(cols) {
				colIdx = i
			}
		}

		if colIdx < 0 || colIdx >= len(cols) {
			continue
		}

		cellVal := strings.TrimSpace(cols[colIdx])
		if cellVal == "" {
			continue
		}

		structField := dest.Field(i)
		if err := setFieldValue(structField, cellVal); err != nil {
			return fmt.Errorf("failed to map field %s (val: %s): %w", field.Name, cellVal, err)
		}
	}

	return nil
}

// setFieldValue 安全转换各类型并对 struct 字段属性赋值
func setFieldValue(field reflect.Value, val string) error {
	if !field.CanSet() {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(val, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal := false
		if strings.ToLower(val) == "true" || val == "1" || val == "是" {
			boolVal = true
		}
		field.SetBool(boolVal)
	case reflect.Struct:
		// 支持 time.Time 类型转换
		if field.Type() == reflect.TypeOf(time.Time{}) {
			// 支持多格式自适应解析
			layouts := []string{
				"2006-01-02 15:04:05",
				"2006-01-02",
				time.RFC3339,
			}
			var parsedTime time.Time
			var err error
			for _, l := range layouts {
				if l == time.RFC3339 {
					parsedTime, err = time.Parse(l, val)
				} else {
					parsedTime, err = time.ParseInLocation(l, val, time.Local)
				}
				if err == nil {
					break
				}
			}
			if err != nil {
				return fmt.Errorf("invalid time value %s", val)
			}
			field.Set(reflect.ValueOf(parsedTime))
		}
	case reflect.Pointer:
		// 对指针类型反射处理
		elem := reflect.New(field.Type().Elem())
		if err := setFieldValue(elem.Elem(), val); err != nil {
			return err
		}
		field.Set(elem)
	}

	return nil
}
