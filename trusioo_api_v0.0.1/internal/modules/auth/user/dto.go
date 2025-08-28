package user

import (
	"trusioo_api_v0.0.1/internal/modules/auth"
)

// ========== 请求 DTO ==========

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
	Name     string `json:"name" binding:"required,min=2" example:"John Doe"`
}

// SimpleRegisterRequest 简化注册请求（仅需email和password）
type SimpleRegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// LoginRequest 用户登录请求（发送验证码）
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// VerifyLoginRequest 登录验证请求
type VerifyLoginRequest struct {
	Email     string `json:"email" binding:"required,email" example:"user@example.com"`
	Password  string `json:"password" binding:"required,min=6" example:"password123"`
	LoginCode string `json:"login_code" binding:"required,len=6" example:"123456"`
	UserAgent string `json:"user_agent" binding:"omitempty" example:"Mozilla/5.0..."`
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
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Email            string `json:"email" binding:"required,email" example:"user@example.com"`
	VerificationCode string `json:"verification_code" binding:"required,len=6" example:"123456"`
	NewPassword      string `json:"new_password" binding:"required,min=6" example:"newpassword123"`
}

// UpdateProfileRequest 更新用户资料请求
type UpdateProfileRequest struct {
	Name  string `json:"name" binding:"omitempty,min=2" example:"Jane Doe"`
	Email string `json:"email" binding:"omitempty,email" example:"updated@example.com"`
}

// ListUsersRequest 用户列表查询请求
type ListUsersRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1" example:"1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100" example:"10"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive suspended" example:"active"`
	Search   string `form:"search" binding:"omitempty" example:"john"`
}

// UpdateUserStatusRequest 更新用户状态请求
type UpdateUserStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive suspended" example:"active"`
}

// ========== 响应 DTO ==========

// RegisterResponse 注册响应
type RegisterResponse struct {
	Message string    `json:"message" example:"Registration successful"`
	User    *UserInfo `json:"user"`
}

