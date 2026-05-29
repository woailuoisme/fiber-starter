package excel

import (
	"context"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"lfiber/internal/providers"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
)

// ExportOption 定义导出的配置选项。
type ExportOption func(*exportConfig)

type exportConfig struct {
	sheetName string
}

// WithSheetName 允许在导出时覆盖默认的 Sheet 名称。
func WithSheetName(name string) ExportOption {
	return func(c *exportConfig) {
		c.sheetName = name
	}
}

// Export 统一导出入口。根据传入的对象实现的 Concern 接口，执行对应的 Excel/CSV 导出。
func Export(ctx context.Context, export interface{}, w io.Writer, opts ...ExportOption) error {
	cfg := &exportConfig{
		sheetName: "Sheet1",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	// 优先使用 WithTitle 指定的 Sheet 名称
	if t, ok := export.(WithTitle); ok {
		cfg.sheetName = t.Title()
	}

	// 检查流式大表数据导出
	if _, isQuery := export.(FromQuery); isQuery {
		return exportStreamDirect(ctx, export, w, cfg.sheetName)
	}

	// 内存导出 (FromSlice)
	if _, isSlice := export.(FromSlice); !isSlice {
		return fmt.Errorf("export object must implement either FromSlice or FromQuery")
	}

	f := excelize.NewFile()
	defer func() { _ = f.Close() }()

	// 确保默认 Sheet 存在且更名
	origSheet := "Sheet1"
	if cfg.sheetName != origSheet {
		_, _ = f.NewSheet(cfg.sheetName)
		_ = f.DeleteSheet(origSheet)
	}

	var err error
	if err = exportMemory(ctx, export, f, cfg.sheetName); err != nil {
		return fmt.Errorf("failed to write excel data: %w", err)
	}

	// 调整列宽 (在保存前应用)
	if err = adjustColumnWidths(export, f, cfg.sheetName); err != nil {
		return fmt.Errorf("failed to adjust column widths: %w", err)
	}

	// 将 Excel 渲染输出到 writer
	if err = f.Write(w); err != nil {
		return fmt.Errorf("failed to write file to destination: %w", err)
	}

	return nil
}

// exportMemory 处理内存全量数据导出 (FromSlice)
func exportMemory(_ context.Context, export interface{}, f *excelize.File, sheetName string) error {
	fromSlice, _ := export.(FromSlice)
	sliceVal := reflect.ValueOf(fromSlice.FromSlice())

	if sliceVal.Kind() == reflect.Pointer {
		sliceVal = sliceVal.Elem()
	}
	if sliceVal.Kind() != reflect.Slice && sliceVal.Kind() != reflect.Array {
		return fmt.Errorf("FromSlice() must return a slice or array, got %s", sliceVal.Kind())
	}

	// 写入表头
	rowOffset := 1
	var headers []string
	if h, ok := export.(WithHeadings); ok {
		headers = h.Headings()
	} else if sliceVal.Len() > 0 {
		// fallback: 没有 headings 时尝试通过第一个元素的 struct 标签提取
		headers = getHeadersFromStruct(sliceVal.Index(0).Interface())
	}

	if len(headers) > 0 {
		for colIdx, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowOffset)
			_ = f.SetCellValue(sheetName, cell, header)
		}
		rowOffset++
	}

	// 写入每一行
	for i := 0; i < sliceVal.Len(); i++ {
		item := sliceVal.Index(i).Interface()
		var cols []interface{}

		if mapper, ok := export.(WithMapping); ok {
			cols = mapper.Mapping(item)
		} else {
			cols = getRowValuesFromStruct(item)
		}

		for colIdx, val := range cols {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowOffset)
			_ = f.SetCellValue(sheetName, cell, formatValue(val))
		}
		rowOffset++
	}

	return nil
}

