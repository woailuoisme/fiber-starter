package middleware

import (
	"time"

	helpers "fiber-starter/app/Support"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/timeout"
)

// TimeoutHandler 返回带超时保护的路由 handler。
// 作用：限制慢请求占用资源，避免单个请求拖垮服务。
// 场景：数据库访问、外部 API 调用、长耗时计算等后端接口。
// 使用方式：在最终业务 handler 外层包裹，不要挂到 app.Use 或中间件链中。
func TimeoutHandler(handler fiber.Handler) fiber.Handler {
	return TimeoutHandlerWithDuration(handler, 30*time.Second)
}

// TimeoutHandlerWithDuration 返回带自定义超时的路由 handler。
// 作用：为不同路由组提供差异化超时控制。
// 场景：登录、查询、写操作等接口需要不同响应时限时使用。
// 使用方式：作为最终业务 handler 的最外层包装。
func TimeoutHandlerWithDuration(handler fiber.Handler, duration time.Duration) fiber.Handler {
	return timeout.New(func(c fiber.Ctx) error {
		return handler(c)
	}, timeout.Config{
		Timeout: duration,
		OnTimeout: func(c fiber.Ctx) error {
			return helpers.HandleHTTPError(c, fiber.ErrRequestTimeout)
		},
	})
}

// TimeoutRouter 为一组路由提供统一的超时包装。
// 作用：减少每条路由手动包 TimeoutHandler 的重复代码。
// 场景：业务路由组需要统一超时策略，但仍要保留 Fiber 的路由声明风格时使用。
// 使用方式：先用 NewTimeoutRouter 包装 router/group，再像普通 Router 一样注册路由；最后一个 handler 会自动加超时。
type TimeoutRouter struct {
	router   fiber.Router
	duration time.Duration
}

// NewTimeoutRouter 创建统一超时的路由包装器。
func NewTimeoutRouter(router fiber.Router, duration time.Duration) *TimeoutRouter {
	if duration <= 0 {
		duration = 30 * time.Second
	}

	return &TimeoutRouter{
		router:   router,
		duration: duration,
	}
}

type routeRegister func(path string, handler any, handlers ...any) fiber.Router

func (r *TimeoutRouter) register(register routeRegister, path string, handlers ...fiber.Handler) fiber.Router {
	if len(handlers) == 0 {
		panic("timeout router requires at least one handler")
	}

	args := make([]any, 0, len(handlers))
	for _, handler := range handlers[:len(handlers)-1] {
		args = append(args, handler)
	}
	args = append(args, TimeoutHandlerWithDuration(handlers[len(handlers)-1], r.duration))

	if len(args) == 1 {
		return register(path, args[0])
	}

	return register(path, args[0], args[1:]...)
}

// Get 注册 GET 路由，最后一个 handler 自动获得超时保护。
func (r *TimeoutRouter) Get(path string, handlers ...fiber.Handler) fiber.Router {
	return r.register(r.router.Get, path, handlers...)
}

// Post 注册 POST 路由，最后一个 handler 自动获得超时保护。
func (r *TimeoutRouter) Post(path string, handlers ...fiber.Handler) fiber.Router {
	return r.register(r.router.Post, path, handlers...)
}

// Put 注册 PUT 路由，最后一个 handler 自动获得超时保护。
func (r *TimeoutRouter) Put(path string, handlers ...fiber.Handler) fiber.Router {
	return r.register(r.router.Put, path, handlers...)
}

// Delete 注册 DELETE 路由，最后一个 handler 自动获得超时保护。
func (r *TimeoutRouter) Delete(path string, handlers ...fiber.Handler) fiber.Router {
	return r.register(r.router.Delete, path, handlers...)
}

// Patch 注册 PATCH 路由，最后一个 handler 自动获得超时保护。
func (r *TimeoutRouter) Patch(path string, handlers ...fiber.Handler) fiber.Router {
	return r.register(r.router.Patch, path, handlers...)
}

// Head 注册 HEAD 路由，最后一个 handler 自动获得超时保护。
func (r *TimeoutRouter) Head(path string, handlers ...fiber.Handler) fiber.Router {
	return r.register(r.router.Head, path, handlers...)
}

// Options 注册 OPTIONS 路由，最后一个 handler 自动获得超时保护。
func (r *TimeoutRouter) Options(path string, handlers ...fiber.Handler) fiber.Router {
	return r.register(r.router.Options, path, handlers...)
}

// Trace 注册 TRACE 路由，最后一个 handler 自动获得超时保护。
func (r *TimeoutRouter) Trace(path string, handlers ...fiber.Handler) fiber.Router {
	return r.register(r.router.Trace, path, handlers...)
}

// Connect 注册 CONNECT 路由，最后一个 handler 自动获得超时保护。
func (r *TimeoutRouter) Connect(path string, handlers ...fiber.Handler) fiber.Router {
	return r.register(r.router.Connect, path, handlers...)
}

// All 注册 ALL 路由，最后一个 handler 自动获得超时保护。
func (r *TimeoutRouter) All(path string, handlers ...fiber.Handler) fiber.Router {
	return r.register(r.router.All, path, handlers...)
}
