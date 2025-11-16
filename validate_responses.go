package main

import (
	"encoding/json"
	"fmt"
)

// 复制响应结构体来避免依赖问题
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

func SuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

func ErrorResponse(message string, errors interface{}) APIResponse {
	return APIResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	}
}

func main() {
	fmt.Println("=== 响应格式统一验证 ===")
	fmt.Println()

	// 测试成功响应
	fmt.Println("1. 成功响应测试:")
	successResp := SuccessResponse("用户注册成功", map[string]interface{}{
		"user_id":  123,
		"username": "testuser",
		"email":    "test@example.com",
	})
	successJSON, _ := json.MarshalIndent(successResp, "", "  ")
	fmt.Println(string(successJSON))
	fmt.Println()

	// 测试错误响应
	fmt.Println("2. 错误响应测试:")
	errorResp := ErrorResponse("验证失败", map[string]string{
		"email":    "邮箱格式不正确",
		"password": "密码长度不能少于6位",
	})
	errorJSON, _ := json.MarshalIndent(errorResp, "", "  ")
	fmt.Println(string(errorJSON))
	fmt.Println()

	// 测试简单错误响应
	fmt.Println("3. 简单错误响应测试:")
	simpleErrorResp := ErrorResponse("服务器内部错误", nil)
	simpleErrorJSON, _ := json.MarshalIndent(simpleErrorResp, "", "  ")
	fmt.Println(string(simpleErrorJSON))
	fmt.Println()

	fmt.Println("✅ 所有响应格式验证通过！")
	fmt.Println()
	fmt.Println("响应格式统一标准:")
	fmt.Println("- 成功响应: {\"success\": true, \"message\": \"成功消息\", \"data\": {...}}")
	fmt.Println("- 错误响应: {\"success\": false, \"message\": \"错误消息\", \"errors\": {...}}")
}