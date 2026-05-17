package user

import (
	"fmt"
	"strconv"

	middleware "fiber-starter/internal/common/middleware"
	helpers "fiber-starter/internal/support"

	"github.com/gofiber/fiber/v3"
)

// UserController 用户控制器
type UserController struct {
	userService UserService
}

// NewUserController 创建用户控制器实例
func NewUserController(userService UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// GetUsers 获取用户列表
func (c *UserController) GetUsers(ctx fiber.Ctx) error {
	var req UserListRequest
	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}
	page, err := c.userService.ListUsers(ctx.Context(), req.ToQuery())
	if err != nil {
		return helpers.HandleInternalServerError(ctx, err.Error())
	}

	return helpers.HandlePaginationResponse(ctx, "Users fetched successfully", NewUserCollection(page.Items), page.Total, page.Page, page.Limit)
}

// GetUser 获取单个用户
func (c *UserController) GetUser(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return helpers.HandleBadRequest(ctx, "Invalid user ID")
	}

	user, err := c.userService.GetUserByID(ctx.Context(), id)
	if err != nil {
		return helpers.HandleNotFound(ctx, err.Error())
	}

	return helpers.HandleSuccess(ctx, "User fetched successfully", NewUserResource(user).ToResponse())
}

// UpdateUser 更新用户信息
func (c *UserController) UpdateUser(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return helpers.HandleBadRequest(ctx, "Invalid user ID")
	}

	var req UpdateProfileRequest
	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	user, err := c.userService.UpdateUser(ctx.Context(), id, req.ToInput())
	if err != nil {
		return helpers.HandleBadRequest(ctx, err.Error())
	}

	return helpers.HandleSuccess(ctx, "User updated successfully", NewUserResource(user).ToResponse())
}

// DeleteUser 删除用户
func (c *UserController) DeleteUser(ctx fiber.Ctx) error {
	idStr := ctx.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return helpers.HandleBadRequest(ctx, "Invalid user ID")
	}

	if err := c.userService.DeleteUser(ctx.Context(), id); err != nil {
		return helpers.HandleBadRequest(ctx, err.Error())
	}

	return helpers.HandleSuccess(ctx, "User deleted successfully", nil)
}

// UpdateProfile 更新个人资料
func (c *UserController) UpdateProfile(ctx fiber.Ctx) error {
	userID := middleware.GetCurrentUserID(ctx)
	if userID == 0 {
		return helpers.HandleUnauthorized(ctx, "Unauthenticated user")
	}

	var req UpdateProfileRequest
	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	updatedUser, err := c.userService.UpdateProfile(ctx.Context(), userID, req.ToInput())
	if err != nil {
		return helpers.HandleBadRequest(ctx, err.Error())
	}

	return helpers.HandleSuccess(ctx, "Profile updated successfully", NewUserResource(updatedUser).ToResponse())
}

// GetCurrentUser 获取当前登录用户的信息
func (c *UserController) GetCurrentUser(ctx fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(int64)
	if !ok {
		return helpers.HandleUnauthorized(ctx, "Unauthorized")
	}

	currentUser, err := c.userService.GetUserByID(ctx.Context(), userID)
	if err != nil {
		return helpers.HandleNotFound(ctx, "User not found")
	}

	return helpers.HandleSuccess(ctx, "User fetched successfully", NewUserResource(currentUser).ToResponse())
}

// SearchUsers 搜索用户
func (c *UserController) SearchUsers(ctx fiber.Ctx) error {
	var req SearchUsersRequest
	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}
	page, err := c.userService.SearchUsers(ctx.Context(), req.ToQuery())
	if err != nil {
		return helpers.HandleInternalServerError(ctx, err.Error())
	}

	return helpers.HandlePaginationResponse(ctx, "Users searched successfully", NewUserCollection(page.Items), page.Total, page.Page, page.Limit)
}

// ExportUsers 导出用户列表
func (c *UserController) ExportUsers(ctx fiber.Ctx) error {
	// 获取所有用户（这里简单处理，获取前 1000 条）
	page, err := c.userService.ListUsers(ctx.Context(), UserListQuery{Page: 1, Limit: 1000})
	if err != nil {
		return helpers.HandleInternalServerError(ctx, "Failed to fetch users for export")
	}

	export := &UserExport{
		Users: page.Items,
	}

	excel := &helpers.Excel{}
	return excel.Download(ctx, export, "users_export.xlsx")
}

// ImportUsers 导入用户列表
func (c *UserController) ImportUsers(ctx fiber.Ctx) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		return helpers.HandleBadRequest(ctx, "File is required")
	}

	f, err := file.Open()
	if err != nil {
		return helpers.HandleInternalServerError(ctx, "Failed to open uploaded file")
	}
	defer func() { _ = f.Close() }()

	importer := &UserImport{}
	excel := &helpers.Excel{}

	models_any, err := excel.Import(f, importer)
	if err != nil {
		return helpers.HandleBadRequest(ctx, fmt.Sprintf("Failed to import Excel: %v", err))
	}

	// 为了演示，我们只返回导入的数量
	return helpers.HandleSuccess(ctx, fmt.Sprintf("Successfully imported %d users", len(models_any)), fiber.Map{
		"count": len(models_any),
	})
}
