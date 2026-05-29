package excel

import (
	"context"

	"github.com/uptrace/bun"
)

// ExportConcern 标识导出相关的 Concern 接口。
type ExportConcern interface{}

// ImportConcern 标识导入相关的 Concern 接口。
type ImportConcern interface{}

// FromSlice 允许从 Slice (切片) 导出数据。
// 用于全量导出，底层会通过反射读取数据。
type FromSlice interface {
	ExportConcern
	// FromSlice 返回待导出的切片数据。
	FromSlice() interface{}
}

// FromQuery 允许大文件低内存流式从数据库导出数据。
// 与项目核心的 Bun ORM 深度融合，利用 ScanRows 迭代写入。
type FromQuery interface {
	ExportConcern
	// FromQuery 返回用于流式导出的 Bun 查询。
	FromQuery(ctx context.Context) (*bun.SelectQuery, error)
}

// WithHeadings 提供自定义的表头定义。
type WithHeadings interface {
	ExportConcern
	// Headings 返回导出的第一行表头。
	Headings() []string
}

// WithMapping 显式定义如何将单行数据转换映射为 Excel 列。
// 如果不实现，底层将尝试通过反射自动提取 struct 属性。
type WithMapping interface {
	ExportConcern
	// Mapping 接收单行数据，返回写入 Excel 的列数据切片。
	Mapping(row interface{}) []interface{}
}

// ShouldAutoSize 允许列宽自适应。
type ShouldAutoSize interface {
	ExportConcern
	// ShouldAutoSize 返回是否需要根据内容自动计算列宽。
	ShouldAutoSize() bool
}

// WithColumnWidths 用于显式定义每一列的宽度。
type WithColumnWidths interface {
	ExportConcern
	// ColumnWidths 返回列名到宽度（字符数）的映射，例如 {"A": 20, "B": 35}。
	ColumnWidths() map[string]float64
}

// WithTitle 允许定义 Sheet 的名称。
type WithTitle interface {
	ExportConcern
	// Title 返回 Sheet 标签页的名称。
	Title() string
}

// ToSlice 允许将 Excel 中的数据解析后读入外部传入的目标 slice 中。
type ToSlice interface {
	ImportConcern
	// ToSlice 返回接收解析结果的切片指针，如 &[]User{}。
	ToSlice() interface{}
}

// ToModel 定义如何将读取到的行转换为 Bun 数据库 Model。
// 如果结合 WithBatchInserts，可进行底层的 Bulk 批量入库。
type ToModel interface {
	ImportConcern
	// ToModel 将单行字符数据转化为具体的 Model 实例，例如 *User{}。
	ToModel(row []string) (interface{}, error)
}

// OnRow 提供一种比 ToModel 更通用的低耦合逐行读取回调。
type OnRow interface {
	ImportConcern
	// OnRow 接收单行字符数据，用户可在其中执行自定义逻辑。
	OnRow(row []string) error
}

// WithHeadingRow 定义哪一行是标题行。
// 数据行将从该行之下一行开始解析，并且用于自动按列名映射属性。
type WithHeadingRow interface {
	ImportConcern
	// HeadingRow 返回标题行行号（从 1 开始计）。
	HeadingRow() int
}

// WithValidation 支持对读取出来的 Model 或每一行进行数据有效性校验。
// 为什么这样做：允许开发者在导入数据时针对已转换的 Model 结构体执行自定义校验或业务规则检查。
type WithValidation interface {
	ImportConcern
	// Validate 接收解析或转换后的数据 model，返回校验错误。如果返回 error，则认为该行校验失败。
	Validate(model any) error
}

// WithBatchInserts 允许配合 ToModel 以批量形式将数据高速插入数据库。
type WithBatchInserts interface {
	ImportConcern
	// BatchSize 返回每批批量写入的数量上限。
	BatchSize() int
}

// WithQueueNotification 允许在队列异步处理成功或失败时执行通知或收尾回调。
type WithQueueNotification interface {
	// OnQueueSuccess 在异步任务成功并上传到存储后调用，fileUrl 为生成的下载/访问链接（本地/S3）
	OnQueueSuccess(ctx context.Context, fileUrl string) error
	// OnQueueFailure 在异步任务执行失败时被调用
	OnQueueFailure(ctx context.Context, err error) error
}
