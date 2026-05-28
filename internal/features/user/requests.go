package user

import "mime/multipart"

// UserURIRequest 用户路径参数请求。
type UserURIRequest struct {
	ID int64 `uri:"id" validate:"required,gt=0"`
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Name   string `json:"name" validate:"omitempty,min=2,max=100" example:"Alice"`
	Phone  string `json:"phone" validate:"omitempty,e164" example:"+8613800138000"`
	Avatar string `json:"avatar" validate:"omitempty,url" example:"https://example.com/avatar.jpg"` //nolint:lll
}

func (r UpdateProfileRequest) ToInput() UpdateUserInput {
	input := UpdateUserInput{}
	if r.Name != "" {
		input.Name = &r.Name
	}
	if r.Phone != "" {
		input.Phone = &r.Phone
	}
	if r.Avatar != "" {
		input.Avatar = &r.Avatar
	}
	return input
}

// UserListRequest 用户列表查询请求
type UserListRequest struct {
	Page  int `query:"page" validate:"omitempty,gte=1"`
	Limit int `query:"limit" validate:"omitempty,gte=1,lte=100"`
}

func (r UserListRequest) ToQuery() UserListQuery {
	r.Normalize()
	return UserListQuery{
		Page:  r.Page,
		Limit: r.Limit,
	}
}

// SearchUsersRequest 用户搜索查询请求
type SearchUsersRequest struct {
	Q     string `query:"q" validate:"required,min=1,max=100"`
	Page  int    `query:"page" validate:"omitempty,gte=1"`
	Limit int    `query:"limit" validate:"omitempty,gte=1,lte=100"`
}

func (r SearchUsersRequest) ToQuery() UserListQuery {
	r.Normalize()
	return UserListQuery{
		Search: r.Q,
		Page:   r.Page,
		Limit:  r.Limit,
	}
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
	query := r.ToQuery()
	return query.Page, query.Limit
}

// ImportUsersRequest 用户导入表单请求。
type ImportUsersRequest struct {
	File *multipart.FileHeader `form:"file" validate:"required,uploaded_file"`
}
