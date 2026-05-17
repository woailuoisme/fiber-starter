package Imports

import (
	models "fiber-starter/app/Models"
)

// UserImport 用户导入类
type UserImport struct{}

// Model 将行数据转换为模型
func (i *UserImport) Model(row []string) any {
	if len(row) < 3 {
		return nil
	}

	// 假设 Excel 列顺序为：姓名, 邮箱, 电话
	// 注意：这里跳过了 ID，因为通常导入是新增或通过邮箱匹配
	user := &models.User{
		Name:  row[0],
		Email: row[1],
	}

	if len(row) > 2 && row[2] != "" {
		phone := row[2]
		user.Phone = &phone
	}

	// 默认状态
	user.Status = models.UserStatusActive

	return user
}
