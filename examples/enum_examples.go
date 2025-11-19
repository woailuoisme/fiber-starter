package main

import (
	"fmt"
)

// 1. 使用iota实现枚举（最常用）
type Status int

const (
	StatusPending   Status = iota // 0
	StatusActive                  // 1
	StatusInactive                // 2
	StatusSuspended               // 3
)

// 为枚举添加String()方法，方便打印
func (s Status) String() string {
	switch s {
	case StatusPending:
		return "Pending"
	case StatusActive:
		return "Active"
	case StatusInactive:
		return "Inactive"
	case StatusSuspended:
		return "Suspended"
	default:
		return "Unknown"
	}
}

// 2. 字符串枚举
type Role string

const (
	RoleAdmin     Role = "admin"
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
	RoleGuest     Role = "guest"
)

// 3. 位掩码枚举（用于组合权限）
type Permission int

const (
	PermissionRead   Permission = 1 << iota // 1
	PermissionWrite                         // 2
	PermissionDelete                        // 4
	PermissionAdmin                         // 8
)

// 组合权限
const (
	PermissionReadWrite = PermissionRead | PermissionWrite                                      // 3
	PermissionAll       = PermissionRead | PermissionWrite | PermissionDelete | PermissionAdmin // 15
)

// 检查权限的方法
func (p Permission) Has(permission Permission) bool {
	return p&permission == permission
}

// 4. 带验证的枚举
type Weekday int

const (
	Sunday Weekday = iota + 1
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

// 验证枚举值是否有效
func (w Weekday) IsValid() bool {
	return w >= Sunday && w <= Saturday
}

// 获取所有有效值
func (w Weekday) Values() []Weekday {
	return []Weekday{Sunday, Monday, Tuesday, Wednesday, Thursday, Friday, Saturday}
}

func main() {
	// 1. 使用iota枚举示例
	userStatus := StatusActive
	fmt.Printf("用户状态: %d (%s)\n", userStatus, userStatus.String())

	// 2. 字符串枚举示例
	userRole := RoleAdmin
	fmt.Printf("用户角色: %s\n", userRole)

	// 3. 位掩码枚举示例
	userPermission := PermissionReadWrite
	fmt.Printf("用户权限: %d\n", userPermission)
	fmt.Printf("有读权限: %t\n", userPermission.Has(PermissionRead))
	fmt.Printf("有写权限: %t\n", userPermission.Has(PermissionWrite))
	fmt.Printf("有删除权限: %t\n", userPermission.Has(PermissionDelete))

	// 4. 带验证的枚举示例
	today := Wednesday
	if today.IsValid() {
		fmt.Printf("今天是: %d\n", today)
	}

	// 遍历所有工作日
	fmt.Println("所有工作日:")
	for _, day := range today.Values() {
		fmt.Printf("- %d\n", day)
	}
}
