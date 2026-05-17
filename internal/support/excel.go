package support

import (
	"fmt"
	"io"
	"reflect"

	"github.com/gofiber/fiber/v3"
	"github.com/xuri/excelize/v2"
)

// WithHeadings 带有表头的接口
type WithHeadings interface {
	Headings() []string
}

// WithMapping 带有映射的接口
type WithMapping interface {
	Map(item any) []any
}

// FromCollection 带有数据的接口
type FromCollection interface {
	Collection() any
}

// ToModel 导入到模型的接口
type ToModel interface {
	Model(row []string) any
}

// Excel Excel工具类
type Excel struct{}

// Download 下载Excel文件
func (e *Excel) Download(c fiber.Ctx, export any, filename string) error {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	sheetName := "Sheet1"
	index, _ := f.NewSheet(sheetName)

	rowIdx := 1

	// 写入表头
	if h, ok := export.(WithHeadings); ok {
		headings := h.Headings()
		for i, heading := range headings {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowIdx)
			_ = f.SetCellValue(sheetName, cell, heading)
		}
		rowIdx++
	}

	// 写入数据
	if col, ok := export.(FromCollection); ok {
		items := col.Collection()
		v := reflect.ValueOf(items)
		if v.Kind() == reflect.Slice {
			for i := 0; i < v.Len(); i++ {
				item := v.Index(i).Interface()
				var rowData []any

				if m, ok := export.(WithMapping); ok {
					rowData = m.Map(item)
				} else {
					// 如果没有 Map 方法，尝试通过反射获取
					rowData = e.structToSlice(item)
				}

				for j, val := range rowData {
					cell, _ := excelize.CoordinatesToCellName(j+1, rowIdx)
					_ = f.SetCellValue(sheetName, cell, val)
				}
				rowIdx++
			}
		}
	}

	f.SetActiveSheet(index)

	// 设置响应头
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))

	// 将文件写入响应
	if err := f.Write(c.Response().BodyWriter()); err != nil {
		return err
	}

	return nil
}

// Import 导入Excel文件
func (e *Excel) Import(reader io.Reader, importer any) ([]any, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	rows, err := f.GetRows("Sheet1")
	if err != nil {
		return nil, err
	}

	var results []any
	if m, ok := importer.(ToModel); ok {
		// 跳过第一行（假设是表头）
		for i := 1; i < len(rows); i++ {
			row := rows[i]
			if model := m.Model(row); model != nil {
				results = append(results, model)
			}
		}
	}

	return results, nil
}

// structToSlice 将结构体转换为切片（备选方案）
func (e *Excel) structToSlice(item any) []any {
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	var res []any
	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			res = append(res, v.Field(i).Interface())
		}
	} else {
		res = append(res, item)
	}
	return res
}
