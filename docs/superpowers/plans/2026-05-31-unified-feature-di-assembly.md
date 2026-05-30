# Unified Feature DI Assembly Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 `user` 和 `auth` 业务 Feature 中的路由依赖装配（DI）从 `appctx.App()` 接口化获取迁移为 `providers.App()` 强类型方式，实现架构标准统一，清理无谓包装。

**Architecture:** 重构 features 下的 `routes.go`，用强类型 Runtime 替换只读抽象接口的调用（如 `rt.Connection` 代替 `rt.ConnectionValue()` 等）。

**Tech Stack:** Go 1.26.3, Go Fiber v3, testify

---

### Task 1: 重构 `user` 模块路由依赖注入为强类型

**Files:**
- Modify: `internal/features/user/routes.go:1-40`
- Test: `tests/integration/http/user_endpoints_test.go`

- [ ] **Step 1: 运行现有测试进行编译与基准验证**

运行：`rtk go test ./tests/integration/http/user_endpoints_test.go -run "TestUserEndpoints_CurrentUserAndList" -v`
Expected: 测试被成功编译（可能报错 400，但不能有语法或编译错误）。

- [ ] **Step 2: 重构 `internal/features/user/routes.go` 文件**

修改 `internal/features/user/routes.go` 内容为下述无 `appctx` 依赖的强类型版本：

```go
package user

import (
	"time"

	middleware "lfiber/internal/common/middleware"
	providers "lfiber/internal/providers"
	authorization "lfiber/internal/providers/authorization"

	"github.com/gofiber/fiber/v3"
)

const routeTimeout = 30 * time.Second

// RegisterRoutes registers user routes under the provided router group.
func RegisterRoutes(router fiber.Router) {
	rt := providers.App()

	userService := NewUserService(rt.Connection)
	userDataExchange := NewUserDataExchange(rt.Connection)
	userController := NewUserController(userService, userDataExchange)
	jwtProtected := middleware.JWTProtected(rt.Config, rt.Cache)
	requirePermission := authorization.RequirePermissions

	usersRouter := middleware.NewTimeoutRouter(
		router.Group("/users", middleware.IdempotencyMiddleware()),
		routeTimeout,
	)

	usersRouter.Get("/", jwtProtected, requirePermission("users:list"), userController.GetUsers)
	usersRouter.Get("/me", jwtProtected, userController.GetCurrentUser)
	usersRouter.Get("/search", jwtProtected, requirePermission("users:read"), userController.SearchUsers)
	usersRouter.Put("/profile", jwtProtected, userController.UpdateProfile)
	usersRouter.Get("/export", jwtProtected, requirePermission("users:export"), userController.ExportUsers)
	usersRouter.Post("/import", jwtProtected, requirePermission("users:import"), userController.ImportUsers)
	usersRouter.Get("/:id", jwtProtected, requirePermission("users:read"), userController.GetUser)
	usersRouter.Put("/:id", jwtProtected, requirePermission("users:update"), userController.UpdateUser)
	usersRouter.Delete("/:id", jwtProtected, requirePermission("users:delete"), userController.DeleteUser)
}
```

- [ ] **Step 3: 重新运行测试以验证重构后编译和逻辑是否正常**

运行：`rtk go test ./tests/integration/http/user_endpoints_test.go -run "TestUserEndpoints_CurrentUserAndList" -v`
Expected: 测试被成功编译。

- [ ] **Step 4: Commit 提交**

```bash
git add internal/features/user/routes.go
git commit -m "refactor(user): migrate route di assembly to providers.App()"
```

---

### Task 2: 重构 `auth` 模块路由依赖注入为强类型

**Files:**
- Modify: `internal/features/auth/routes.go:1-43`
- Test: `tests/integration/http/auth_endpoints_test.go`

- [ ] **Step 1: 运行现有测试进行编译与基准验证**

运行：`rtk go test ./tests/integration/http/auth_endpoints_test.go -run "TestAuthEndpoints_SignUpAndLogin" -v`
Expected: 测试被成功编译（可能报错，但不能有语法编译错误）。

- [ ] **Step 2: 重构 `internal/features/auth/routes.go` 文件**

修改 `internal/features/auth/routes.go` 内容为下述强类型版本：

```go
package auth

import (
	"time"

	middleware "lfiber/internal/common/middleware"
	providers "lfiber/internal/providers"

	"github.com/gofiber/fiber/v3"
)

const routeTimeout = 30 * time.Second

// RegisterRoutes registers auth routes under the provided router group.
func RegisterRoutes(router fiber.Router) {
	rt := providers.App()

	authService := NewAuthService(
		rt.Connection,
		rt.Config,
		rt.Cache,
		rt.EmailService,
	)
	authController := NewAuthController(authService)
	jwtProtected := middleware.JWTProtected(rt.Config, rt.Cache)

	authRouter := middleware.NewTimeoutRouter(
		router.Group("/auth"),
		routeTimeout,
	)

	authRouter.Post("/sign-up", authController.SignUp)
	authRouter.Post("/sign-up/verify", authController.VerifySignUp)
	authRouter.Post("/sign-in", authController.SignIn)
	authRouter.Post("/refresh", authController.RefreshSession)
	authRouter.Post("/sign-out", jwtProtected, authController.SignOut)
	authRouter.Post("/change-password", jwtProtected, authController.UpdatePassword)
	authRouter.Post("/reset-password", authController.SendPasswordReset)
	authRouter.Post("/reset-password/verify", authController.VerifyPasswordReset)
	authRouter.Post("/reset-password/confirm", authController.ConfirmPasswordReset)
	authRouter.Get("/session", jwtProtected, authController.Session)
}
```

- [ ] **Step 3: 重新运行测试以验证重构后编译和逻辑是否正常**

运行：`rtk go test ./tests/integration/http/auth_endpoints_test.go -run "TestAuthEndpoints_SignUpAndLogin" -v`
Expected: 测试被成功编译。

- [ ] **Step 4: Commit 提交**

```bash
git add internal/features/auth/routes.go
git commit -m "refactor(auth): migrate route di assembly to providers.App()"
```

---

### Task 3: 全局静态检查与质量验收

**Files:**
- None (Verification task)

- [ ] **Step 1: 运行 `rtk just check` 确保符合 Lint 规范**

运行：`rtk just check`
Expected: `0 issues` 编译及静态 Lint 全部通过。
