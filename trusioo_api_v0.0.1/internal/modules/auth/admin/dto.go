package admin

import (
	"trusioo_api_v0.0.1/internal/modules/auth"
)

// ========== 请求 DTO ==========

// LoginRequest 管理员登录请求（发送验证码）
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"admin@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// VerifyLoginRequest 管理员登录验证请求
type VerifyLoginRequest struct {
	Email     string `json:"email" binding:"required,email" example:"admin@example.com"`
	Password  string `json:"password" binding:"required,min=6" example:"password123"`
	LoginCode string `json:"login_code" binding:"required,len=6" example:"123456"`
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required" example:"oldpassword123"`
	NewPassword     string `json:"new_password" binding:"required,min=6" example:"newpassword123"`
}

// ForgotPasswordRequest 忘记密码请求
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email" example:"admin@example.com"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email            string `json:"email" binding:"required,email" example:"admin@example.com"`
	VerificationCode string `json:"verification_code" binding:"required,len=6" example:"123456"`
	NewPassword      string `json:"new_password" binding:"required,min=6" example:"newpassword123"`
}

// CreateAdminRequest 创建管理员请求（系统内部使用）
type CreateAdminRequest struct {
	Email    string `json:"email" binding:"required,email" example:"newadmin@example.com"`
	Name     string `json:"name" binding:"required,min=2" example:"John Doe"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
	Role     string `json:"role" binding:"required,oneof=super_admin admin" example:"admin"`
}

// UpdateAdminRequest 更新管理员信息请求
type UpdateAdminRequest struct {
	Name   string `json:"name" binding:"omitempty,min=2" example:"Jane Doe"`
	Email  string `json:"email" binding:"omitempty,email" example:"updated@example.com"`
	Role   string `json:"role" binding:"omitempty,oneof=super_admin admin" example:"admin"`
	Active *bool  `json:"active" binding:"omitempty" example:"true"`
}

// ListAdminsRequest 管理员列表查询请求
type ListAdminsRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1" example:"1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100" example:"10"`
	Role     string `form:"role" binding:"omitempty,oneof=super_admin admin" example:"admin"`
	Active   *bool  `form:"active" binding:"omitempty" example:"true"`
	Search   string `form:"search" binding:"omitempty" example:"john"`
}

// ========== 响应 DTO ==========

// LoginResponse 登录响应（发送验证码）
type LoginResponse struct {
	Message          string `json:"message" example:"Verification code sent"`
	VerificationCode string `json:"verification_code" example:"123456"` // 仅用于测试，生产环境应删除
	ExpiresIn        int    `json:"expires_in" example:"300"`
}

// VerifyLoginResponse 验证登录响应
type VerifyLoginResponse struct {
	Message string          `json:"message" example:"Login successful"`
	Admin   *AdminInfo      `json:"admin"`
	Tokens  *auth.TokenPair `json:"tokens"`
}

// AdminInfo 管理员信息结构（用于API响应）
type AdminInfo struct {
	ID     string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email  string `json:"email" example:"admin@example.com"`
	Name   string `json:"name" example:"John Doe"`
	Role   string `json:"role" example:"admin"`
	Active bool   `json:"active" example:"true"`
}

// RefreshTokenResponse 刷新令牌响应
type RefreshTokenResponse struct {
	Message string          `json:"message" example:"Token refreshed successfully"`
	Tokens  *auth.TokenPair `json:"tokens"`
}

// ProfileResponse 获取资料响应
type ProfileResponse struct {
	Admin *AdminInfo `json:"admin"`
}

// ForgotPasswordResponse 忘记密码响应
type ForgotPasswordResponse struct {
	Message          string `json:"message" example:"Password reset code sent"`
	Email            string `json:"email" example:"admin@example.com"`
	VerificationCode string `json:"verification_code" example:"123456"` // 仅用于测试，生产环境应删除
	ExpiresIn        int    `json:"expires_in" example:"300"`
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Message string `json:"message" example:"Password reset successfully"`
}

// CreateAdminResponse 创建管理员响应
type CreateAdminResponse struct {
	Message string     `json:"message" example:"Admin created successfully"`
	Admin   *AdminInfo `json:"admin"`
}

// UpdateAdminResponse 更新管理员响应
type UpdateAdminResponse struct {
	Message string     `json:"message" example:"Admin updated successfully"`
	Admin   *AdminInfo `json:"admin"`
}

// AdminListItem 管理员列表项
type AdminListItem struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string `json:"email" example:"admin@example.com"`
	Name      string `json:"name" example:"John Doe"`
	Role      string `json:"role" example:"admin"`
	Active    bool   `json:"active" example:"true"`
	CreatedAt string `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// ListAdminsResponse 管理员列表响应
type ListAdminsResponse struct {
	Admins     []*AdminListItem `json:"admins"`
	Pagination *PaginationInfo  `json:"pagination"`
}

// PaginationInfo 分页信息
type PaginationInfo struct {
	Page       int   `json:"page" example:"1"`
	PageSize   int   `json:"page_size" example:"10"`
	Total      int64 `json:"total" example:"100"`
	TotalPages int   `json:"total_pages" example:"10"`
}

// ========== 通用响应 DTO ==========

// MessageResponse 通用消息响应
type MessageResponse struct {
	Message string `json:"message" example:"Operation successful"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error" example:"Bad Request"`
	Message string `json:"message" example:"Invalid input parameters"`
	Details any    `json:"details,omitempty"`
}

// ========== 数据转换方法 ==========

// ToAdminInfo 将Admin模型转换为AdminInfo DTO
func (a *Admin) ToAdminInfo() *AdminInfo {
	return &AdminInfo{
		ID:     a.ID,
		Email:  a.Email,
		Name:   a.Name,
		Role:   a.Role,
		Active: a.Active,
	}
}

// ToAdminListItem 将Admin模型转换为AdminListItem DTO
func (a *Admin) ToAdminListItem() *AdminListItem {
	return &AdminListItem{
		ID:        a.ID,
		Email:     a.Email,
		Name:      a.Name,
		Role:      a.Role,
		Active:    a.Active,
		CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ToAdmin 将CreateAdminRequest转换为Admin模型
func (req *CreateAdminRequest) ToAdmin() *Admin {
	return &Admin{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password, // 注意：这里应该在service层进行加密
		Role:     req.Role,
		Active:   true, // 默认激活
	}
}

// ApplyUpdate 将UpdateAdminRequest的更新应用到Admin模型
func (req *UpdateAdminRequest) ApplyUpdate(admin *Admin) {
	if req.Name != "" {
		admin.Name = req.Name
	}
	if req.Email != "" {
		admin.Email = req.Email
	}
	if req.Role != "" {
		admin.Role = req.Role
	}
	if req.Active != nil {
		admin.Active = *req.Active
	}
}

// CalculatePagination 计算分页信息
func CalculatePagination(page, pageSize int, total int64) *PaginationInfo {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	return &PaginationInfo{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}

// GetOffset 计算数据库查询的偏移量
func (req *ListAdminsRequest) GetOffset() int {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	return (req.Page - 1) * req.PageSize
}

// GetLimit 获取查询限制数量
func (req *ListAdminsRequest) GetLimit() int {
	if req.PageSize <= 0 {
		return 10
	}
	if req.PageSize > 100 {
		return 100
	}
	return req.PageSize
}
