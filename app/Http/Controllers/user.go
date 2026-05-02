package controllers

import (
	"strconv"

	"fiber-starter/app/Http/Middleware"
	requests "fiber-starter/app/Http/Requests"
	services "fiber-starter/app/Http/Services"
	models "fiber-starter/app/Models"
	helpers "fiber-starter/app/Support"

	"github.com/gofiber/fiber/v3"
)

// UserController 用户控制器
type UserController struct {
	userService services.UserService
}

// NewUserController 创建用户控制器实例
func NewUserController(userService services.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// GetUsers 获取用户列表
// @Summary 获取用户列表
// @Description 获取分页用户列表（仅管理员可用）。
// @Tags 用户
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Security ApiKeyAuth
// @Success 200 {object} helpers.APIResponse "成功"
// @Failure 401 {object} helpers.APIResponse "未授权"
// @Failure 500 {object} helpers.APIResponse "服务器错误"
// @Router /api/users [get]
func (c *UserController) GetUsers(ctx fiber.Ctx) error {
	var req requests.UserListRequest
	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}
	page, limit := req.Pagination()

	users, total, err := c.userService.GetUsers(page, limit)
	if err != nil {
		return helpers.HandleInternalServerError(ctx, err.Error())
	}

	userResponses := make([]models.SafeUser, len(users))
	for i, user := range users {
		userResponses[i] = user.ToSafeUser()
	}

	return helpers.HandlePaginationResponse(ctx, "Users fetched successfully", userResponses, total, page, limit)
}

// GetUser 获取单个用户
// @Summary 获取单个用户
// @Description 根据 ID 获取用户详情（仅管理员可用）。
// @Tags 用户
// @Accept json
// @Produce json
// @Param id path int true "用户 ID"
// @Security ApiKeyAuth
// @Success 200 {object} helpers.APIResponse "成功"
// @Failure 400 {object} helpers.APIResponse "用户 ID 无效"
// @Failure 401 {object} helpers.APIResponse "未授权"
// @Failure 404 {object} helpers.APIResponse "用户不存在"
// @Router /api/users/{id} [get]
func (c *UserController) GetUser(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return helpers.HandleBadRequest(ctx, "Invalid user ID")
	}

	user, err := c.userService.GetUserByID(id)
	if err != nil {
		return helpers.HandleNotFound(ctx, err.Error())
	}

	return helpers.HandleUserResponse(ctx, "User fetched successfully", user)
}

// UpdateUser 更新用户信息
// @Summary 更新用户信息
// @Description 根据 ID 更新用户信息（仅管理员可用）。
// @Tags 用户
// @Accept json
// @Produce json
// @Param id path int true "用户 ID"
// @Param user body requests.UpdateProfileRequest true "更新用户参数"
// @Security ApiKeyAuth
// @Success 200 {object} helpers.APIResponse "成功"
// @Failure 400 {object} helpers.APIResponse "请求错误"
// @Failure 401 {object} helpers.APIResponse "未授权"
// @Failure 404 {object} helpers.APIResponse "用户不存在"
// @Router /api/users/{id} [put]
func (c *UserController) UpdateUser(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return helpers.HandleBadRequest(ctx, "Invalid user ID")
	}

	var req requests.UpdateProfileRequest
	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}

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

	if err := c.userService.UpdateUser(id, updates); err != nil {
		return helpers.HandleBadRequest(ctx, err.Error())
	}

	user, err := c.userService.GetUserByID(id)
	if err != nil {
		return helpers.HandleInternalServerError(ctx, "Failed to fetch updated user")
	}

	return helpers.HandleUserResponse(ctx, "User updated successfully", user)
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 根据 ID 删除用户（仅管理员可用）。
// @Tags 用户
// @Accept json
// @Produce json
// @Param id path int true "用户 ID"
// @Security ApiKeyAuth
// @Success 200 {object} helpers.APIResponse "成功"
// @Failure 400 {object} helpers.APIResponse "用户 ID 无效"
// @Failure 401 {object} helpers.APIResponse "未授权"
// @Failure 404 {object} helpers.APIResponse "用户不存在"
// @Router /api/users/{id} [delete]
func (c *UserController) DeleteUser(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return helpers.HandleBadRequest(ctx, "Invalid user ID")
	}

	if err := c.userService.DeleteUser(id); err != nil {
		return helpers.HandleBadRequest(ctx, err.Error())
	}

	return helpers.HandleSuccess(ctx, "User deleted successfully", nil)
}

