package migrations

import (
	"fiber-starter/app/models"
	"fiber-starter/database"

	"gorm.io/gorm"
)

// MigrateUsers 迁移用户表
func MigrateUsers(db *gorm.DB) error {
	// 自动迁移User模型
	return db.AutoMigrate(&models.User{})
}

// RollbackUsers 回滚用户表
func RollbackUsers(db *gorm.DB) error {
	// 删除users表
	return db.Migrator().DropTable(&models.User{})
}

// CreateUserTable 创建用户表的migration
func CreateUserTable() error {
	db := database.GetDB()
	if db == nil {
		return gorm.ErrInvalidTransaction
	}

	// 执行迁移
	if err := MigrateUsers(db); err != nil {
		return err
	}

	return nil
}

// DropUserTable 删除用户表的migration
func DropUserTable() error {
	db := database.GetDB()
	if db == nil {
		return gorm.ErrInvalidTransaction
	}

	// 执行回滚
	if err := RollbackUsers(db); err != nil {
		return err
	}

	return nil
}
