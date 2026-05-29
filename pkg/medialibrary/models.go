package medialibrary

import (
	"time"

	"github.com/uptrace/bun"
)

// Media 媒体文件模型
type Media struct {
	bun.BaseModel `bun:"table:media,alias:m"`

	ID               int64     `bun:"id,pk,autoincrement" json:"id"`
	ModelType        string    `bun:"model_type,notnull" json:"model_type"`
	ModelID          string    `bun:"model_id,notnull" json:"model_id"`
	UUID             string    `bun:"uuid,notnull,unique" json:"uuid"`
	CollectionName   string    `bun:"collection_name,notnull" json:"collection_name"`
	Name             string    `bun:"name,notnull" json:"name"`
	FileName         string    `bun:"file_name,notnull" json:"file_name"`
	MimeType         string    `bun:"mime_type,notnull" json:"mime_type"`
	Disk             string    `bun:"disk,notnull" json:"disk"`
	Size             int64     `bun:"size,notnull" json:"size"`
	ConversionStatus string    `bun:"conversion_status,notnull,default:'completed'" json:"conversion_status"`
	OrderColumn      *int      `bun:"order_column" json:"order_column,omitempty"`
	CreatedAt        time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt        time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`

	// JSON 数据区：用于记录子图列表、自定属性等
	Manipulations    map[string]any `bun:"manipulations,type:json" json:"manipulations"`
	CustomProperties map[string]any `bun:"custom_properties,type:json" json:"custom_properties"`
	ResponsiveImages map[string]any `bun:"responsive_images,type:json" json:"responsive_images"`
}
