package migrations

import (
	"fiber-starter/database"
	"log"

	"gorm.io/gorm"
)

// RunAllMigrations 运行所有数据库迁移
func RunAllMigrations() error {
	db := database.GetDB()
	if db == nil {
		return gorm.ErrInvalidTransaction
	}

	log.Println("开始执行数据库迁移...")

	// 迁移用户表
	if err := MigrateUsers(db); err != nil {
		log.Printf("用户表迁移失败: %v", err)
		return err
	}
	log.Println("用户表迁移完成")

	// 在这里添加其他表的迁移
	// if err := MigratePosts(db); err != nil {
	//     log.Printf("文章表迁移失败: %v", err)
	//     return err
	// }
	// log.Println("文章表迁移完成")

	log.Println("所有数据库迁移完成")
	return nil
}

// RollbackAllMigrations 回滚所有数据库迁移
func RollbackAllMigrations() error {
	db := database.GetDB()
	if db == nil {
		return gorm.ErrInvalidTransaction
	}

	log.Println("开始回滚数据库迁移...")

	// 回滚用户表
	if err := RollbackUsers(db); err != nil {
		log.Printf("用户表回滚失败: %v", err)
		return err
	}
	log.Println("用户表回滚完成")

	// 在这里添加其他表的回滚
	// if err := RollbackPosts(db); err != nil {
	//     log.Printf("文章表回滚失败: %v", err)
	//     return err
	// }
	// log.Println("文章表回滚完成")

	log.Println("所有数据库迁移回滚完成")
	return nil
}

// ResetDatabase 重置数据库（先回滚再迁移）
func ResetDatabase() error {
	log.Println("开始重置数据库...")

	// 先回滚所有迁移
	if err := RollbackAllMigrations(); err != nil {
		log.Printf("数据库回滚失败: %v", err)
		return err
	}

	// 再执行所有迁移
	if err := RunAllMigrations(); err != nil {
		log.Printf("数据库迁移失败: %v", err)
		return err
	}

	log.Println("数据库重置完成")
	return nil
}
