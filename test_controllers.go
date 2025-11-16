package main

import (
	"encoding/json"
	"fmt"
	"fiber-starter/app/helpers"
)

// 测试控制器响应格式
func main() {
	fmt.Println("=== 控制器响应格式验证 ===\n")

	// 测试成功响应格式
	fmt.Println("1. 控制器成功响应示例:")
	successResponse := helpers.SuccessResponse("获取用户列表成功", map[string]interface{}{
		"users": []map[string]interface{}{
			{"id": 1, "name": "张三", "email": "zhangsan@example.com"},
			{"id": 2, "name": "李四", "email": "lisi@example.com"},
		},
		"pagination": map[string]interface{}{
			"page":  1,
			"limit": 10,
			"total": 2,
			"pages": 1,
		},
	})

	successJSON, _ := json.MarshalIndent(successResponse, "", "  ")
	fmt.Println(string(successJSON))

	fmt.Println("\n" + "=".repeat(50))

	// 测试错误响应格式
	fmt.Println("\n2. 控制器错误响应示例:")
	errorResponse := helpers.ErrorResponse("请求参数验证失败", map[string]interface{}{
		"name":     "姓名不能为空",
		"email":    "邮箱格式不正确",
		"password": "密码长度不能少于6位",
	})

	errorJSON, _ := json.MarshalIndent(errorResponse, "", "  ")
	fmt.Println(string(errorJSON))

	fmt.Println("\n" + "=".repeat(50))

	// 测试简单错误响应格式
	fmt.Println("\n3. 控制器简单错误响应示例:")
	simpleErrorResponse := helpers.ErrorResponse("用户不存在", nil)

	simpleErrorJSON, _ := json.MarshalIndent(simpleErrorResponse, "", "  ")
	fmt.Println(string(simpleErrorJSON))

	fmt.Println("\n✅ 控制器响应格式验证完成！")
	fmt.Println("\n📋 响应格式统一标准:")
	fmt.Println("   - 成功响应: {\"success\": true, \"message\": \"成功消息\", \"data\": {...}}")
	fmt.Println("   - 错误响应: {\"success\": false, \"message\": \"错误消息\", \"errors\": {...}}")
	fmt.Println("   - 简单错误: {\"success\": false, \"message\": \"错误消息\"}")
}

// repeat 字符串重复函数
func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}