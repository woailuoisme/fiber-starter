package controllers

import (
	"strconv"

	"fiber-starter/app/Http/Middleware"
	"fiber-starter/app/Http/Resources"
	services "fiber-starter/app/Http/Services"
	models "fiber-starter/app/Models"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
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

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Name   string `json:"name" validate:"omitempty,min=2,max=100" example:"Alice" swagger:"required,user_name"`
	Phone  string `json:"phone" validate:"omitempty,e164" example:"+8613800138000" swagger:"optional,phone_number"`
	Avatar string `json:"avatar" validate:"omitempty,url" example:"https://example.com/avatar.jpg" swagger:"optional,avatar_url"` //nolint:lll
}

type requestValidationError struct {
	message string
	details interface{}
}

func (e *requestValidationError) Error() string {
	return e.message
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
// @Success 200 {object} resources.APIResponse "成功"
// @Failure 401 {object} resources.APIResponse "未授权"
// @Failure 500 {object} resources.APIResponse "服务器错误"
// @Router /api/users [get]
func (c *UserController) GetUsers(ctx fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, total, err := c.userService.GetUsers(page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(resources.ErrorResponse(err.Error(), nil))
	}

	userResponses := make([]models.SafeUser, len(users))
	for i, user := range users {
		userResponses[i] = user.ToSafeUser()
	}

	return ctx.Status(fiber.StatusOK).JSON(resources.SuccessResponse("Users fetched successfully", fiber.Map{
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
// @Description 根据 ID 获取用户详情（仅管理员可用）。
// @Tags 用户
// @Accept json
// @Produce json
// @Param id path int true "用户 ID"
// @Security ApiKeyAuth
// @Success 200 {object} resources.APIResponse "成功"
// @Failure 400 {object} resources.APIResponse "用户 ID 无效"
// @Failure 401 {object} resources.APIResponse "未授权"
// @Failure 404 {object} resources.APIResponse "用户不存在"
// @Router /api/users/{id} [get]
func (c *UserController) GetUser(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse("Invalid user ID", nil))
	}

	user, err := c.userService.GetUserByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(resources.ErrorResponse(err.Error(), nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(resources.SuccessResponse("User fetched successfully", user.ToSafeUser()))
}

// UpdateUser 更新用户信息
// @Summary 更新用户信息
// @Description 根据 ID 更新用户信息（仅管理员可用）。
// @Tags 用户
// @Accept json
// @Produce json
// @Param id path int true "用户 ID"
// @Param user body UpdateProfileRequest true "更新用户参数"
// @Security ApiKeyAuth
// @Success 200 {object} resources.APIResponse "成功"
// @Failure 400 {object} resources.APIResponse "请求错误"
// @Failure 401 {object} resources.APIResponse "未授权"
// @Failure 404 {object} resources.APIResponse "用户不存在"
// @Router /api/users/{id} [put]
func (c *UserController) UpdateUser(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse("Invalid user ID", nil))
	}

	req, reqErr := c.bindAndValidateUpdateProfileRequest(ctx)
	if reqErr != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse(reqErr.message, reqErr.details))
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
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse(err.Error(), nil))
	}

	user, err := c.userService.GetUserByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(resources.ErrorResponse("Failed to fetch updated user", nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(resources.SuccessResponse("User updated successfully", fiber.Map{
		"user": user.ToSafeUser(),
	}))
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 根据 ID 删除用户（仅管理员可用）。
// @Tags 用户
// @Accept json
// @Produce json
// @Param id path int true "用户 ID"
// @Security ApiKeyAuth
// @Success 200 {object} resources.APIResponse "成功"
// @Failure 400 {object} resources.APIResponse "用户 ID 无效"
// @Failure 401 {object} resources.APIResponse "未授权"
// @Failure 404 {object} resources.APIResponse "用户不存在"
// @Router /api/users/{id} [delete]
func (c *UserController) DeleteUser(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse("Invalid user ID", nil))
	}

	if err := c.userService.DeleteUser(id); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse(err.Error(), nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(resources.SuccessResponse("User deleted successfully", nil))
}

// UpdateProfile 更新个人资料
// @Summary 更新个人资料
// @Description 更新当前登录用户的个人资料。
// @Tags 用户
// @Accept json
// @Produce json
// @Param user body UpdateProfileRequest true "更新个人资料参数"
// @Security ApiKeyAuth
// @Success 200 {object} resources.APIResponse "成功"
// @Failure 400 {object} resources.APIResponse "请求错误"
// @Failure 401 {object} resources.APIResponse "未授权"
// @Router /api/v1/users/profile [put]
func (c *UserController) UpdateProfile(ctx fiber.Ctx) error {
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(resources.ErrorResponse("Unauthenticated user", nil))
	}

	req, reqErr := c.bindAndValidateUpdateProfileRequest(ctx)
	if reqErr != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse(reqErr.message, reqErr.details))
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
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse(err.Error(), nil))
	}

	updatedUser, err := c.userService.GetUserByID(user.ID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(resources.ErrorResponse("Failed to fetch updated user", nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(resources.SuccessResponse("Profile updated successfully", fiber.Map{
		"user": updatedUser.ToSafeUser(),
	}))
}

func (c *UserController) bindAndValidateUpdateProfileRequest(ctx fiber.Ctx) (UpdateProfileRequest, *requestValidationError) {
	var req UpdateProfileRequest

	if err := ctx.Bind().Body(&req); err != nil {
		return req, &requestValidationError{
			message: "Failed to parse request body",
			details: err.Error(),
		}
	}

	if err := c.validator.Struct(&req); err != nil {
		return req, &requestValidationError{
			message: "Request validation failed",
			details: resources.FormatValidationErrors(err),
		}
	}

	return req, nil
}

// GetCurrentUser 获取当前登录用户的信息
// @Summary 获取当前登录用户信息
// @Description 获取当前认证用户的详细信息。
// @Tags 资料
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} resources.APIResponse "成功"
// @Failure 401 {object} resources.APIResponse "未授权"
// @Failure 404 {object} resources.APIResponse "用户不存在"
// @Router /api/me [get]
func (c *UserController) GetCurrentUser(ctx fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(int64)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(resources.ErrorResponse("Unauthorized", nil))
	}

	currentUser, err := c.userService.GetUserByID(userID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(resources.ErrorResponse("User not found", nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(resources.SuccessResponse("User fetched successfully", fiber.Map{
		"user": currentUser.ToSafeUser(),
	}))
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
// @Success 200 {object} resources.APIResponse "成功"
// @Failure 400 {object} resources.APIResponse "请求错误"
// @Failure 401 {object} resources.APIResponse "未授权"
// @Failure 500 {object} resources.APIResponse "服务器错误"
// @Router /api/users/search [get]
func (c *UserController) SearchUsers(ctx fiber.Ctx) error {
	query := ctx.Query("q")
	if query == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse("Search keyword is required", nil))
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	users, total, err := c.userService.SearchUsers(query, page, limit)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(resources.ErrorResponse(err.Error(), nil))
	}

	userResponses := make([]models.SafeUser, len(users))
	for i, user := range users {
		userResponses[i] = user.ToSafeUser()
	}

	return ctx.Status(fiber.StatusOK).JSON(resources.SuccessResponse("Users searched successfully", fiber.Map{
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
