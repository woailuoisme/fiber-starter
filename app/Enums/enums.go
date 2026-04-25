package enums

import (
	"fmt"
	"slices"
	"strings"
)

// EnumInfo 通用枚举信息结构
type EnumInfo[T any] struct {
	Value    T
	Label    string
	Color    string
	Priority int
}

type BaseEnum[T any] interface {
	String() string
	Label() string
	Color() string
	Priority() int
	IsValid() bool
}

type EnumOperations[T any] interface {
	FromString(string) (T, error)
	MustFromString(string) T
	Values() []T
	GetOptions() map[string]string
	GetOptionsWithColor() map[string]map[string]string
	GetAllInfo() []EnumInfo[T]
	GetSortedInfo() []EnumInfo[T]
}

type StringEnum string
type IntEnum int

func NewStringEnum(value string) StringEnum         { return StringEnum(value) }
func (e StringEnum) String() string                 { return string(e) }
func (e StringEnum) IsValid() bool                  { return false }
func (e StringEnum) Equals(other StringEnum) bool   { return e == other }
func (e StringEnum) EqualsString(other string) bool { return string(e) == other }
func (e StringEnum) EqualsIgnoreCase(other string) bool {
	return strings.EqualFold(string(e), other)
}

func NewIntEnum(value int) IntEnum          { return IntEnum(value) }
func (e IntEnum) String() string            { return fmt.Sprintf("%d", e) }
func (e IntEnum) IsValid() bool             { return false }
func (e IntEnum) Equals(other IntEnum) bool { return e == other }
func (e IntEnum) EqualsInt(other int) bool  { return int(e) == other }

func FindInMap[T comparable](value T, infoMap map[T]EnumInfo[T]) (EnumInfo[T], bool) {
	info, exists := infoMap[value]
	return info, exists
}

func FindByLabel[T comparable](label string, infoMap map[T]EnumInfo[T]) (T, bool) {
	var zero T
	for _, info := range infoMap {
		if strings.EqualFold(info.Label, label) {
			return info.Value, true
		}
	}
	return zero, false
}

func FindByLegacyValue[T comparable](value string, legacyMap map[string]T) (T, bool) {
	var zero T
	if role, exists := legacyMap[value]; exists {
		return role, true
	}
	for legacy, role := range legacyMap {
		if strings.EqualFold(legacy, value) {
			return role, true
		}
	}
	return zero, false
}

func SortByPriority[T comparable](infos []EnumInfo[T]) []EnumInfo[T] {
	sorted := slices.Clone(infos)
	slices.SortFunc(sorted, func(a, b EnumInfo[T]) int {
		switch {
		case a.Priority < b.Priority:
			return -1
		case a.Priority > b.Priority:
			return 1
		default:
			return 0
		}
	})
	return sorted
}

func GetValuesFromInfos[T comparable](infos []EnumInfo[T]) []T {
	values := make([]T, len(infos))
	for i, info := range infos {
		values[i] = info.Value
	}
	return values
}

func GetValuesFromMap[T comparable](infoMap map[T]EnumInfo[T]) []T {
	values := make([]T, 0, len(infoMap))
	for value := range infoMap {
		values = append(values, value)
	}
	return values
}

func GetInfosFromMap[T comparable](infoMap map[T]EnumInfo[T]) []EnumInfo[T] {
	infos := make([]EnumInfo[T], 0, len(infoMap))
	for _, info := range infoMap {
		infos = append(infos, info)
	}
	return infos
}

func CreateOptionsMap[T comparable](infoMap map[T]EnumInfo[T]) map[string]string {
	options := make(map[string]string)
	for _, info := range infoMap {
		options[fmt.Sprintf("%v", info.Value)] = info.Label
	}
	return options
}

func CreateOptionsWithColorMap[T comparable](infoMap map[T]EnumInfo[T]) map[string]map[string]string {
	options := make(map[string]map[string]string)
	for _, info := range infoMap {
		options[fmt.Sprintf("%v", info.Value)] = map[string]string{
			"label": info.Label,
			"color": info.Color,
		}
	}
	return options
}

// AdminRole 管理员角色枚举
type AdminRole string

const (
	SuperAdmin AdminRole = "super_admin"
	Admin      AdminRole = "admin"
	Editor     AdminRole = "editor"
	Agent      AdminRole = "agent"
	Operator   AdminRole = "operator"
	Ops        AdminRole = "ops"
	RepWorker  AdminRole = "rep_worker"
)

var adminRoleMap = map[AdminRole]EnumInfo[AdminRole]{
	SuperAdmin: {Value: SuperAdmin, Label: "Super Admin", Color: "danger", Priority: 1},
	Admin:      {Value: Admin, Label: "Admin", Color: "warning", Priority: 2},
	Editor:     {Value: Editor, Label: "Editor", Color: "info", Priority: 3},
	Agent:      {Value: Agent, Label: "Agent", Color: "primary", Priority: 4},
	Operator:   {Value: Operator, Label: "Operator", Color: "success", Priority: 5},
	Ops:        {Value: Ops, Label: "Ops", Color: "purple", Priority: 6},
	RepWorker:  {Value: RepWorker, Label: "Customer Service", Color: "orange", Priority: 7},
}

var legacyRoleMap = map[string]AdminRole{
	"Super Admin":      SuperAdmin,
	"Admin":            Admin,
	"Editor":           Editor,
	"Agent":            Agent,
	"Ops":              Ops,
	"Operator":         Operator,
	"Customer Service": RepWorker,
}

