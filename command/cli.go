// Package command 包含所有命令行工具的实现
package command

import "fiber-starter/config"

// CLI 启动命令行工具
func CLI() {
	// 初始化配置（用于命令行工具）
	_ = config.Init()

	// 执行命令
	Execute()
}
