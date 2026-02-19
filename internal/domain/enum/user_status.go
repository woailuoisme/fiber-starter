package enums

import "fmt"

// UserStatus 用户状态枚举
type UserStatus string

// 定义所有用户状态常量
const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusPending   UserStatus = "pending"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusBanned    UserStatus = "banned"
)

// 用户状态信息映射表
var userStatusMap = map[UserStatus]EnumInfo[UserStatus]{
	UserStatusActive:    {Value: UserStatusActive, Label: "Active", Color: "success", Priority: 1},
	UserStatusInactive:  {Value: UserStatusInactive, Label: "Inactive", Color: "secondary", Priority: 2},
	UserStatusPending:   {Value: UserStatusPending, Label: "Pending", Color: "warning", Priority: 3},
	UserStatusSuspended: {Value: UserStatusSuspended, Label: "Suspended", Color: "warning", Priority: 4},
	UserStatusBanned:    {Value: UserStatusBanned, Label: "Banned", Color: "danger", Priority: 5},
}

// Legacy user status value mapping table for backward compatibility
var userStatusLegacyMap = map[string]UserStatus{
	"Active":    UserStatusActive,
	"Inactive":  UserStatusInactive,
	"Pending":   UserStatusPending,
	"Suspended": UserStatusSuspended,
	"Banned":    UserStatusBanned,
}

// String 实现Stringer接口
func (s UserStatus) String() string {
	if info, exists := userStatusMap[s]; exists {
		return info.Label
	}
	return string(s)
}

// Label 获取标签文本
func (s UserStatus) Label() string {
	if info, exists := userStatusMap[s]; exists {
		return info.Label
	}
	return string(s)
}

// Color 获取颜色
func (s UserStatus) Color() string {
	if info, exists := userStatusMap[s]; exists {
		return info.Color
	}
	return "gray"
}

// Priority 获取优先级
func (s UserStatus) Priority() int {
	if info, exists := userStatusMap[s]; exists {
		return info.Priority
	}
	return 999
}

// IsValid 检查状态是否有效
func (s UserStatus) IsValid() bool {
	_, exists := userStatusMap[s]
	return exists
}

// IsActive 检查是否为活跃状态
func (s UserStatus) IsActive() bool {
	return s == UserStatusActive
}

// IsBlocked 检查是否被阻止（暂停或封禁）
func (s UserStatus) IsBlocked() bool {
	return s == UserStatusSuspended || s == UserStatusBanned
}

// CanLogin 检查是否可以登录
func (s UserStatus) CanLogin() bool {
	return s == UserStatusActive || s == UserStatusPending
}

// UserStatusFromString 通过字符串值获取枚举实例
func UserStatusFromString(value string) (UserStatus, error) {
	// 首先尝试直接匹配枚举值
	status := UserStatus(value)
	if status.IsValid() {
		return status, nil
	}

	// 尝试匹配标签文本
	if status, exists := FindByLabel(value, userStatusMap); exists {
		return status, nil
	}

	// 尝试匹配旧版本值
	if status, exists := FindByLegacyValue(value, userStatusLegacyMap); exists {
		return status, nil
	}

	return "", fmt.Errorf("invalid user status: %s", value)
}

// UserStatusMustFromString 类似FromString，但遇到错误时panic
func UserStatusMustFromString(value string) UserStatus {
	status, err := UserStatusFromString(value)
	if err != nil {
		panic(err)
	}
	return status
}

// UserStatusValues 获取所有有效的状态值
func UserStatusValues() []UserStatus {
	return GetValuesFromMap(userStatusMap)
}

// UserStatusSortedValues 获取按优先级排序的状态值
func UserStatusSortedValues() []UserStatus {
	infos := GetInfosFromMap(userStatusMap)
	sortedInfos := SortByPriority(infos)

	values := make([]UserStatus, len(sortedInfos))
	for i, info := range sortedInfos {
		values[i] = info.Value
	}
	return values
}

// UserStatusGetOptions 获取所有用户状态及其标签的数组
func UserStatusGetOptions() map[string]string {
	return CreateOptionsMap(userStatusMap)
}

// UserStatusGetOptionsWithColor 获取包含颜色的选项数组
func UserStatusGetOptionsWithColor() map[string]map[string]string {
	return CreateOptionsWithColorMap(userStatusMap)
}

// UserStatusGetAllInfo 获取所有状态信息
func UserStatusGetAllInfo() []EnumInfo[UserStatus] {
	return GetInfosFromMap(userStatusMap)
}

// UserStatusGetSortedInfo 获取按优先级排序的状态信息
func UserStatusGetSortedInfo() []EnumInfo[UserStatus] {
	infos := GetInfosFromMap(userStatusMap)
	return SortByPriority(infos)
}
