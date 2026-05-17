package requests

import (
	exceptions "fiber-starter/internal/common/exceptions"

	"github.com/gofiber/fiber/v3"
)

// BindAndValidateBody 绑定请求体并执行结构体校验。
func BindAndValidateBody(c fiber.Ctx, req interface{}) error {
	if err := c.Bind().Body(req); err != nil {
		return exceptions.BadRequestWithDetails("Invalid request body", err.Error())
	}

	return ValidateStructWithContext(c, req)
}

// BindAndValidateQuery 绑定查询参数并执行结构体校验。
func BindAndValidateQuery(c fiber.Ctx, req interface{}) error {
	if err := c.Bind().Query(req); err != nil {
		return exceptions.BadRequestWithDetails("Invalid query parameters", err.Error())
	}

	return ValidateStructWithContext(c, req)
}
