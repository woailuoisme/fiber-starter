package medialibrary

// HasMedia 是任何需要绑定媒体文件的模型实体都必须实现的接口
type HasMedia interface {
	// GetModelType 返回模型的唯一类型标识（多态关联使用），例如 "users", "posts"
	GetModelType() string
	// GetModelID 返回模型实例的唯一主键 ID（转化为 string，支持 int64/UUID 等不同类型）
	GetModelID() string
	// RegisterMediaCollections 在模型上定义 Collections 限制（如大小、MIME 等）以及其 Conversions
	RegisterMediaCollections(reg *CollectionRegistry)
}

// MediaCollection 代表媒体库的一个分类集合约束定义
type MediaCollection struct {
	Name         string            // 集合的唯一名字，如 "avatar", "gallery"
	SingleFile   bool              // 是否是单文件覆盖模式（新上传覆盖旧文件，适用于头像等）
	AllowedMimes []string          // 允许的 MIME 格式切片，为空则不限制，如 []string{"image/jpeg", "image/png"}
	MaxFileSize  int64             // 允许的最大文件字节数，为 0 则不限制
	Conversions  []*ConversionRule // 该集合绑定的图片转换规则
}

// ConversionRule 定义了图片在上传后需要自动衍生转换出的子图（如缩略图）规格
type ConversionRule struct {
	Name   string // 转换后的命名标识，如 "thumb", "medium"
	Width  int    // 目标宽度，为 0 则按高度比例自适应缩放
	Height int    // 目标高度，为 0 则按宽度比例自适应缩放
	Crop   bool   // 是否进行居中裁剪以匹配精确宽高比
	Format string // 转换目标格式，如 "jpeg", "png"，为空则保持原图格式
}

// CollectionRegistry 是模型在注册方法里将 Collection 约束存入的注册表
type CollectionRegistry struct {
	collections map[string]*MediaCollection
}

// NewCollectionRegistry 创建一个空的注册表
func NewCollectionRegistry() *CollectionRegistry {
	return &CollectionRegistry{
		collections: make(map[string]*MediaCollection),
	}
}

// Add 注册并添加一个 Collection 规则
func (r *CollectionRegistry) Add(name string) *MediaCollection {
	col := &MediaCollection{
		Name: name,
	}
	r.collections[name] = col
	return col
}

// Get 根据名字获取一个已注册的 Collection
func (r *CollectionRegistry) Get(name string) (*MediaCollection, bool) {
	col, ok := r.collections[name]
	return col, ok
}

// SingleFileMode 将当前集合设为单文件覆盖模式
func (c *MediaCollection) SingleFileMode() *MediaCollection {
	c.SingleFile = true
	return c
}

// AcceptMimes 约束集合只接受指定的 MIME 类型
func (c *MediaCollection) AcceptMimes(mimes ...string) *MediaCollection {
	c.AllowedMimes = mimes
	return c
}

// MaxSize 约束集合只接受指定最大字节的文件
func (c *MediaCollection) MaxSize(bytes int64) *MediaCollection {
	c.MaxFileSize = bytes
	return c
}

// RegisterConversion 在集合上关联一个图片转换规则
func (c *MediaCollection) RegisterConversion(name string, width, height int, crop bool, format ...string) *MediaCollection {
	fmt := ""
	if len(format) > 0 {
		fmt = format[0]
	}
	c.Conversions = append(c.Conversions, &ConversionRule{
		Name:   name,
		Width:  width,
		Height: height,
		Crop:   crop,
		Format: fmt,
	})
	return c
}
