package routes

import (
	"fiber-starter/app/controllers"

	"github.com/gofiber/fiber/v2"
)

// SetupStorageRoutes 设置存储相关路由
func SetupStorageRoutes(api fiber.Router, storageController *controllers.StorageController) {
	storage := api.Group("/storage")
	{
		// 设置键值对
		storage.Post("/set", storageController.SetKey)
		
		// 获取键值
		storage.Get("/get/:key", storageController.GetKey)
		
		// 删除键
		storage.Delete("/delete/:key", storageController.DeleteKey)
		
		// 检查键是否存在
		storage.Get("/exists/:key", storageController.Exists)
		
		// 设置过期时间
		storage.Post("/expire", storageController.SetExpire)
		
		// 重置存储
		storage.Post("/reset", storageController.Reset)
	}
}