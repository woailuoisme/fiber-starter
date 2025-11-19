package controllers

import (
	"time"

	"fiber-starter/app/helpers"
	"fiber-starter/app/services"

	"github.com/gofiber/fiber/v2"
)

// StorageController 存储控制器
type StorageController struct {
	storageService *services.StorageService
}

// NewStorageController 创建新的存储控制器
func NewStorageController(storageService *services.StorageService) *StorageController {
	return &StorageController{
		storageService: storageService,
	}
}

// SetKey 设置存储键值对
// @Summary 设置存储键值对
// @Description 设置一个键值对到存储中
// @Tags storage
// @Accept json
// @Produce json
// @Param request body SetKeyRequest true "设置键值对请求"
// @Success 200 {object} helpers.Response
// @Failure 400 {object} helpers.Response
// @Router /api/storage/set [post]
func (c *StorageController) SetKey(ctx *fiber.Ctx) error {
	var req SetKeyRequest
	if err := ctx.BodyParser(&req); err != nil {
		return helpers.HandleBadRequest(ctx, "无效的请求参数")
	}

	// 验证请求参数
	if err := req.Validate(); err != nil {
		return helpers.HandleValidationError(ctx, err)
	}

	// 设置存储值
	var err error
	if req.TTL > 0 {
		err = c.storageService.Set(req.Key, []byte(req.Value), time.Duration(req.TTL)*time.Second)
	} else {
		err = c.storageService.SetWithDefaultTTL(req.Key, []byte(req.Value))
	}

	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "存储值设置成功", nil)
}

// GetKey 获取存储值
// @Summary 获取存储值
// @Description 根据键获取存储中的值
// @Tags storage
// @Accept json
// @Produce json
// @Param key path string true "存储键"
// @Success 200 {object} helpers.Response
// @Failure 400 {object} helpers.Response
// @Failure 404 {object} helpers.Response
// @Router /api/storage/get/{key} [get]
func (c *StorageController) GetKey(ctx *fiber.Ctx) error {
	key := ctx.Params("key")
	if key == "" {
		return helpers.HandleBadRequest(ctx, "存储键不能为空")
	}

	value, err := c.storageService.Get(key)
	if err != nil {
		return helpers.HandleNotFound(ctx, "存储值不存在")
	}

	response := GetKeyResponse{
		Key:   key,
		Value: string(value),
	}

	return helpers.HandleSuccess(ctx, "获取存储值成功", response)
}

// DeleteKey 删除存储键
// @Summary 删除存储键
// @Description 根据键删除存储中的值
// @Tags storage
// @Accept json
// @Produce json
// @Param key path string true "存储键"
// @Success 200 {object} helpers.Response
// @Failure 400 {object} helpers.Response
// @Router /api/storage/delete/{key} [delete]
func (c *StorageController) DeleteKey(ctx *fiber.Ctx) error {
	key := ctx.Params("key")
	if key == "" {
		return helpers.HandleBadRequest(ctx, "存储键不能为空")
	}

	err := c.storageService.Delete(key)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "删除存储值成功", nil)
}

// Exists 检查键是否存在
// @Summary 检查键是否存在
// @Description 检查存储中是否存在指定的键
// @Tags storage
// @Accept json
// @Produce json
// @Param key path string true "存储键"
// @Success 200 {object} helpers.Response
// @Failure 400 {object} helpers.Response
// @Router /api/storage/exists/{key} [get]
func (c *StorageController) Exists(ctx *fiber.Ctx) error {
	key := ctx.Params("key")
	if key == "" {
		return helpers.HandleBadRequest(ctx, "存储键不能为空")
	}

	exists, err := c.storageService.Exists(key)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	response := ExistsResponse{
		Key:    key,
		Exists: exists,
	}

	return helpers.HandleSuccess(ctx, "检查键存在性成功", response)
}

// SetExpire 设置键的过期时间
// @Summary 设置键的过期时间
// @Description 为已存在的键设置过期时间
// @Tags storage
// @Accept json
// @Produce json
// @Param request body SetExpireRequest true "设置过期时间请求"
// @Success 200 {object} helpers.Response
// @Failure 400 {object} helpers.Response
// @Router /api/storage/expire [post]
func (c *StorageController) SetExpire(ctx *fiber.Ctx) error {
	var req SetExpireRequest
	if err := ctx.BodyParser(&req); err != nil {
		return helpers.HandleBadRequest(ctx, "无效的请求参数")
	}

	// 验证请求参数
	if err := req.Validate(); err != nil {
		return helpers.HandleValidationError(ctx, err)
	}

	err := c.storageService.SetExpire(req.Key, time.Duration(req.TTL)*time.Second)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "设置过期时间成功", nil)
}

// Reset 重置存储
// @Summary 重置存储
// @Description 清空存储中的所有数据
// @Tags storage
// @Accept json
// @Produce json
// @Success 200 {object} helpers.Response
// @Failure 500 {object} helpers.Response
// @Router /api/storage/reset [post]
func (c *StorageController) Reset(ctx *fiber.Ctx) error {
	err := c.storageService.Reset()
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "重置存储成功", nil)
}

// 请求和响应结构体

// SetKeyRequest 设置键值对请求
type SetKeyRequest struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
	TTL   int    `json:"ttl"` // 过期时间（秒），0表示使用默认TTL
}

// Validate 验证设置键值对请求
func (r *SetKeyRequest) Validate() error {
	if r.Key == "" {
		return fiber.NewError(fiber.StatusBadRequest, "键不能为空")
	}
	if r.Value == "" {
		return fiber.NewError(fiber.StatusBadRequest, "值不能为空")
	}
	return nil
}

// GetKeyResponse 获取键值对响应
type GetKeyResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ExistsResponse 检查键存在性响应
type ExistsResponse struct {
	Key    string `json:"key"`
	Exists bool   `json:"exists"`
}

// SetExpireRequest 设置过期时间请求
type SetExpireRequest struct {
	Key string `json:"key" validate:"required"`
	TTL int    `json:"ttl" validate:"required,gt=0"` // 过期时间（秒）
}

// Validate 验证设置过期时间请求
func (r *SetExpireRequest) Validate() error {
	if r.Key == "" {
		return fiber.NewError(fiber.StatusBadRequest, "键不能为空")
	}
	if r.TTL <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "TTL必须大于0")
	}
	return nil
}
