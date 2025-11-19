package main

import (
	"log"
	"os"

	"fiber-starter/config"
	"fiber-starter/database"
)

func main() {
	// 设置日志格式，包含时间戳和文件位置
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Printf("开始测试数据库连接日志功能")

	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Printf("加载配置失败: %v", err)
		return
	}

	log.Printf("配置加载成功")

	// 测试数据库连接
	_, err = database.NewConnection(cfg)
	if err != nil {
		log.Printf("数据库连接测试失败: %v", err)
		return
	}

	log.Printf("数据库连接测试完成")

	// 测试自动迁移
	err = database.AutoMigrate()
	if err != nil {
		log.Printf("数据库自动迁移测试失败: %v", err)
		return
	}

	log.Printf("数据库自动迁移测试完成")

	// 获取数据库连接并关闭
	db := database.GetDB()
	if db != nil {
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("获取底层sql.DB失败: %v", err)
			return
		}

		log.Printf("准备关闭数据库连接")
		if err := sqlDB.Close(); err != nil {
			log.Printf("关闭数据库连接失败: %v", err)
		} else {
			log.Printf("数据库连接已成功关闭")
		}
	}

	log.Printf("数据库连接日志功能测试结束")
}
