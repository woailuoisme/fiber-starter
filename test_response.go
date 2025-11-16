package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/seaside/fiber-starter/app/helpers"
)

func main() {
	fmt.Println("测试响应格式...")

	// 测试成功响应
	successResp := helpers.SuccessResponse("操作成功", map[string]interface{}{
		"user_id": 123,
		"username": "testuser",
	})

	successJSON, err := json.MarshalIndent(successResp, "", "  ")
	if err != nil {
		log.Fatal("成功响应序列化失败:", err)
	}

	fmt.Println("成功响应格式:")
	fmt.Println(string(successJSON))
	fmt.Println()

	// 测试错误响应
	errorResp := helpers.ErrorResponse("操作失败", "详细错误信息")

	errorJSON, err := json.MarshalIndent(errorResp, "", "  ")
	if err != nil {
		log.Fatal("错误响应序列化失败:", err)
	}

	fmt.Println("错误响应格式:")
	fmt.Println(string(errorJSON))
	fmt.Println()

	fmt.Println("✅ 响应格式测试完成！")
}