package search

import (
	"lfiber/configs"
)

// Register 初始化并返回 Search.Manager 与默认的 Search.Engine
// 为什么这样做：统一引导注入的构造入口，供 Runtime 拼装使用。
func Register(cfg *configs.Config) (Manager, Engine, error) {
	manager := NewManager(cfg)
	engine := manager.Drive()
	return manager, engine, nil
}
