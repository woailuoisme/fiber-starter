package providers

import (
	database "fiber-starter/internal/db"

	"go.uber.org/dig"
)

// DatabaseProvider 数据库服务提供者
type DatabaseProvider struct {
	container *dig.Container
}

// NewDatabaseProvider 创建数据库服务提供者
func NewDatabaseProvider(container *dig.Container) *DatabaseProvider {
	return &DatabaseProvider{
		container: container,
	}
}

// Register 注册数据库相关的依赖
func (p *DatabaseProvider) Register() error {
	// 注册数据库连接管理器
	if err := p.container.Provide(database.NewConnection); err != nil {
		return err
	}

	// 注意：不再直接提供 *gorm.DB 实例，以支持懒加载
	// 所有需要数据库的服务都应该注入 *database.Connection
	// 并使用 GetDB() 方法获取数据库实例

	return nil
}
