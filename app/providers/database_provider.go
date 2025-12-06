package providers

import (
	"fiber-starter/database"

	"go.uber.org/dig"
	"gorm.io/gorm"
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
	// 注册数据库连接
	if err := p.container.Provide(database.NewConnection); err != nil {
		return err
	}

	// 注册GORM实例
	if err := p.container.Provide(func(conn *database.Connection) *gorm.DB {
		return conn.DB
	}); err != nil {
		return err
	}

	return nil
}
