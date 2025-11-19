package seeders

import (
	"fiber-starter/database"
	"log"

	"gorm.io/gorm"
)

// RunAllSeeders 运行所有种子数据
func RunAllSeeders() error {
	db := database.GetDB()
	if db == nil {
		return gorm.ErrInvalidTransaction
	}

	log.Println("开始执行种子数据...")

	// 创建用户种子数据
	if err := SeedUsers(db); err != nil {
		log.Printf("用户种子数据创建失败: %v", err)
		return err
	}
	log.Println("用户种子数据创建完成")

	// 在这里添加其他表的种子数据
	// if err := SeedPosts(db); err != nil {
	//     log.Printf("文章种子数据创建失败: %v", err)
	//     return err
	// }
	// log.Println("文章种子数据创建完成")

	log.Println("所有种子数据创建完成")
	return nil
}

// RunRandomSeeders 运行随机种子数据生成
func RunRandomSeeders(userCount int) error {
	db := database.GetDB()
	if db == nil {
		return gorm.ErrInvalidTransaction
	}

	log.Printf("开始生成 %d 个随机用户数据...", userCount)

	// 生成随机用户数据
	if err := GenerateTestUsers(db, userCount); err != nil {
		log.Printf("随机用户数据生成失败: %v", err)
		return err
	}
	log.Printf("成功生成 %d 个随机用户数据", userCount)

	return nil
}

// ClearAllSeeders 清除所有种子数据
func ClearAllSeeders() error {
	db := database.GetDB()
	if db == nil {
		return gorm.ErrInvalidTransaction
	}

	log.Println("开始清除种子数据...")

	// 清除用户种子数据
	if err := ClearUsers(db); err != nil {
		log.Printf("用户种子数据清除失败: %v", err)
		return err
	}
	log.Println("用户种子数据清除完成")

	// 在这里添加其他表的种子数据清除
	// if err := ClearPosts(db); err != nil {
	//     log.Printf("文章种子数据清除失败: %v", err)
	//     return err
	// }
	// log.Println("文章种子数据清除完成")

	log.Println("所有种子数据清除完成")
	return nil
}

// RefreshAllSeeders 刷新所有种子数据（先清除再创建）
func RefreshAllSeeders() error {
	log.Println("开始刷新种子数据...")

	// 先清除所有种子数据
	if err := ClearAllSeeders(); err != nil {
		log.Printf("种子数据清除失败: %v", err)
		return err
	}

	// 再创建所有种子数据
	if err := RunAllSeeders(); err != nil {
		log.Printf("种子数据创建失败: %v", err)
		return err
	}

	log.Println("种子数据刷新完成")
	return nil
}

// SetupDatabase 设置数据库（迁移 + 种子数据）
func SetupDatabase() error {
	log.Println("开始设置数据库...")

	// 先执行迁移
	if err := database.AutoMigrate(); err != nil {
		log.Printf("数据库迁移失败: %v", err)
		return err
	}

	// 再执行种子数据
	if err := RunAllSeeders(); err != nil {
		log.Printf("种子数据创建失败: %v", err)
		return err
	}

	log.Println("数据库设置完成")
	return nil
}
