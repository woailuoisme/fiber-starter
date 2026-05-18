// Package main is the single entry point for the fiber-starter application.
//
// Usage:
//
//	go run . serve              # Start the HTTP server
//	go run . migrate run        # Run database migrations
//	go run . routes             # Show all routes
//	go run . --help             # Show available commands
//	@title						Fiber Template API
//	@version					1.0.0
//	@description.markdown		api.md
//	@contact.name				Developer Support
//	@contact.email				dev@example.com
//	@BasePath					/
//	@schemes					http https
//	@securityDefinitions.apikey	Bearer
//	@in							header
//	@name						Authorization
//	@description				JWT 访问令牌。格式: "Bearer <Your_JWT_Token>"

// @tag.name			认证中心
// @tag.description	用户登录、注册、刷新令牌、密码找回等账号安全相关接口
// @tag.name			用户管理
// @tag.description	提供用户资料获取、修改、以及管理员级别的用户增删改查、导入导出功能
// @tag.name			系统监控
// @tag.description	服务健康检查、就绪检查及运行状态监控
package main

import command "fiber-starter/internal/console/commands"

func main() {
	command.CLI()
}
