package enums

import (
	"fmt"
	"strings"
)

// EnumInfo 通用枚举信息结构
type EnumInfo[T any] struct {
	Value    T
	Label    string
	Color    string
	Priority int
}

// BaseEnum 基础枚举接口
type BaseEnum[T any] interface {
	String() string
	Label() string
	Color() string
	Priority() int
	IsValid() bool
}

// EnumOperations 枚举操作接口
type EnumOperations[T any] interface {
	FromString(string) (T, error)
	MustFromString(string) T
	Values() []T
	GetOptions() map[string]string
	GetOptionsWithColor() map[string]map[string]string
	GetAllInfo() []EnumInfo[T]
	GetSortedInfo() []EnumInfo[T]
}

// StringEnum 字符串枚举的基础实现
type StringEnum string

// NewStringEnum 创建新的字符串枚举
func NewStringEnum(value string) StringEnum {
	return StringEnum(value)
}

// String 实现Stringer接口
func (e StringEnum) String() string {
	return string(e)
}

// IsValid 检查枚举值是否有效（需要在具体枚举中重写）
func (e StringEnum) IsValid() bool {
	return false // 基础实现，子类需要重写
}

// Equals 检查是否等于另一个枚举值
func (e StringEnum) Equals(other StringEnum) bool {
	return e == other
}

// EqualsString 检查是否等于字符串值
func (e StringEnum) EqualsString(other string) bool {
	return string(e) == other
}

// EqualsIgnoreCase 检查是否等于字符串值（忽略大小写）
func (e StringEnum) EqualsIgnoreCase(other string) bool {
	return strings.EqualFold(string(e), other)
}

// IntEnum 整数枚举的基础实现
type IntEnum int

// NewIntEnum 创建新的整数枚举
func NewIntEnum(value int) IntEnum {
	return IntEnum(value)
}

// String 实现Stringer接口
func (e IntEnum) String() string {
	return fmt.Sprintf("%d", e)
}

// IsValid 检查枚举值是否有效（需要在具体枚举中重写）
func (e IntEnum) IsValid() bool {
	return false // 基础实现，子类需要重写
}

// Equals 检查是否等于另一个枚举值
func (e IntEnum) Equals(other IntEnum) bool {
	return e == other
}

// EqualsInt 检查是否等于整数值
func (e IntEnum) EqualsInt(other int) bool {
	return int(e) == other
}

// EnumUtils 枚举工具函数
type EnumUtils struct{}

// FindInMap 在映射表中查找枚举信息
func FindInMap[T comparable](value T, infoMap map[T]EnumInfo[T]) (EnumInfo[T], bool) {
	info, exists := infoMap[value]
	return info, exists
}

// FindByLabel 在映射表中通过标签查找枚举值
func FindByLabel[T comparable](label string, infoMap map[T]EnumInfo[T]) (T, bool) {
	var zero T
	for _, info := range infoMap {
		if strings.EqualFold(info.Label, label) {
			return info.Value, true
		}
	}
	return zero, false
}

// FindByLegacyValue 在遗留值映射表中查找枚举值
func FindByLegacyValue[T comparable](value string, legacyMap map[string]T) (T, bool) {
	var zero T
	if role, exists := legacyMap[value]; exists {
		return role, true
	}

	// 尝试不区分大小写匹配
	lowerValue := strings.ToLower(value)
	for legacy, role := range legacyMap {
		if strings.EqualFold(legacy, lowerValue) {
			return role, true
		}
	}

	return zero, false
}

// SortByPriority 按优先级排序枚举信息
func SortByPriority[T comparable](infos []EnumInfo[T]) []EnumInfo[T] {
	sorted := make([]EnumInfo[T], len(infos))
	copy(sorted, infos)

	// 简单的冒泡排序
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].Priority > sorted[j].Priority {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted
}

// GetValuesFromMap 从映射表获取所有枚举值
func GetValuesFromMap[T comparable](infoMap map[T]EnumInfo[T]) []T {
	values := make([]T, 0, len(infoMap))
	for value := range infoMap {
		values = append(values, value)
	}
	return values
}

// GetInfosFromMap 从映射表获取所有枚举信息
func GetInfosFromMap[T comparable](infoMap map[T]EnumInfo[T]) []EnumInfo[T] {
	infos := make([]EnumInfo[T], 0, len(infoMap))
	for _, info := range infoMap {
		infos = append(infos, info)
	}
	return infos
}

// CreateOptionsMap 创建选项映射表
func CreateOptionsMap[T comparable](infoMap map[T]EnumInfo[T]) map[string]string {
	options := make(map[string]string)
	for _, info := range infoMap {
		options[fmt.Sprintf("%v", info.Value)] = info.Label
	}
	return options
}

// CreateOptionsWithColorMap 创建包含颜色的选项映射表
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
