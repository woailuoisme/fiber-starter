# 4. Feature 路由文件的组织与依赖注入标准

为了保持业务特性（Feature）目录的整洁度并统一系统层装配风格，我们达成以下架构决策：

1. **单独路由文件**：每个业务 Feature 必须且仅使用单独的 `routes.go` 文件来管理其 HTTP 路由配置。该文件充当该特性的“拼装胶水层”，不应将路由表和中间件声明混写于 `controllers.go` 中，以保证控制器专注在输入验证和输出映射上。
2. **统一 DI 装配方式**：在 `routes.go` 中装配 Service、Controller 的依赖时，优先统一使用 `providers.App()` 方式来获取全局已装载的强类型 Runtime 实例（直接访问 `rt.Connection`、`rt.Config` 等强类型成员），逐步废弃以 `appctx.App()` 接口获取各 `Value()` / `Store()` 后缀只读方法的设计，以直抒胸臆、减少无谓的接口层包装。
