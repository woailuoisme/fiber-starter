package Exports

import (
	models "fiber-starter/app/Models"
)

// UserExport 用户导出类
type UserExport struct {
	Users []models.User
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
	user := item.(models.User)

	phone := ""
	if user.Phone != nil {
		phone = *user.Phone
	}

	return []any{
		user.ID,
		user.Name,
		user.Email,
		phone,
		user.Status,
		user.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
