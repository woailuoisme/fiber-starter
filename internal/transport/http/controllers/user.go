package controllers

import (
	"strconv"

	models "fiber-starter/internal/domain/model"
	"fiber-starter/internal/services"
	"fiber-starter/internal/transport/http/middleware"
	"fiber-starter/internal/transport/http/resources"

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

// UpdateProfileRequest 更新资料请求结构
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
// @Summary List users
// @Description Get a paginated list of users (admin only).
// @Tags Users
// @Accept JSON
// @Produce JSON
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Security ApiKeyAuth
// @Success 200 {object} resources.APIResponse "OK"
// @Failure 401 {object} resources.APIResponse "Unauthorized"
// @Failure 500 {object} resources.APIResponse "Internal server error"
// @Router /api/users [get]
func (c *UserController) GetUsers(ctx fiber.Ctx) error {
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
		return ctx.Status(fiber.StatusInternalServerError).JSON(resources.ErrorResponse(err.Error(), nil))
	}

	// 转换为安全响应
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
// @Summary Get user
// @Description Get a user by ID (admin only).
// @Tags Users
// @Accept JSON
// @Produce JSON
// @Param id path int true "User ID"
// @Security ApiKeyAuth
// @Success 200 {object} resources.APIResponse "OK"
// @Failure 400 {object} resources.APIResponse "Invalid user ID"
// @Failure 401 {object} resources.APIResponse "Unauthorized"
// @Failure 404 {object} resources.APIResponse "User not found"
// @Router /api/users/{id} [get]
func (c *UserController) GetUser(ctx fiber.Ctx) error {
	// 获取用户ID
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse("Invalid user ID", nil))
	}

	// 调用用户服务获取用户
	user, err := c.userService.GetUserByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(resources.ErrorResponse(err.Error(), nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(resources.SuccessResponse("User fetched successfully", user.ToSafeUser()))
}

// UpdateUser 更新用户信息
// @Summary Update user
// @Description Update a user by ID (admin only).
// @Tags Users
// @Accept JSON
// @Produce JSON
// @Param id path int true "User ID"
// @Param user body UpdateProfileRequest true "Update user payload"
// @Security ApiKeyAuth
// @Success 200 {object} resources.APIResponse "OK"
// @Failure 400 {object} resources.APIResponse "Bad request"
// @Failure 401 {object} resources.APIResponse "Unauthorized"
// @Failure 404 {object} resources.APIResponse "User not found"
// @Router /api/users/{id} [put]
func (c *UserController) UpdateUser(ctx fiber.Ctx) error {
	// 获取用户ID
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse("Invalid user ID", nil))
	}

	req, reqErr := c.bindAndValidateUpdateProfileRequest(ctx)
	if reqErr != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse(reqErr.message, reqErr.details))
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
	err = c.userService.UpdateUser(id, updates)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse(err.Error(), nil))
	}

	// 获取更新后的用户信息
	user, err := c.userService.GetUserByID(id)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(resources.ErrorResponse("Failed to fetch updated user", nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(resources.SuccessResponse("User updated successfully", fiber.Map{
		"user": user.ToSafeUser(),
	}))
}

// DeleteUser 删除用户
// @Summary Delete user
// @Description Delete a user by ID (admin only).
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Security ApiKeyAuth
// @Success 200 {object} resources.APIResponse "OK"
// @Failure 400 {object} resources.APIResponse "Invalid user ID"
// @Failure 401 {object} resources.APIResponse "Unauthorized"
// @Failure 404 {object} resources.APIResponse "User not found"
// @Router /api/users/{id} [delete]
func (c *UserController) DeleteUser(ctx fiber.Ctx) error {
	// 获取用户ID
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse("Invalid user ID", nil))
	}

	// 调用用户服务删除用户
	err = c.userService.DeleteUser(id)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse(err.Error(), nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(resources.SuccessResponse("User deleted successfully", nil))
}

// UpdateProfile 更新个人资料
// @Summary Update profile
// @Description Update the current user's profile.
// @Tags Users
// @Accept json
// @Produce json
// @Param user body UpdateProfileRequest true "Update profile payload"
// @Security ApiKeyAuth
// @Success 200 {object} resources.APIResponse "OK"
// @Failure 400 {object} resources.APIResponse "Bad request"
// @Failure 401 {object} resources.APIResponse "Unauthorized"
// @Router /api/v1/users/profile [put]
func (c *UserController) UpdateProfile(ctx fiber.Ctx) error {
	// 从上下文中获取用户信息
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(resources.ErrorResponse("Unauthenticated user", nil))
	}

	req, reqErr := c.bindAndValidateUpdateProfileRequest(ctx)
	if reqErr != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse(reqErr.message, reqErr.details))
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
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse(err.Error(), nil))
	}

	// 获取更新后的用户信息
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
// @Summary Get current user
// @Description Get the current authenticated user's details.
// @Tags Profile
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} resources.APIResponse "OK"
// @Failure 401 {object} resources.APIResponse "Unauthorized"
// @Failure 404 {object} resources.APIResponse "User not found"
// @Router /api/me [get]
func (c *UserController) GetCurrentUser(ctx fiber.Ctx) error {
	// 从上下文中获取用户ID
	userID, ok := ctx.Locals("user_id").(int64)
	if !ok {
		return ctx.Status(fiber.StatusUnauthorized).JSON(resources.ErrorResponse("Unauthorized", nil))
	}

	// 获取完整的用户信息
	currentUser, err := c.userService.GetUserByID(userID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(resources.ErrorResponse("User not found", nil))
	}

	return ctx.Status(fiber.StatusOK).JSON(resources.SuccessResponse("User fetched successfully", fiber.Map{
		"user": currentUser.ToSafeUser(),
	}))
}

// SearchUsers 搜索用户
// @Summary Search users
// @Description Search users by keyword (admin only).
// @Tags Users
// @Accept json
// @Produce json
// @Param q query string true "Search keyword"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Security ApiKeyAuth
// @Success 200 {object} resources.APIResponse "OK"
// @Failure 400 {object} resources.APIResponse "Bad request"
// @Failure 401 {object} resources.APIResponse "Unauthorized"
// @Failure 500 {object} resources.APIResponse "Internal server error"
// @Router /api/users/search [get]
func (c *UserController) SearchUsers(ctx fiber.Ctx) error {
	// 获取搜索参数
	query := ctx.Query("q")
	if query == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(resources.ErrorResponse("Search keyword is required", nil))
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
		return ctx.Status(fiber.StatusInternalServerError).JSON(resources.ErrorResponse(err.Error(), nil))
	}

	// 转换为安全响应
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