// exportStreamDirect 流式大表数据导出，并直接输出至 output writer，规避锁拷贝问题
func exportStreamDirect(ctx context.Context, export interface{}, w io.Writer, sheetName string) error {
	fromQuery, _ := export.(FromQuery)
	query, err := fromQuery.FromQuery(ctx)
	if err != nil {
		return err
	}

	// 优先从配置库获取临时目录
	tempDir := ".cache/excel"
	if app := providers.App(); app != nil && app.Config != nil {
		if app.Config.Excel.TempPath != "" {
			tempDir = app.Config.Excel.TempPath
		}
	}
	if err = os.MkdirAll(tempDir, 0o750); err != nil {
		return err
	}

	tempFile := filepath.Join(tempDir, fmt.Sprintf("stream_%s.xlsx", uuid.New().String()))
	defer func() { _ = os.Remove(tempFile) }()

	// 保存为空文件以供 StreamWriter 读写
	f := excelize.NewFile()
	origSheet := "Sheet1"
	if sheetName != origSheet {
		_, _ = f.NewSheet(sheetName)
		_ = f.DeleteSheet(origSheet)
	}

	if err = f.SaveAs(tempFile); err != nil {
		_ = f.Close()
		return err
	}
	_ = f.Close()

	// 打开临时文件开始流式操作
	file, err := excelize.OpenFile(tempFile)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	sw, err := file.NewStreamWriter(sheetName)
	if err != nil {
		return err
	}

	rowOffset := 1
	var headers []string
	if h, ok := export.(WithHeadings); ok {
		headers = h.Headings()
	}

	// 写入表头
	if len(headers) > 0 {
		cell, _ := excelize.CoordinatesToCellName(1, rowOffset)
		vals := make([]interface{}, len(headers))
		for idx, h := range headers {
			vals[idx] = h
		}
		if err = sw.SetRow(cell, vals); err != nil {
			return err
		}
		rowOffset++
	}

	// 反射判定 Query 绑定的 Model 类型
	var rowType reflect.Type
	bunModel := query.GetModel()
	if bunModel != nil {
		val := reflect.ValueOf(bunModel.Value())
		if val.IsValid() {
			t := val.Type()
			if t.Kind() == reflect.Pointer {
				t = t.Elem()
			}
			if t.Kind() == reflect.Slice {
				rowType = t.Elem()
				if rowType.Kind() == reflect.Pointer {
					rowType = rowType.Elem()
				}
			} else if t.Kind() == reflect.Struct {
				rowType = t
			}
		}
	}

	if rowType == nil {
		return fmt.Errorf("unable to determine row type from select query, make sure Model() is called on the query")
	}

	// 获取数据库连接并开始流式 Scan
	connection := providers.App().Connection
	db, err := connection.BunDB()
	if err != nil {
		return err
	}
	rows, err := query.Rows(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	mapper, hasMapper := export.(WithMapping)

	for rows.Next() {
		rowVal := reflect.New(rowType).Interface()
		if err = db.ScanRow(ctx, rows, rowVal); err != nil {
			return err
		}

		var cols []interface{}
		if hasMapper {
			cols = mapper.Mapping(rowVal)
		} else {
			cols = getRowValuesFromStruct(rowVal)
		}

		cell, _ := excelize.CoordinatesToCellName(1, rowOffset)
		vals := make([]interface{}, len(cols))
		for idx, c := range cols {
			vals[idx] = formatValue(c)
		}
		if err = sw.SetRow(cell, vals); err != nil {
			return err
		}
		rowOffset++
	}

	if err = rows.Err(); err != nil {
		return err
	}

	if err = sw.Flush(); err != nil {
		return err
	}

	// 保存流式操作内容
	if err = file.Save(); err != nil {
		return err
	}
	_ = file.Close()

	// 重新加载以自动缩放列宽
	adjustFile, err := excelize.OpenFile(tempFile)
	if err != nil {
		return err
	}
	defer func() { _ = adjustFile.Close() }()

	if err = adjustColumnWidths(export, adjustFile, sheetName); err != nil {
		return err
	}

	if err = adjustFile.Save(); err != nil {
		return err
	}
	_ = adjustFile.Close()

	// 复制到最终输出流
	//nolint:gosec // tempFile is internally generated in safe .cache directory
	fh, err := os.Open(tempFile)
	if err != nil {
		return err
	}
	defer func() { _ = fh.Close() }()

	_, err = io.Copy(w, fh)
	return err
}

// adjustColumnWidths 调整工作表的列宽
func adjustColumnWidths(export interface{}, f *excelize.File, sheetName string) error {
	// 显式列宽设定优先
	if w, ok := export.(WithColumnWidths); ok {
		for col, width := range w.ColumnWidths() {
			_ = f.SetColWidth(sheetName, col, col, width)
		}
		return nil
	}

	// 自动适应列宽
	if autosize, ok := export.(ShouldAutoSize); ok && autosize.ShouldAutoSize() {
		cols, err := f.GetCols(sheetName)
		if err != nil {
			return err
		}
		for idx, col := range cols {
			maxLen := 0
			for _, val := range col {
				cellLen := len([]rune(val))
				if cellLen > maxLen {
					maxLen = cellLen
				}
			}
			colName, err := excelize.ColumnNumberToName(idx + 1)
			if err == nil {
				// 给一点 padding，最小宽度为 10
				w := math.Max(float64(maxLen)+3, 10)
				_ = f.SetColWidth(sheetName, colName, colName, w)
			}
		}
	}

	return nil
}

// getHeadersFromStruct 反射提取结构体的表头名称 (支持 excel 和 json tag)
func getHeadersFromStruct(item interface{}) []string {
	var headers []string
	val := reflect.ValueOf(item)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return headers
	}

	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" || field.Anonymous {
			continue // 忽略私有字段和匿名嵌套字段 (例如 bun.BaseModel)
		}
		tag := field.Tag.Get("excel")
		if tag == "" {
			tag = field.Tag.Get("json")
		}
		if tag == "" || tag == "-" {
			tag = field.Name
		}
		headers = append(headers, tag)
	}
	return headers
}

// getRowValuesFromStruct 反射提取结构体的字段值
func getRowValuesFromStruct(item interface{}) []interface{} {
	var vals []interface{}
	val := reflect.ValueOf(item)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		// 非 struct 结构直接放回单元素
		return []interface{}{item}
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		if field.PkgPath != "" || field.Anonymous {
			continue // 忽略私有字段和匿名嵌套字段 (例如 bun.BaseModel)
		}
		vals = append(vals, val.Field(i).Interface())
	}
	return vals
}

// formatValue 将各种类型规范格式化，确保输出到 Excel 的内容美观
func formatValue(val interface{}) interface{} {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case time.Time:
		if v.IsZero() {
			return ""
		}
		// 使用默认年月日时分秒格式，提升阅读感
		return v.Format("2006-01-02 15:04:05")
	case *time.Time:
		if v == nil || v.IsZero() {
			return ""
		}
		return v.Format("2006-01-02 15:04:05")
	case bool:
		if v {
			return "是"
		}
		return "否"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		// 转为数值类型保留，excelize 会以数值单元格写入
		return v
	case float32, float64:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}
