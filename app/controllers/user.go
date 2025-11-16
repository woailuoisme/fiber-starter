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
	validator   *validator.Validate
}

// NewUserController 创建用户控制器实例
func NewUserController(userService services.UserService, validate *validator.Validate) *UserController {
	return &UserController{
		userService: userService,
		validator:   validate,
	}
}

// UpdateProfileRequest 更新资料请求结构
type UpdateProfileRequest struct {
	Name   string `json:"name" validate:"omitempty,min=2,max=100" example:"张三" swagger:"required,用户姓名"`
	Phone  string `json:"phone" validate:"omitempty,e164" example:"+8613800138000" swagger:"optional,手机号码"`
	Avatar string `json:"avatar" validate:"omitempty,url" example:"https://example.com/avatar.jpg" swagger:"optional,头像URL"`
}

// GetUsers 获取用户列表
// @Summary 获取用户列表
// @Description 分页获取用户列表，需要管理员权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Security ApiKeyAuth
// @Success 200 {object} helpers.APIResponse "获取成功"
// @Failure 401 {object} helpers.APIResponse "未授权"
// @Failure 500 {object} helpers.APIResponse "服务器错误"
// @Router /api/users [get]
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
	userResponses := make([]models.SafeUser, len(users))
	for i, user := range users {
		userResponses[i] = user.ToSafeUser()
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
// @Summary 获取单个用户
// @Description 根据用户ID获取用户详细信息，需要管理员权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Security ApiKeyAuth
// @Success 200 {object} helpers.APIResponse "获取成功"
// @Failure 400 {object} helpers.APIResponse "无效的用户ID"
// @Failure 401 {object} helpers.APIResponse "未授权"
// @Failure 404 {object} helpers.APIResponse "用户不存在"
// @Router /api/users/{id} [get]
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
// @Summary 更新用户信息
// @Description 更新指定用户的信息，需要管理员权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param user body UpdateProfileRequest true "更新用户信息请求"
// @Security ApiKeyAuth
// @Success 200 {object} helpers.APIResponse "更新成功"
// @Failure 400 {object} helpers.APIResponse "请求参数错误"
// @Failure 401 {object} helpers.APIResponse "未授权"
// @Failure 404 {object} helpers.APIResponse "用户不存在"
// @Router /api/users/{id} [put]
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
	if err := c.validator.Struct(&req); err != nil {
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
// @Summary 删除用户
// @Description 删除指定用户，需要管理员权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Security ApiKeyAuth
// @Success 200 {object} helpers.APIResponse "删除成功"
// @Failure 400 {object} helpers.APIResponse "无效的用户ID"
// @Failure 401 {object} helpers.APIResponse "未授权"
// @Failure 404 {object} helpers.APIResponse "用户不存在"
// @Router /api/users/{id} [delete]
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

// UpdateProfile 更新个人资料
// @Summary 更新个人资料
// @Description 更新当前登录用户的个人信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param user body UpdateProfileRequest true "更新个人资料请求"
// @Security ApiKeyAuth
// @Success 200 {object} helpers.APIResponse "更新成功"
// @Failure 400 {object} helpers.APIResponse "请求参数错误"
// @Failure 401 {object} helpers.APIResponse "未授权"
// @Router /api/v1/users/profile [put]
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
	if err := c.validator.Struct(&req); err != nil {
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

// GetCurrentUser 获取当前登录用户的信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户资料
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} helpers.APIResponse "获取成功"
// @Failure 401 {object} helpers.APIResponse "未授权"
// @Failure 404 {object} helpers.APIResponse "用户不存在"
// @Router /api/me [get]
func (c *UserController) GetCurrentUser(ctx *fiber.Ctx) error {
	// 从上下文中获取用户ID
	userID, ok := ctx.Locals("user_id").(uint)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(helpers.ErrorResponse("未授权", nil))
	}

	// 获取完整的用户信息
	currentUser, err := c.userService.GetUserByID(userID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(helpers.ErrorResponse("用户不存在", nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(helpers.SuccessResponse("获取用户信息成功", fiber.Map{
		"user": currentUser.ToSafeUser(),
	}))
}

// SearchUsers 搜索用户
// @Summary 搜索用户
// @Description 根据关键词搜索用户，需要管理员权限
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param q query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Security ApiKeyAuth
// @Success 200 {object} helpers.APIResponse "搜索成功"
// @Failure 400 {object} helpers.APIResponse "请求参数错误"
// @Failure 401 {object} helpers.APIResponse "未授权"
// @Failure 500 {object} helpers.APIResponse "服务器错误"
// @Router /api/users/search [get]
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
	userResponses := make([]models.SafeUser, len(users))
	for i, user := range users {
		userResponses[i] = user.ToSafeUser()
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