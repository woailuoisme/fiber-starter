package controllers

import (
	"fiber-starter/app/services"
	"fiber-starter/app/models"
	"fiber-starter/app/middleware"
	"fiber-starter/app/helpers"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/go-playground/validator/v10"
)

// UserController 用户控制器
type UserController struct {
	userService services.UserService
	validate    *validator.Validate
}

// NewUserController 创建用户控制器实例
func NewUserController(userService services.UserService) *UserController {
	return &UserController{
		userService: userService,
		validate:    validator.New(),
	}
}

// UpdateProfileRequest 更新资料请求结构
type UpdateProfileRequest struct {
	Name   string `json:"name" validate:"omitempty,min=2,max=100"`
	Phone  string `json:"phone" validate:"omitempty,e164"`
	Avatar string `json:"avatar" validate:"omitempty,url"`
}

// GetUsers 获取用户列表
func (c *UserController) GetUsers(ctx *fiber.Ctx) error {
	// 获取分页参数
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// 调用用户服务获取用户列表
	users, total, err := c.userService.GetUsers(page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(helpers.ErrorResponse(err.Error(), nil))
	}

	// 转换为安全响应
	userResponses := make([]map[string]interface{}, len(users))
	for i, user := range users {
		userResponses[i] = user.ToSafeResponse()
	}

	return ctx.Status(fiber.StatusOK).JSON(helpers.SuccessResponse("获取用户列表成功", fiber.Map{
		"users": userResponses,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	}))
}

// GetUser 获取单个用户
func (c *UserController) GetUser(ctx *fiber.Ctx) error {
	// 获取用户ID
	idStr := ctx.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("无效的用户ID", nil))
	}

	// 调用用户服务获取用户
	user, err := c.userService.GetUserByID(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(helpers.ErrorResponse(err.Error(), nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(helpers.SuccessResponse("获取用户信息成功", user.ToSafeResponse()))
}

// UpdateUser 更新用户信息
func (c *UserController) UpdateUser(ctx *fiber.Ctx) error {
	// 获取用户ID
	idStr := ctx.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("无效的用户ID", nil))
	}

	// 解析请求体
	var req UpdateProfileRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("请求参数解析失败", err.Error()))
	}

	// 验证请求参数
	if err := c.validate.Struct(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("请求参数验证失败", helpers.FormatValidationErrors(err)))
	}

	// 构建更新数据
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Phone != "" {
		updates["phone"] = &req.Phone
	}
	if req.Avatar != "" {
		updates["avatar"] = &req.Avatar
	}

	// 调用用户服务更新用户
	err = c.userService.UpdateUser(uint(id), updates)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse(err.Error(), nil))
	}

	// 获取更新后的用户信息
	user, err := c.userService.GetUserByID(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(helpers.ErrorResponse("获取更新后的用户信息失败", nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(helpers.SuccessResponse("用户信息更新成功", fiber.Map{
		"user": user.ToSafeResponse(),
	}))
}

// DeleteUser 删除用户
func (c *UserController) DeleteUser(ctx *fiber.Ctx) error {
	// 获取用户ID
	idStr := ctx.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("无效的用户ID", nil))
	}

	// 调用用户服务删除用户
	err = c.userService.DeleteUser(uint(id))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse(err.Error(), nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(helpers.SuccessResponse("用户删除成功", nil))
}

// UpdateProfile 更新当前用户资料
func (c *UserController) UpdateProfile(ctx *fiber.Ctx) error {
	// 从上下文中获取用户信息
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(helpers.ErrorResponse("未认证用户", nil))
	}

	// 解析请求体
	var req UpdateProfileRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("请求参数解析失败", err.Error()))
	}

	// 验证请求参数
	if err := c.validate.Struct(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("请求参数验证失败", helpers.FormatValidationErrors(err)))
	}

	// 构建更新资料
	profile := &models.User{
		Name: req.Name,
	}
	if req.Phone != "" {
		profile.Phone = &req.Phone
	}
	if req.Avatar != "" {
		profile.Avatar = &req.Avatar
	}

	// 调用用户服务更新资料
	err := c.userService.UpdateProfile(user.ID, profile)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse(err.Error(), nil))
	}

	// 获取更新后的用户信息
	updatedUser, err := c.userService.GetUserByID(user.ID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(helpers.ErrorResponse("获取更新后的用户信息失败", nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(helpers.SuccessResponse("资料更新成功", fiber.Map{
		"user": updatedUser.ToSafeResponse(),
	}))
}

// SearchUsers 搜索用户
func (c *UserController) SearchUsers(ctx *fiber.Ctx) error {
	// 获取搜索参数
	query := ctx.Query("q")
	if query == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(helpers.ErrorResponse("搜索关键词不能为空", nil))
	}

	// 获取分页参数
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))

	// 验证分页参数
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// 调用用户服务搜索用户
	users, total, err := c.userService.SearchUsers(query, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(helpers.ErrorResponse(err.Error(), nil))
	}

	// 转换为安全响应
	userResponses := make([]map[string]interface{}, len(users))
	for i, user := range users {
		userResponses[i] = user.ToSafeResponse()
	}

	return ctx.Status(fiber.StatusOK).JSON(helpers.SuccessResponse("搜索用户成功", fiber.Map{
		"users": userResponses,
		"query": query,
		"pagination": fiber.Map{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	}))
}