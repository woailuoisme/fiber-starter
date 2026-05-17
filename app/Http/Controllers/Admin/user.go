package Admin

import (
	"fmt"
	"strconv"

	exports "fiber-starter/app/Exports"
	middleware "fiber-starter/app/Http/Middleware"
	requests "fiber-starter/app/Http/Requests"
	resources "fiber-starter/app/Http/Resources"
	services "fiber-starter/app/Http/Services"
	imports "fiber-starter/app/Imports"
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
//
//	@Summary		获取用户列表
//	@Description	获取分页用户列表（需管理员权限）。
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Param			page	query	int	false	"页码"	default(1)
//	@Param			limit	query	int	false	"每页数量"	default(10)
//	@Security		Bearer
//	@Success		200	{object}	helpers.APIResponse{data=object{items=[]models.SafeUser,meta=helpers.PaginationMeta}}	"成功"
//	@Router			/api/v1/users [get]
func (c *UserController) GetUsers(ctx fiber.Ctx) error {
	var req requests.UserListRequest
	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}
	page, err := c.userService.ListUsers(ctx.Context(), req.ToQuery())
	if err != nil {
		return helpers.HandleInternalServerError(ctx, err.Error())
	}

	return helpers.HandlePaginationResponse(ctx, "Users fetched successfully", resources.NewUserCollection(page.Items), page.Total, page.Page, page.Limit)
}

// GetUser 获取单个用户
//
//	@Summary		获取单个用户
//	@Description	根据 ID 获取用户详情（需管理员权限）。
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"用户 ID"	example(1)
//	@Security		Bearer
//	@Success		200	{object}	helpers.APIResponse{data=object{user=models.SafeUser}}	"成功"
//	@Success		200	{object}	helpers.APIResponse{data=object{user=models.SafeUser}}	"成功"
//	@Router			/api/v1/users/{id} [get]
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

	return helpers.HandleSuccess(ctx, "User fetched successfully", resources.NewUserResource(user).ToResponse())
}

// UpdateUser 更新用户信息
//
//	@Summary		更新用户信息
//	@Description	根据 ID 更新指定用户信息（需管理员权限）。
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Param			id		path	int								true	"用户 ID"	example(1)
//	@Param			user	body	requests.UpdateProfileRequest	true	"更新参数"
//	@Security		Bearer
//	@Success		200	{object}	helpers.APIResponse{data=object{user=models.SafeUser}}	"更新成功"
//	@Success		200	{object}	helpers.APIResponse{data=object{user=models.SafeUser}}	"更新成功"
//	@Router			/api/v1/users/{id} [put]
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

	user, err := c.userService.UpdateUser(ctx.Context(), id, req.ToInput())
	if err != nil {
		return helpers.HandleBadRequest(ctx, err.Error())
	}

	return helpers.HandleSuccess(ctx, "User updated successfully", resources.NewUserResource(user).ToResponse())
}

// DeleteUser 删除用户
//
//	@Summary		删除用户
//	@Description	根据 ID 永久删除用户账号（需管理员权限）。
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Param			id	path	int	true	"用户 ID"	example(1)
//	@Security		Bearer
//	@Success		200	{object}	helpers.APIResponse	"删除成功"
//	@Success		200	{object}	helpers.APIResponse	"删除成功"
//	@Router			/api/v1/users/{id} [delete]
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
//
//	@Summary		更新个人资料
//	@Description	更新当前登录用户的个人资料。
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Param			user	body	requests.UpdateProfileRequest	true	"更新参数"
//	@Security		Bearer
//	@Success		200	{object}	helpers.APIResponse{data=object{user=models.SafeUser}}	"更新成功"
//	@Success		200	{object}	helpers.APIResponse{data=object{user=models.SafeUser}}	"更新成功"
//	@Router			/api/v1/users/profile [put]
func (c *UserController) UpdateProfile(ctx fiber.Ctx) error {
	user := middleware.GetUserFromContext(ctx)
	if user == nil {
		return helpers.HandleUnauthorized(ctx, "Unauthenticated user")
	}

	var req requests.UpdateProfileRequest
	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}

	updatedUser, err := c.userService.UpdateProfile(ctx.Context(), user.ID, req.ToInput())
	if err != nil {
		return helpers.HandleBadRequest(ctx, err.Error())
	}

	return helpers.HandleSuccess(ctx, "Profile updated successfully", resources.NewUserResource(updatedUser).ToResponse())
}

