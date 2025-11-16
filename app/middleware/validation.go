package middleware

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// 验证器实例
var validate = validator.New()

// 自定义验证器注册
func init() {
	// 注册自定义验证标签
	validate.RegisterValidation("phone", validatePhone)
}

// validatePhone 自定义手机号验证
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	// 简单的手机号验证，可以根据需要调整
	return len(phone) >= 10 && len(phone) <= 15
}

// ValidationMiddleware 验证中间件
func ValidationMiddleware(model interface{}) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 获取请求体类型
		contentType := c.Get("Content-Type")
		
		if strings.Contains(contentType, "application/json") {
			// JSON 请求体验证
			if err := c.BodyParser(model); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": "请求参数解析失败",
					"error":   err.Error(),
				})
			}
		} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			// 表单数据验证
			if err := c.BodyParser(model); err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"success": false,
					"message": "表单数据解析失败",
					"error":   err.Error(),
				})
			}
		}

		// 验证模型
		if err := validate.Struct(model); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "请求参数验证失败",
				"errors":  formatValidationErrors(err),
			})
		}

		// 将验证后的模型存储到上下文中
		c.Locals("validated_model", model)
		
		return c.Next()
	}
}

// formatValidationErrors 格式化验证错误
func formatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()
			
			switch tag {
			case "required":
				errors[field] = fmt.Sprintf("%s 是必填字段", field)
			case "min":
				errors[field] = fmt.Sprintf("%s 长度不能少于 %s", field, e.Param())
			case "max":
				errors[field] = fmt.Sprintf("%s 长度不能超过 %s", field, e.Param())
			case "email":
				errors[field] = fmt.Sprintf("%s 必须是有效的邮箱地址", field)
			case "e164":
				errors[field] = fmt.Sprintf("%s 必须是有效的手机号码", field)
			case "url":
				errors[field] = fmt.Sprintf("%s 必须是有效的URL", field)
			case "phone":
				errors[field] = fmt.Sprintf("%s 必须是有效的手机号码", field)
			default:
				errors[field] = fmt.Sprintf("%s 格式不正确", field)
			}
		}
	}
	
	return errors
}

// GetValidatedModel 从上下文中获取验证后的模型
func GetValidatedModel(c *fiber.Ctx, model interface{}) bool {
	if validatedModel := c.Locals("validated_model"); validatedModel != nil {
		// 使用反射复制值
		srcValue := reflect.ValueOf(validatedModel)
		dstValue := reflect.ValueOf(model)
		
		if srcValue.Kind() == reflect.Ptr && dstValue.Kind() == reflect.Ptr {
			srcValue = srcValue.Elem()
			dstValue = dstValue.Elem()
			
			if srcValue.Type().AssignableTo(dstValue.Type()) {
				dstValue.Set(srcValue)
				return true
			}
		}
	}
	return false
}

// QueryValidationMiddleware 查询参数验证中间件
func QueryValidationMiddleware(rules map[string]string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		errors := make(map[string]string)
		
		for field, rule := range rules {
			value := c.Query(field)
			
			// 检查必填字段
			if strings.Contains(rule, "required") && value == "" {
				errors[field] = fmt.Sprintf("%s 是必填字段", field)
				continue
			}
			
			// 如果字段为空且不是必填，跳过其他验证
			if value == "" {
				continue
			}
			
			// 验证数字
			if strings.Contains(rule, "number") {
				var num int64
				if _, err := fmt.Sscanf(value, "%d", &num); err != nil {
					errors[field] = fmt.Sprintf("%s 必须是数字", field)
				}
			}
			
			// 验证最小长度
			if strings.Contains(rule, "min:") {
				minLength := 0
				fmt.Sscanf(strings.Split(rule, "min:")[1], "%d", &minLength)
				if len(value) < minLength {
					errors[field] = fmt.Sprintf("%s 长度不能少于 %d", field, minLength)
				}
			}
			
			// 验证最大长度
			if strings.Contains(rule, "max:") {
				maxLength := 0
				fmt.Sscanf(strings.Split(rule, "max:")[1], "%d", &maxLength)
				if len(value) > maxLength {
					errors[field] = fmt.Sprintf("%s 长度不能超过 %d", field, maxLength)
				}
			}
		}
		
		if len(errors) > 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "查询参数验证失败",
				"errors":  errors,
			})
		}
		
		return c.Next()
	}
}

// FileValidationMiddleware 文件验证中间件
func FileValidationMiddleware(maxSize int64, allowedTypes []string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": "文件上传失败",
				"error":   err.Error(),
			})
		}
		
		// 验证文件大小
		if file.Size > maxSize {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": fmt.Sprintf("文件大小不能超过 %d MB", maxSize/(1024*1024)),
			})
		}
		
		// 验证文件类型
		contentType := file.Header.Get("Content-Type")
		isAllowed := false
		for _, allowedType := range allowedTypes {
			if contentType == allowedType {
				isAllowed = true
				break
			}
		}
		
		if !isAllowed {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"message": fmt.Sprintf("不支持的文件类型: %s", contentType),
			})
		}
		
		// 将文件信息存储到上下文中
		c.Locals("uploaded_file", file)
		
		return c.Next()
	}
}