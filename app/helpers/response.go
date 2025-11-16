package helpers

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// FormatValidationErrors 格式化验证错误
func FormatValidationErrors(err error) map[string]string {
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

// PaginationResponse 分页响应结构
type PaginationResponse struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
	Pages int64 `json:"pages"`
}

// NewPaginationResponse 创建分页响应
func NewPaginationResponse(page, limit int, total int64) PaginationResponse {
	pages := (total + int64(limit) - 1) / int64(limit)
	return PaginationResponse{
		Page:  page,
		Limit: limit,
		Total: total,
		Pages: pages,
	}
}

// APIResponse 统一API响应结构
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// ErrorResponse 错误响应
func ErrorResponse(message string, errors interface{}) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	}
}

// SanitizeString 清理字符串
func SanitizeString(s string) string {
	// 去除首尾空格
	s = strings.TrimSpace(s)
	// 可以添加更多的清理逻辑
	return s
}

// IsValidEmail 验证邮箱格式
func IsValidEmail(email string) bool {
	// 简单的邮箱验证
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// GenerateSlug 生成URL友好的slug
func GenerateSlug(text string) string {
	// 转换为小写
	slug := strings.ToLower(text)
	// 替换空格为连字符
	slug = strings.ReplaceAll(slug, " ", "-")
	// 移除特殊字符
	slug = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return -1
	}, slug)
	
	return slug
}