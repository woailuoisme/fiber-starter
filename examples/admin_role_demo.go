package main

import (
	"fmt"

	"fiber-starter/app/enums"
)

func main() {
	fmt.Println("=== Go AdminRole 枚举示例 ===\n")

	// 1. 基本用法
	fmt.Println("1. 基本用法:")
	superAdmin := enums.SuperAdmin
	fmt.Printf("角色值: %s\n", string(superAdmin))
	fmt.Printf("角色标签: %s\n", superAdmin.Label())
	fmt.Printf("角色颜色: %s\n", superAdmin.Color())
	fmt.Printf("角色优先级: %d\n", superAdmin.Priority())
	fmt.Println()

	// 2. 从字符串创建枚举
	fmt.Println("2. 从字符串创建枚举:")
	testValues := []string{"admin", "管理员", "editor", "invalid_role"}
	for _, value := range testValues {
		role, err := enums.AdminRoleFromString(value)
		if err != nil {
			fmt.Printf("'%s' -> 错误: %v\n", value, err)
		} else {
			fmt.Printf("'%s' -> 角色: %s (%s)\n", value, role, role.Label())
		}
	}
	fmt.Println()

	// 3. 获取所有角色
	fmt.Println("3. 获取所有角色:")
	allRoles := enums.AdminRoleValues()
	fmt.Printf("所有角色值: %v\n", allRoles)
	fmt.Println()

	// 4. 获取按优先级排序的角色
	fmt.Println("4. 按优先级排序的角色:")
	sortedRoles := enums.AdminRoleSortedValues()
	for _, role := range sortedRoles {
		fmt.Printf("%d. %s - %s (%s)\n", role.Priority(), role, role.Label(), role.Color())
	}
	fmt.Println()

	// 5. 获取选项数组
	fmt.Println("5. 获取选项数组:")
	options := enums.AdminRoleGetOptions()
	for value, label := range options {
		fmt.Printf("%s: %s\n", value, label)
	}
	fmt.Println()

	// 6. 获取包含颜色的选项
	fmt.Println("6. 获取包含颜色的选项:")
	colorOptions := enums.AdminRoleGetOptionsWithColor()
	for value, info := range colorOptions {
		fmt.Printf("%s: %s (%s)\n", value, info["label"], info["color"])
	}
	fmt.Println()

	// 7. 权限检查
	fmt.Println("7. 权限检查:")
	testRoles := []enums.AdminRole{enums.SuperAdmin, enums.Admin, enums.Editor}
	testPermissions := []string{"system_config", "content_edit", "user_manage"}

	for _, role := range testRoles {
		fmt.Printf("\n%s 的权限:\n", role.Label())
		for _, permission := range testPermissions {
			fmt.Printf("  %s: %t\n", permission, role.HasPermission(permission))
		}
	}
	fmt.Println()

	// 8. 角色级别比较
	fmt.Println("8. 角色级别比较:")
	admin := enums.Admin
	editor := enums.Editor

	fmt.Printf("%s 是否高于 %s: %t\n", admin.Label(), editor.Label(), admin.IsHigherThan(editor))
	fmt.Printf("%s 是否低于 %s: %t\n", admin.Label(), editor.Label(), admin.IsLowerThan(editor))
	fmt.Printf("%s 是否等于或高于 %s: %t\n", admin.Label(), editor.Label(), admin.IsEqualOrHigherThan(editor))
	fmt.Println()

	// 9. 角色验证
	fmt.Println("9. 角色验证:")
	testRole := enums.Operator
	fmt.Printf("%s 是否有效: %t\n", testRole, testRole.IsValid())

	var invalidRole enums.AdminRole = "invalid"
	fmt.Printf("%s 是否有效: %t\n", invalidRole, invalidRole.IsValid())
	fmt.Println()

	// 10. 使用MustFromString（会panic的错误示例）
	fmt.Println("10. 使用MustFromString:")
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("捕获到panic: %v\n", r)
		}
	}()

	// 这个会成功
	validRole := enums.AdminRoleMustFromString("agent")
	fmt.Printf("成功创建角色: %s\n", validRole.Label())

	// 这个会panic（已注释掉，避免程序中断）
	// invalidRole := enums.AdminRoleMustFromString("invalid_role")
	// fmt.Printf("这行不会执行: %s\n", invalidRole.Label())
}
