package requests

import (
	"github.com/gofiber/fiber/v3"
)

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Name   string `json:"name" validate:"omitempty,min=2,max=100" example:"Alice"`
	Phone  string `json:"phone" validate:"omitempty,e164" example:"+8613800138000"`
	Avatar string `json:"avatar" validate:"omitempty,url" example:"https://example.com/avatar.jpg"` //nolint:lll
}

func (r *UpdateProfileRequest) BindAndValidate(c fiber.Ctx) error {
	return BindAndValidateBody(c, r)
}

// UserListRequest 用户列表查询请求
type UserListRequest struct {
	Page  int `query:"page"`
	Limit int `query:"limit"`
}

func (r *UserListRequest) BindAndValidate(c fiber.Ctx) error {
	if err := BindAndValidateQuery(c, r); err != nil {
		return err
	}
	r.Normalize()
	return nil
}

// SearchUsersRequest 用户搜索查询请求
type SearchUsersRequest struct {
	Q     string `query:"q" validate:"required"`
	Page  int    `query:"page"`
	Limit int    `query:"limit"`
}

func (r *SearchUsersRequest) BindAndValidate(c fiber.Ctx) error {
	if err := BindAndValidateQuery(c, r); err != nil {
		return err
	}
	r.Normalize()
	return nil
}

// Normalize 将分页参数规范化为当前项目默认值。
func (r *UserListRequest) Normalize() {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.Limit < 1 || r.Limit > 100 {
		r.Limit = 10
	}
}

// Normalize 将分页参数规范化为当前项目默认值。
func (r *SearchUsersRequest) Normalize() {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.Limit < 1 || r.Limit > 100 {
		r.Limit = 10
	}
}

func (r UserListRequest) Pagination() (int, int) {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.Limit < 1 || r.Limit > 100 {
		r.Limit = 10
	}
	return r.Page, r.Limit
}

func (r SearchUsersRequest) Pagination() (int, int) {
	if r.Page < 1 {
		r.Page = 1
	}
	if r.Limit < 1 || r.Limit > 100 {
		r.Limit = 10
	}
	return r.Page, r.Limit
}