// SimpleRegisterResponse 简化注册响应
type SimpleRegisterResponse struct {
	Message string `json:"message" example:"Registration successful"`
	UserID  string `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// LoginResponse 登录响应（发送验证码）
type LoginResponse struct {
	Message          string `json:"message" example:"Verification code sent"`
	VerificationCode string `json:"verification_code" example:"123456"` // 仅用于测试，生产环境应删除
	ExpiresIn        int    `json:"expires_in" example:"300"`
}

// VerifyLoginResponse 验证登录响应
type VerifyLoginResponse struct {
	Message string           `json:"message" example:"Login successful"`
	User    *UserInfo        `json:"user"`
	Tokens  *auth.TokenPair  `json:"tokens"`
	Session *UserSessionInfo `json:"session"` // 新增会话信息
}

// UserInfo 用户信息结构（用于API响应）
type UserInfo struct {
	ID            string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email         string `json:"email" example:"user@example.com"`
	Name          string `json:"name" example:"John Doe"`
	Status        string `json:"status" example:"active"`
	EmailVerified bool   `json:"email_verified" example:"true"`
}

// RefreshTokenResponse 刷新令牌响应
type RefreshTokenResponse struct {
	Message string          `json:"message" example:"Token refreshed successfully"`
	Tokens  *auth.TokenPair `json:"tokens"`
}

// ProfileResponse 获取资料响应
type ProfileResponse struct {
	User *UserInfo `json:"user"`
}

// ForgotPasswordResponse 忘记密码响应
type ForgotPasswordResponse struct {
	Message          string `json:"message" example:"Password reset code sent"`
	Email            string `json:"email" example:"user@example.com"`
	VerificationCode string `json:"verification_code" example:"123456"` // 仅用于测试，生产环境应删除
	ExpiresIn        int    `json:"expires_in" example:"300"`
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Message string `json:"message" example:"Password reset successfully"`
}

// UpdateProfileResponse 更新资料响应
type UpdateProfileResponse struct {
	Message string    `json:"message" example:"Profile updated successfully"`
	User    *UserInfo `json:"user"`
}

// UserListItem 用户列表项
type UserListItem struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string `json:"email" example:"user@example.com"`
	Name      string `json:"name" example:"John Doe"`
	Status    string `json:"status" example:"active"`
	CreatedAt string `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// ListUsersResponse 用户列表响应
type ListUsersResponse struct {
	Users      []*UserListItem `json:"users"`
	Pagination *PaginationInfo `json:"pagination"`
}

// UpdateUserStatusResponse 更新用户状态响应
type UpdateUserStatusResponse struct {
	Message string    `json:"message" example:"User status updated successfully"`
	User    *UserInfo `json:"user"`
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

// ToUserInfo 将User模型转换为UserInfo DTO
func (u *User) ToUserInfo() *UserInfo {
	return &UserInfo{
		ID:            u.ID,
		Email:         u.Email,
		Name:          u.Name,
		Status:        u.Status,
		EmailVerified: u.EmailVerified,
	}
}

// ToUserListItem 将User模型转换为UserListItem DTO
func (u *User) ToUserListItem() *UserListItem {
	return &UserListItem{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Status:    u.Status,
		CreatedAt: u.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ToUser 将RegisterRequest转换为User模型
func (req *RegisterRequest) ToUser() *User {
	return &User{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,              // 注意：这里应该在service层进行加密
		Status:   UserStatusActive.String(), // 默认激活
	}
}

// ToUser 将SimpleRegisterRequest转换为User模型
func (req *SimpleRegisterRequest) ToUser() *User {
	return &User{
		Email:    req.Email,
		Name:     req.Email,                 // 使用邮箱作为默认名称
		Password: req.Password,              // 注意：这里应该在service层进行加密
		Status:   UserStatusActive.String(), // 默认激活
	}
}

// ApplyUpdate 将UpdateProfileRequest的更新应用到User模型
func (req *UpdateProfileRequest) ApplyUpdate(user *User) {
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
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
func (req *ListUsersRequest) GetOffset() int {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	return (req.Page - 1) * req.PageSize
}

// GetLimit 获取查询限制数量
func (req *ListUsersRequest) GetLimit() int {
	if req.PageSize <= 0 {
		return 10
	}
	if req.PageSize > 100 {
		return 100
	}
	return req.PageSize
}

// ========== 会话相关 DTO ==========

// UserSessionInfo 用户会话信息（简化版）
type UserSessionInfo struct {
	ID           string                  `json:"id"`
	SessionID    string                  `json:"session_id"`
	IPAddress    string                  `json:"ip_address"`
	UserAgent    *string                 `json:"user_agent"`
	DeviceInfo   *map[string]interface{} `json:"device_info"`
	LocationInfo *map[string]interface{} `json:"location_info"`
	IsActive     bool                    `json:"is_active"`
	LastActivity string                  `json:"last_activity"`
	ExpiresAt    string                  `json:"expires_at"`
	CreatedAt    string                  `json:"created_at"`
}

// ToUserSessionInfo 将UserSession模型转换为UserSessionInfo DTO
func (s *UserSession) ToUserSessionInfo() *UserSessionInfo {
	return &UserSessionInfo{
		ID:           s.ID,
		SessionID:    s.SessionID,
		IPAddress:    s.IPAddress,
		UserAgent:    s.UserAgent,
		DeviceInfo:   s.DeviceInfo,
		LocationInfo: s.LocationInfo,
		IsActive:     s.IsActive,
		LastActivity: s.LastActivity.Format("2006-01-02T15:04:05Z"),
		ExpiresAt:    s.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:    s.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// UserSessionsResponse 用户会话列表响应
type UserSessionsResponse struct {
	Sessions []*UserSessionInfo `json:"sessions"`
	Total    int                `json:"total"`
}

// ========== 验证方法 ==========

// Validate 验证RegisterRequest
func (req *RegisterRequest) Validate() error {
	if req.Email == "" {
		return auth.ErrInvalidEmail
	}
	if len(req.Password) < 6 {
		return auth.ErrPasswordTooShort
	}
	if len(req.Name) < 2 {
		return auth.ErrNameTooShort
	}
	return nil
}

// Validate 验证SimpleRegisterRequest
func (req *SimpleRegisterRequest) Validate() error {
	if req.Email == "" {
		return auth.ErrInvalidEmail
	}
	if len(req.Password) < 6 {
		return auth.ErrPasswordTooShort
	}
	return nil
}

// Validate 验证LoginRequest
func (req *LoginRequest) Validate() error {
	if req.Email == "" {
		return auth.ErrInvalidEmail
	}
	if req.Password == "" {
		return auth.ErrPasswordRequired
	}
	return nil
}

// Validate 验证VerifyLoginRequest
func (req *VerifyLoginRequest) Validate() error {
	if req.Email == "" {
		return auth.ErrInvalidEmail
	}
	if req.Password == "" {
		return auth.ErrPasswordRequired
	}
	if len(req.LoginCode) != 6 {
		return auth.ErrInvalidVerificationCode
	}
	return nil
}

// ========== 登录日志相关 DTO ==========

// LoginLogInfo 登录日志信息
type LoginLogInfo struct {
	ID            string                  `json:"id"`
	UserID        *string                 `json:"user_id"`
	Email         string                  `json:"email"`
	UserType      string                  `json:"user_type"`
	LoginStatus   string                  `json:"login_status"`
	FailureReason *string                 `json:"failure_reason"`
	IPAddress     string                  `json:"ip_address"`
	UserAgent     *string                 `json:"user_agent"`
	DeviceInfo    *map[string]interface{} `json:"device_info"`
	LocationInfo  *map[string]interface{} `json:"location_info"`
	SessionID     *string                 `json:"session_id"`
	RiskScore     int                     `json:"risk_score"`
	CreatedAt     string                  `json:"created_at"`
}

// ToLoginLogInfo 将LoginLog模型转换为LoginLogInfo DTO
func (l *LoginLog) ToLoginLogInfo() *LoginLogInfo {
	return &LoginLogInfo{
		ID:            l.ID,
		UserID:        l.UserID,
		Email:         l.Email,
		UserType:      l.UserType,
		LoginStatus:   l.LoginStatus,
		FailureReason: l.FailureReason,
		IPAddress:     l.IPAddress,
		UserAgent:     l.UserAgent,
		DeviceInfo:    l.DeviceInfo,
		LocationInfo:  l.LocationInfo,
		SessionID:     l.SessionID,
		RiskScore:     l.RiskScore,
		CreatedAt:     l.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// UserLoginLogsResponse 用户登录日志列表响应
type UserLoginLogsResponse struct {
	Logs       []*LoginLogInfo `json:"logs"`
	Pagination *PaginationInfo `json:"pagination"`
}

// GetUserLoginLogsRequest 获取用户登录日志请求
type GetUserLoginLogsRequest struct {
	UserID   string `form:"user_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Page     int    `form:"page" binding:"omitempty,min=1" example:"1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100" example:"10"`
}

// GetOffset 计算数据库查询的偏移量
func (req *GetUserLoginLogsRequest) GetOffset() int {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	return (req.Page - 1) * req.PageSize
}

// GetLimit 获取查询限制数量
func (req *GetUserLoginLogsRequest) GetLimit() int {
	if req.PageSize <= 0 {
		return 10
	}
	if req.PageSize > 100 {
		return 100
	}
	return req.PageSize
}
