package controllers

import (
	"time"

	"fiber-starter/internal/platform/helpers"
	"fiber-starter/internal/services"

	"github.com/gofiber/fiber/v3"
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
// @Summary Set key
// @Description Store a key-value pair.
// @Tags storage
// @Accept json
// @Produce json
// @Param request body SetKeyRequest true "Set key payload"
// @Success 200 {object} resources.APIResponse
// @Failure 400 {object} resources.APIResponse
// @Router /api/storage/set [post]
func (c *StorageController) SetKey(ctx fiber.Ctx) error {
	var req SetKeyRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return helpers.HandleBadRequest(ctx, "Invalid request body")
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

	return helpers.HandleSuccess(ctx, "Value stored successfully", nil)
}

// GetKey 获取存储值
// @Summary Get key
// @Description Get a value by key.
// @Tags storage
// @Accept json
// @Produce json
// @Param key path string true "Storage key"
// @Success 200 {object} resources.APIResponse
// @Failure 400 {object} resources.APIResponse
// @Failure 404 {object} resources.APIResponse
// @Router /api/storage/get/{key} [get]
func (c *StorageController) GetKey(ctx fiber.Ctx) error {
	key := ctx.Params("key")
	if key == "" {
		return helpers.HandleBadRequest(ctx, "Storage key is required")
	}

	value, err := c.storageService.Get(key)
	if err != nil {
		return helpers.HandleNotFound(ctx, "Value not found")
	}

	response := GetKeyResponse{
		Key:   key,
		Value: string(value),
	}

	return helpers.HandleSuccess(ctx, "Value fetched successfully", response)
}

// DeleteKey 删除存储键
// @Summary Delete key
// @Description Delete a value by key.
// @Tags storage
// @Accept json
// @Produce json
// @Param key path string true "Storage key"
// @Success 200 {object} resources.APIResponse
// @Failure 400 {object} resources.APIResponse
// @Router /api/storage/delete/{key} [delete]
func (c *StorageController) DeleteKey(ctx fiber.Ctx) error {
	key := ctx.Params("key")
	if key == "" {
		return helpers.HandleBadRequest(ctx, "Storage key is required")
	}

	err := c.storageService.Delete(key)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Value deleted successfully", nil)
}

// Exists 检查键是否存在
// @Summary Check key existence
// @Description Check whether a key exists.
// @Tags storage
// @Accept json
// @Produce json
// @Param key path string true "Storage key"
// @Success 200 {object} resources.APIResponse
// @Failure 400 {object} resources.APIResponse
// @Router /api/storage/exists/{key} [get]
func (c *StorageController) Exists(ctx fiber.Ctx) error {
	key := ctx.Params("key")
	if key == "" {
		return helpers.HandleBadRequest(ctx, "Storage key is required")
	}

	exists, err := c.storageService.Exists(key)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	response := ExistsResponse{
		Key:    key,
		Exists: exists,
	}

	return helpers.HandleSuccess(ctx, "Existence checked successfully", response)
}

// SetExpire 设置键的过期时间
// @Summary Set key expiration
// @Description Set expiration time for an existing key.
// @Tags storage
// @Accept json
// @Produce json
// @Param request body SetExpireRequest true "Set expiration payload"
// @Success 200 {object} resources.APIResponse
// @Failure 400 {object} resources.APIResponse
// @Router /api/storage/expire [post]
func (c *StorageController) SetExpire(ctx fiber.Ctx) error {
	var req SetExpireRequest
	if err := ctx.Bind().Body(&req); err != nil {
		return helpers.HandleBadRequest(ctx, "Invalid request body")
	}

	// 验证请求参数
	if err := req.Validate(); err != nil {
		return helpers.HandleValidationError(ctx, err)
	}

	err := c.storageService.SetExpire(req.Key, time.Duration(req.TTL)*time.Second)
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Expiration updated successfully", nil)
}

// Reset 重置存储
// @Summary Reset storage
// @Description Delete all stored data.
// @Tags storage
// @Accept json
// @Produce json
// @Success 200 {object} resources.APIResponse
// @Failure 500 {object} resources.APIResponse
// @Router /api/storage/reset [post]
func (c *StorageController) Reset(ctx fiber.Ctx) error {
	err := c.storageService.Reset()
	if err != nil {
		return helpers.HandleError(ctx, err)
	}

	return helpers.HandleSuccess(ctx, "Storage reset successfully", nil)
}

// 请求和响应结构体

// SetKeyRequest 设置键值对请求
type SetKeyRequest struct {
	Key   string `json:"key" validate:"required"`
	Value string `json:"value" validate:"required"`
	TTL   int    `json:"ttl"` // Expiration in seconds. 0 uses the default TTL.
}

// Validate 验证设置键值对请求
func (r *SetKeyRequest) Validate() error {
	if r.Key == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Key is required")
	}
	if r.Value == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Value is required")
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
	TTL int    `json:"ttl" validate:"required,gt=0"` // Expiration in seconds.
}

// Validate 验证设置过期时间请求
func (r *SetExpireRequest) Validate() error {
	if r.Key == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Key is required")
	}
	if r.TTL <= 0 {
		return fiber.NewError(fiber.StatusBadRequest, "TTL must be greater than 0")
	}
	return nil
}