// UpdateProfile 更新个人资料
// @Summary 更新个人资料
// @Description 更新当前登录用户的个人资料。
// @Tags 用户
// @Accept json
// @Produce json
// @Param user body requests.UpdateProfileRequest true "更新个人资料参数"
// @Security ApiKeyAuth
// @Success 200 {object} helpers.APIResponse "成功"
// @Failure 400 {object} helpers.APIResponse "请求错误"
// @Failure 401 {object} helpers.APIResponse "未授权"
// @Router /api/v1/users/profile [put]
func (c *UserController) UpdateProfile(ctx fiber.Ctx) error {
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return helpers.HandleUnauthorized(ctx, "Unauthenticated user")
	}

	var req requests.UpdateProfileRequest
	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	profile := &models.User{
		Name: req.Name,
	}
	if req.Phone != "" {
		profile.Phone = &req.Phone
	}
	if req.Avatar != "" {
		profile.Avatar = &req.Avatar
	}

	if err := c.userService.UpdateProfile(user.ID, profile); err != nil {
		return helpers.HandleBadRequest(ctx, err.Error())
	}

	updatedUser, err := c.userService.GetUserByID(user.ID)
	if err != nil {
		return helpers.HandleInternalServerError(ctx, "Failed to fetch updated user")
	}

	return helpers.HandleUserResponse(ctx, "Profile updated successfully", updatedUser)
}

// GetCurrentUser 获取当前登录用户的信息
// @Summary 获取当前登录用户信息
// @Description 获取当前认证用户的详细信息。
// @Tags 资料
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} helpers.APIResponse "成功"
// @Failure 401 {object} helpers.APIResponse "未授权"
// @Failure 404 {object} helpers.APIResponse "用户不存在"
// @Router /api/me [get]
func (c *UserController) GetCurrentUser(ctx fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(int64)
	if !ok {
		return helpers.HandleUnauthorized(ctx, "Unauthorized")
	}

	currentUser, err := c.userService.GetUserByID(userID)
	if err != nil {
		return helpers.HandleNotFound(ctx, "User not found")
	}

	return helpers.HandleUserResponse(ctx, "User fetched successfully", currentUser)
}

// SearchUsers 搜索用户
// @Summary 搜索用户
// @Description 根据关键词搜索用户（仅管理员可用）。
// @Tags 用户
// @Accept json
// @Produce json
// @Param q query string true "搜索关键词"
// @Param page query int false "页码" default(1)
// @Param limit query int false "每页数量" default(10)
// @Security ApiKeyAuth
// @Success 200 {object} helpers.APIResponse "成功"
// @Failure 400 {object} helpers.APIResponse "请求错误"
// @Failure 401 {object} helpers.APIResponse "未授权"
// @Failure 500 {object} helpers.APIResponse "服务器错误"
// @Router /api/users/search [get]
func (c *UserController) SearchUsers(ctx fiber.Ctx) error {
	var req requests.SearchUsersRequest
	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}
	page, limit := req.Pagination()

	users, total, err := c.userService.SearchUsers(req.Q, page, limit)
	if err != nil {
		return helpers.HandleInternalServerError(ctx, err.Error())
	}

	userResponses := make([]models.SafeUser, len(users))
	for i, user := range users {
		userResponses[i] = user.ToSafeUser()
	}

	return helpers.HandlePaginationResponse(ctx, "Users searched successfully", userResponses, total, page, limit)
}