func adminRoleInfo(role AdminRole) (EnumInfo[AdminRole], bool) { return FindInMap(role, adminRoleMap) }
func (r AdminRole) String() string {
	if info, exists := adminRoleInfo(r); exists {
		return info.Label
	}
	return string(r)
}
func (r AdminRole) Label() string { return r.String() }
func (r AdminRole) Color() string {
	if info, exists := adminRoleInfo(r); exists {
		return info.Color
	}
	return "gray"
}
func (r AdminRole) Priority() int {
	if info, exists := adminRoleInfo(r); exists {
		return info.Priority
	}
	return 999
}
func (r AdminRole) IsValid() bool { _, exists := adminRoleInfo(r); return exists }
func AdminRoleFromString(value string) (AdminRole, error) {
	role := AdminRole(strings.ToLower(value))
	if role.IsValid() {
		return role, nil
	}
	if role, exists := FindByLabel(value, adminRoleMap); exists {
		return role, nil
	}
	if role, exists := FindByLegacyValue(value, legacyRoleMap); exists {
		return role, nil
	}
	return "", fmt.Errorf("invalid admin role: %s", value)
}
func AdminRoleMustFromString(value string) AdminRole {
	role, err := AdminRoleFromString(value)
	if err != nil {
		panic(err)
	}
	return role
}
func AdminRoleValues() []AdminRole { return GetValuesFromMap(adminRoleMap) }
func AdminRoleSortedValues() []AdminRole {
	return GetValuesFromInfos(SortByPriority(GetInfosFromMap(adminRoleMap)))
}
func AdminRoleGetOptions() map[string]string { return CreateOptionsMap(adminRoleMap) }
func AdminRoleGetOptionsWithColor() map[string]map[string]string {
	return CreateOptionsWithColorMap(adminRoleMap)
}
func AdminRoleGetAllInfo() []EnumInfo[AdminRole] { return GetInfosFromMap(adminRoleMap) }
func AdminRoleGetSortedInfo() []EnumInfo[AdminRole] {
	return SortByPriority(GetInfosFromMap(adminRoleMap))
}
func (r AdminRole) HasPermission(permission string) bool {
	if r == SuperAdmin {
		return true
	}
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
func (r AdminRole) IsHigherThan(other AdminRole) bool        { return r.Priority() < other.Priority() }
func (r AdminRole) IsLowerThan(other AdminRole) bool         { return r.Priority() > other.Priority() }
func (r AdminRole) IsEqualOrHigherThan(other AdminRole) bool { return r.Priority() <= other.Priority() }

// UserStatus 用户状态枚举
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusPending   UserStatus = "pending"
	UserStatusSuspended UserStatus = "suspended"
	UserStatusBanned    UserStatus = "banned"
)

var userStatusMap = map[UserStatus]EnumInfo[UserStatus]{
	UserStatusActive:    {Value: UserStatusActive, Label: "Active", Color: "success", Priority: 1},
	UserStatusInactive:  {Value: UserStatusInactive, Label: "Inactive", Color: "secondary", Priority: 2},
	UserStatusPending:   {Value: UserStatusPending, Label: "Pending", Color: "warning", Priority: 3},
	UserStatusSuspended: {Value: UserStatusSuspended, Label: "Suspended", Color: "warning", Priority: 4},
	UserStatusBanned:    {Value: UserStatusBanned, Label: "Banned", Color: "danger", Priority: 5},
}

var userStatusLegacyMap = map[string]UserStatus{
	"Active":    UserStatusActive,
	"Inactive":  UserStatusInactive,
	"Pending":   UserStatusPending,
	"Suspended": UserStatusSuspended,
	"Banned":    UserStatusBanned,
}

func userStatusInfo(status UserStatus) (EnumInfo[UserStatus], bool) {
	return FindInMap(status, userStatusMap)
}
func (s UserStatus) String() string {
	if info, exists := userStatusInfo(s); exists {
		return info.Label
	}
	return string(s)
}
func (s UserStatus) Label() string { return s.String() }
func (s UserStatus) Color() string {
	if info, exists := userStatusInfo(s); exists {
		return info.Color
	}
	return "gray"
}
func (s UserStatus) Priority() int {
	if info, exists := userStatusInfo(s); exists {
		return info.Priority
	}
	return 999
}
func (s UserStatus) IsValid() bool   { _, exists := userStatusInfo(s); return exists }
func (s UserStatus) IsActive() bool  { return s == UserStatusActive }
func (s UserStatus) IsBlocked() bool { return s == UserStatusSuspended || s == UserStatusBanned }
func (s UserStatus) CanLogin() bool  { return s == UserStatusActive || s == UserStatusPending }
func UserStatusFromString(value string) (UserStatus, error) {
	status := UserStatus(value)
	if status.IsValid() {
		return status, nil
	}
	if status, exists := FindByLabel(value, userStatusMap); exists {
		return status, nil
	}
	if status, exists := FindByLegacyValue(value, userStatusLegacyMap); exists {
		return status, nil
	}
	return "", fmt.Errorf("invalid user status: %s", value)
}
func UserStatusMustFromString(value string) UserStatus {
	status, err := UserStatusFromString(value)
	if err != nil {
		panic(err)
	}
	return status
}
func UserStatusValues() []UserStatus { return GetValuesFromMap(userStatusMap) }
func UserStatusSortedValues() []UserStatus {
	return GetValuesFromInfos(SortByPriority(GetInfosFromMap(userStatusMap)))
}
func UserStatusGetOptions() map[string]string { return CreateOptionsMap(userStatusMap) }
func UserStatusGetOptionsWithColor() map[string]map[string]string {
	return CreateOptionsWithColorMap(userStatusMap)
}
func UserStatusGetAllInfo() []EnumInfo[UserStatus] { return GetInfosFromMap(userStatusMap) }
func UserStatusGetSortedInfo() []EnumInfo[UserStatus] {
	return SortByPriority(GetInfosFromMap(userStatusMap))
}
