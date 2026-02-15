// Package enums 定义应用程序中使用的枚举类型
package enums

import (
	"fmt"
	"strings"
)

// AdminRole 管理员角色枚举
type AdminRole string

// 定义所有管理员角色常量
const (
	SuperAdmin AdminRole = "super_admin"
	Admin      AdminRole = "admin"
	Editor     AdminRole = "editor"
	Agent      AdminRole = "agent"
	Operator   AdminRole = "operator"
	Ops        AdminRole = "ops"
	RepWorker  AdminRole = "rep_worker"
)

// 角色信息映射表
var adminRoleMap = map[AdminRole]EnumInfo[AdminRole]{
	SuperAdmin: {Value: SuperAdmin, Label: "超级管理员", Color: "danger", Priority: 1},
	Admin:      {Value: Admin, Label: "管理员", Color: "warning", Priority: 2},
	Editor:     {Value: Editor, Label: "编辑员", Color: "info", Priority: 3},
	Agent:      {Value: Agent, Label: "代理员", Color: "primary", Priority: 4},
	Operator:   {Value: Operator, Label: "补货员", Color: "success", Priority: 5},
	Ops:        {Value: Ops, Label: "运维员", Color: "purple", Priority: 6},
	RepWorker:  {Value: RepWorker, Label: "客服", Color: "orange", Priority: 7},
}

// 兼容旧版本值的映射表
var legacyRoleMap = map[string]AdminRole{
	"超级管理员": SuperAdmin,
	"管理员":   Admin,
	"编辑员":   Editor,
	"代理员":   Agent,
	"运维员":   Ops,
	"补货员":   Operator,
	"客服":    RepWorker,
}

// String 实现Stringer接口
func (r AdminRole) String() string {
	if info, exists := FindInMap(r, adminRoleMap); exists {
		return info.Label
	}
	return string(r)
}

// Label 获取标签文本（类似PHP的getLabelText）
func (r AdminRole) Label() string {
	if info, exists := FindInMap(r, adminRoleMap); exists {
		return info.Label
	}
	return string(r)
}

// Color 获取颜色（类似PHP的getColor）
func (r AdminRole) Color() string {
	if info, exists := FindInMap(r, adminRoleMap); exists {
		return info.Color
	}
	return "gray"
}

// Priority 获取优先级
func (r AdminRole) Priority() int {
	if info, exists := FindInMap(r, adminRoleMap); exists {
		return info.Priority
	}
	return 999
}

// IsValid 检查角色是否有效
func (r AdminRole) IsValid() bool {
	_, exists := FindInMap(r, adminRoleMap)
	return exists
}

// AdminRoleFromString FromString 通过字符串值获取枚举实例（类似PHP的fromString）
func AdminRoleFromString(value string) (AdminRole, error) {
	// 首先尝试直接匹配枚举值
	role := AdminRole(strings.ToLower(value))
	if role.IsValid() {
		return role, nil
	}

	// 尝试匹配标签文本
	if role, exists := FindByLabel(value, adminRoleMap); exists {
		return role, nil
	}

	// 尝试匹配旧版本值
	if role, exists := FindByLegacyValue(value, legacyRoleMap); exists {
		return role, nil
	}

	return "", fmt.Errorf("invalid admin role: %s", value)
}

// AdminRoleMustFromString MustFromString 类似FromString，但遇到错误时panic
func AdminRoleMustFromString(value string) AdminRole {
	role, err := AdminRoleFromString(value)
	if err != nil {
		panic(err)
	}
	return role
}

// AdminRoleValues Values 获取所有有效的角色值（类似PHP的cases）
func AdminRoleValues() []AdminRole {
	return GetValuesFromMap(adminRoleMap)
}

// AdminRoleSortedValues SortedValues 获取按优先级排序的角色值
func AdminRoleSortedValues() []AdminRole {
	infos := GetInfosFromMap(adminRoleMap)
	sortedInfos := SortByPriority(infos)

	values := make([]AdminRole, len(sortedInfos))
	for i, info := range sortedInfos {
		values[i] = info.Value
	}
	return values
}

// AdminRoleGetOptions GetOptions 获取所有管理员角色及其标签的数组（类似PHP的getOptions）
func AdminRoleGetOptions() map[string]string {
	return CreateOptionsMap(adminRoleMap)
}

// AdminRoleGetOptionsWithColor GetOptionsWithColor 获取包含颜色的选项数组
func AdminRoleGetOptionsWithColor() map[string]map[string]string {
	return CreateOptionsWithColorMap(adminRoleMap)
}

// AdminRoleGetAllInfo GetAllInfo 获取所有角色信息
func AdminRoleGetAllInfo() []EnumInfo[AdminRole] {
	return GetInfosFromMap(adminRoleMap)
}

// AdminRoleGetSortedInfo GetSortedInfo 获取按优先级排序的角色信息
func AdminRoleGetSortedInfo() []EnumInfo[AdminRole] {
	infos := GetInfosFromMap(adminRoleMap)
	return SortByPriority(infos)
}

// HasPermission 检查是否有权限（可以扩展为更复杂的权限系统）
func (r AdminRole) HasPermission(permission string) bool {
	// 这里可以根据实际需求实现权限检查逻辑
	// 例如：超级管理员拥有所有权限
	if r == SuperAdmin {
		return true
	}

	// 其他角色的权限检查逻辑
	switch r {
	case Admin:
		return permission != "system_config"
	case Editor:
		return permission == "content_edit" || permission == "content_view"
	case Operator:
		return permission == "inventory_manage" || permission == "inventory_view"
	case Agent:
		return permission == "agent_manage" || permission == "agent_view"
	case Ops:
		return permission == "system_ops" || permission == "system_monitor"
	case RepWorker:
		return permission == "customer_service" || permission == "customer_view"
	default:
		return false
	}
}

// IsHigherThan 检查当前角色是否比指定角色级别更高
func (r AdminRole) IsHigherThan(other AdminRole) bool {
	return r.Priority() < other.Priority() // 优先级数字越小，级别越高
}

// IsLowerThan 检查当前角色是否比指定角色级别更低
func (r AdminRole) IsLowerThan(other AdminRole) bool {
	return r.Priority() > other.Priority()
}

// IsEqualOrHigherThan 检查当前角色是否等于或高于指定角色
func (r AdminRole) IsEqualOrHigherThan(other AdminRole) bool {
	return r.Priority() <= other.Priority()
}
