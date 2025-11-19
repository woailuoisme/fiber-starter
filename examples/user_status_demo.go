package main

import (
	"fmt"

	"fiber-starter/app/enums"
)

func main() {
	fmt.Println("=== Go UserStatus 枚举示例 ===\n")

	// 1. 基本用法
	fmt.Println("1. 基本用法:")
	activeStatus := enums.UserStatusActive
	fmt.Printf("状态值: %s\n", string(activeStatus))
	fmt.Printf("状态标签: %s\n", activeStatus.Label())
	fmt.Printf("状态颜色: %s\n", activeStatus.Color())
	fmt.Printf("状态优先级: %d\n", activeStatus.Priority())
	fmt.Println()

	// 2. 状态检查方法
	fmt.Println("2. 状态检查方法:")
	testStatuses := []enums.UserStatus{
		enums.UserStatusActive,
		enums.UserStatusPending,
		enums.UserStatusSuspended,
		enums.UserStatusBanned,
	}

	for _, status := range testStatuses {
		fmt.Printf("%s:\n", status.Label())
		fmt.Printf("  是否活跃: %t\n", status.IsActive())
		fmt.Printf("  是否被阻止: %t\n", status.IsBlocked())
		fmt.Printf("  是否可以登录: %t\n", status.CanLogin())
		fmt.Println()
	}

	// 3. 从字符串创建枚举
	fmt.Println("3. 从字符串创建枚举:")
	testValues := []string{"active", "活跃", "pending", "invalid_status"}
	for _, value := range testValues {
		status, err := enums.UserStatusFromString(value)
		if err != nil {
			fmt.Printf("'%s' -> 错误: %v\n", value, err)
		} else {
			fmt.Printf("'%s' -> 状态: %s (%s)\n", value, status, status.Label())
		}
	}
	fmt.Println()

	// 4. 获取所有状态
	fmt.Println("4. 获取所有状态:")
	allStatuses := enums.UserStatusValues()
	fmt.Printf("所有状态值: %v\n", allStatuses)
	fmt.Println()

	// 5. 获取按优先级排序的状态
	fmt.Println("5. 按优先级排序的状态:")
	sortedStatuses := enums.UserStatusSortedValues()
	for _, status := range sortedStatuses {
		fmt.Printf("%d. %s - %s (%s)\n", status.Priority(), status, status.Label(), status.Color())
	}
	fmt.Println()

	// 6. 获取选项数组
	fmt.Println("6. 获取选项数组:")
	options := enums.UserStatusGetOptions()
	for value, label := range options {
		fmt.Printf("%s: %s\n", value, label)
	}
	fmt.Println()

	// 7. 获取包含颜色的选项
	fmt.Println("7. 获取包含颜色的选项:")
	colorOptions := enums.UserStatusGetOptionsWithColor()
	for value, info := range colorOptions {
		fmt.Printf("%s: %s (%s)\n", value, info["label"], info["color"])
	}
	fmt.Println()

	// 8. 使用MustFromString
	fmt.Println("8. 使用MustFromString:")
	validStatus := enums.UserStatusMustFromString("inactive")
	fmt.Printf("成功创建状态: %s\n", validStatus.Label())
	fmt.Println()

	// 9. 实际应用场景示例
	fmt.Println("9. 实际应用场景示例:")

	// 模拟用户登录检查
	userStatus := enums.UserStatusPending
	fmt.Printf("用户状态: %s\n", userStatus.Label())

	if !userStatus.CanLogin() {
		fmt.Printf("❌ 用户无法登录，原因: ")
		if userStatus.IsBlocked() {
			fmt.Println("账户被阻止")
		} else {
			fmt.Println("账户状态不允许登录")
		}
	} else {
		fmt.Println("✅ 用户可以登录")
	}

	// 模拟状态变更
	fmt.Println("\n模拟状态变更流程:")
	currentStatus := enums.UserStatusPending
	fmt.Printf("当前状态: %s\n", currentStatus.Label())

	// 审核通过
	currentStatus = enums.UserStatusActive
	fmt.Printf("审核后状态: %s\n", currentStatus.Label())

	// 违规暂停
	currentStatus = enums.UserStatusSuspended
	fmt.Printf("违规后状态: %s\n", currentStatus.Label())
	fmt.Printf("是否被阻止: %t\n", currentStatus.IsBlocked())
}