// GetCurrentUser 获取当前登录用户的信息
//
//	@Summary		获取当前登录用户信息
//	@Description	根据访问令牌返回当前认证用户的详细信息。
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Security		Bearer
//	@Success		200	{object}	helpers.APIResponse{data=object{user=models.SafeUser}}	"获取成功"
//	@Router			/api/v1/users/me [get]
func (c *UserController) GetCurrentUser(ctx fiber.Ctx) error {
	userID, ok := ctx.Locals("user_id").(int64)
	if !ok {
		return helpers.HandleUnauthorized(ctx, "Unauthorized")
	}

	currentUser, err := c.userService.GetUserByID(ctx.Context(), userID)
	if err != nil {
		return helpers.HandleNotFound(ctx, "User not found")
	}

	return helpers.HandleSuccess(ctx, "User fetched successfully", resources.NewUserResource(currentUser).ToResponse())
}

// SearchUsers 搜索用户
//
//	@Summary		搜索用户
//	@Description	根据姓名或邮箱搜索分页用户列表（需管理员权限）。
//	@Tags			用户管理
//	@Accept			json
//	@Produce		json
//	@Param			q		query	string	true	"关键词"	example("admin")
//	@Param			page	query	int		false	"页码"	default(1)
//	@Param			limit	query	int		false	"每页数量"	default(10)
//	@Security		Bearer
//	@Success		200	{object}	helpers.APIResponse{data=object{items=[]models.SafeUser,meta=helpers.PaginationMeta}}	"搜索成功"
//	@Router			/api/v1/users/search [get]
func (c *UserController) SearchUsers(ctx fiber.Ctx) error {
	var req requests.SearchUsersRequest
	if err := req.BindAndValidate(ctx); err != nil {
		return helpers.HandleAppError(ctx, err)
	}
	page, err := c.userService.SearchUsers(ctx.Context(), req.ToQuery())
	if err != nil {
		return helpers.HandleInternalServerError(ctx, err.Error())
	}

	return helpers.HandlePaginationResponse(ctx, "Users searched successfully", resources.NewUserCollection(page.Items), page.Total, page.Page, page.Limit)
}

// ExportUsers 导出用户列表
//
//	@Summary		导出用户列表
//	@Description	将系统中的用户列表导出为 Excel 文件下载（需管理员权限）。
//	@Tags			用户管理
//	@Produce		application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
//	@Security		Bearer
//	@Success		200	{file}	file	"users_export.xlsx"
//	@Router			/api/v1/users/export [get]
func (c *UserController) ExportUsers(ctx fiber.Ctx) error {
	// 获取所有用户（这里简单处理，获取前 1000 条）
	page, err := c.userService.ListUsers(ctx.Context(), services.UserListQuery{Page: 1, Limit: 1000})
	if err != nil {
		return helpers.HandleInternalServerError(ctx, "Failed to fetch users for export")
	}

	export := &exports.UserExport{
		Users: page.Items,
	}

	excel := &helpers.Excel{}
	return excel.Download(ctx, export, "users_export.xlsx")
}

// ImportUsers 导入用户列表
//
//	@Summary		导入用户列表
//	@Description	上传 Excel 文件并批量导入用户信息（需管理员权限）。
//	@Tags			用户管理
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"Excel 文件"
//	@Security		Bearer
//	@Success		200	{object}	helpers.APIResponse{data=object{count=int}}	"导入成功"
//	@Success		200	{object}	helpers.APIResponse{data=object{count=int}}	"导入成功"
//	@Router			/api/v1/users/import [post]
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

	importer := &imports.UserImport{}
	excel := &helpers.Excel{}

	models_any, err := excel.Import(f, importer)
	if err != nil {
		return helpers.HandleBadRequest(ctx, fmt.Sprintf("Failed to import Excel: %v", err))
	}

	// 这里可以进一步处理导入的模型，例如保存到数据库
	// 为了演示，我们只返回导入的数量
	return helpers.HandleSuccess(ctx, fmt.Sprintf("Successfully imported %d users", len(models_any)), fiber.Map{
		"count": len(models_any),
	})
}
