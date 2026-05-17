package Exports

import (
	"fiber-starter/internal/features/user"
)

// UserExport 用户导出类
type UserExport struct {
	Users []user.User
}

// Collection 获取导出数据
func (e *UserExport) Collection() any {
	return e.Users
}

// Headings 设置表头
func (e *UserExport) Headings() []string {
	return []string{
		"ID",
		"Name",
		"Email",
		"Phone",
		"Status",
		"Created At",
	}
}

// Map 映射每行数据
func (e *UserExport) Map(item any) []any {
	u := item.(user.User)

	phone := ""
	if u.Phone != nil {
		phone = *u.Phone
	}

	return []any{
		u.ID,
		u.Name,
		u.Email,
		phone,
		u.Status,
		u.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
