package user_management

import (
	"time"

	"trusioo_api_v0.0.1/internal/modules/auth/user"
)

// === 请求DTO ===

// GetUsersRequest 获取用户列表请求
type GetUsersRequest struct {
	// 分页参数
	Page     int    `form:"page" binding:"omitempty,min=1" example:"1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100" example:"20"`
	SortBy   string `form:"sort_by" binding:"omitempty,oneof=created_at updated_at email name last_login_at login_count" example:"created_at"`
	SortDir  string `form:"sort_dir" binding:"omitempty,oneof=asc desc" example:"desc"`

	// 搜索过滤器
	Email         string `form:"email" binding:"omitempty,email" example:"user@example.com"`
	Name          string `form:"name" binding:"omitempty" example:"张三"`
	Status        string `form:"status" binding:"omitempty,oneof=active inactive suspended" example:"active"`
	EmailVerified *bool  `form:"email_verified" binding:"omitempty" example:"true"`

	// 时间范围过滤
	CreatedFrom   string `form:"created_from" binding:"omitempty" example:"2024-01-01"`
	CreatedTo     string `form:"created_to" binding:"omitempty" example:"2024-12-31"`
	LastLoginFrom string `form:"last_login_from" binding:"omitempty" example:"2024-01-01"`
	LastLoginTo   string `form:"last_login_to" binding:"omitempty" example:"2024-12-31"`

	// 搜索关键词（支持邮箱、姓名模糊搜索）
	Search string `form:"search" binding:"omitempty" example:"张三"`
}

// UpdateUserStatusRequest 更新用户状态请求
type UpdateUserStatusRequest struct {
	Status string  `json:"status" binding:"required,oneof=active inactive suspended" example:"suspended"`
	Reason *string `json:"reason" binding:"omitempty" example:"违反用户协议"`
}

// SuspendUserRequest 暂停用户请求
type SuspendUserRequest struct {
	Reason   string     `json:"reason" binding:"required" example:"违反用户协议"`
	Duration *int       `json:"duration" binding:"omitempty,min=1" example:"7"` // 暂停天数，nil表示永久暂停
	Until    *time.Time `json:"until" binding:"omitempty" example:"2024-12-31T23:59:59Z"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	NewPassword      string `json:"new_password" binding:"required,min=6" example:"newpassword123"`
	SendNotification bool   `json:"send_notification" example:"true"` // 是否发送邮件通知用户
}

// UpdateUserEmailRequest 更新用户邮箱请求
type UpdateUserEmailRequest struct {
	NewEmail         string `json:"new_email" binding:"required,email" example:"newemail@example.com"`
	SendVerification bool   `json:"send_verification" example:"true"` // 是否发送验证邮件
}

// ForceLogoutRequest 强制登出请求
type ForceLogoutRequest struct {
	LogoutAll bool   `json:"logout_all" example:"true"` // 是否登出所有会话
	Reason    string `json:"reason" binding:"required" example:"安全原因"`
}

// GetStatisticsRequest 获取统计信息请求
type GetStatisticsRequest struct {
	DateFrom   string `form:"date_from" binding:"omitempty" example:"2024-01-01"`
	DateTo     string `form:"date_to" binding:"omitempty" example:"2024-12-31"`
	GroupBy    string `form:"group_by" binding:"omitempty,oneof=day week month" example:"day"`
	MetricType string `form:"metric_type" binding:"omitempty,oneof=registration verification activity" example:"registration"`
}

// ExportUsersRequest 导出用户请求
type ExportUsersRequest struct {
	Format   string `form:"format" binding:"omitempty,oneof=csv excel" example:"csv"`
	Fields   string `form:"fields" binding:"omitempty" example:"id,email,name,status,created_at"`
	Timezone string `form:"timezone" binding:"omitempty" example:"Asia/Shanghai"`

	// 继承过滤器参数
	GetUsersRequest
}

// === 响应DTO ===

// UserDetailResponse 用户详情响应
type UserDetailResponse struct {
	// 基础信息
	ID              string     `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email           string     `json:"email" example:"user@example.com"`
	Name            string     `json:"name" example:"张三"`
	Status          string     `json:"status" example:"active"`
	EmailVerified   bool       `json:"email_verified" example:"true"`
	EmailVerifiedAt *time.Time `json:"email_verified_at" example:"2024-01-15T10:30:00Z"`

	// 活动信息
	LastLoginAt    *time.Time `json:"last_login_at" example:"2024-01-20T14:30:00Z"`
	LoginCount     int64      `json:"login_count" example:"25"`
	FailedAttempts int        `json:"failed_attempts" example:"0"`
	ActiveSessions int        `json:"active_sessions" example:"2"`

	// 管理信息
	SuspendedAt     *time.Time `json:"suspended_at,omitempty" example:"2024-01-22T09:00:00Z"`
	SuspendedReason *string    `json:"suspended_reason,omitempty" example:"违反用户协议"`
	Notes           *string    `json:"notes,omitempty" example:"VIP用户"`

	// 审计信息
	CreatedAt time.Time  `json:"created_at" example:"2024-01-01T08:00:00Z"`
	UpdatedAt time.Time  `json:"updated_at" example:"2024-01-20T14:30:00Z"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	Users      []UserSummaryResponse `json:"users"`
	Total      int64                 `json:"total" example:"150"`
	Page       int                   `json:"page" example:"1"`
	PageSize   int                   `json:"page_size" example:"20"`
	TotalPages int                   `json:"total_pages" example:"8"`
	HasNext    bool                  `json:"has_next" example:"true"`
	HasPrev    bool                  `json:"has_prev" example:"false"`
}

// UserSummaryResponse 用户摘要响应（用于列表）
type UserSummaryResponse struct {
	ID             string     `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email          string     `json:"email" example:"user@example.com"`
	Name           string     `json:"name" example:"张三"`
	Status         string     `json:"status" example:"active"`
	EmailVerified  bool       `json:"email_verified" example:"true"`
	LastLoginAt    *time.Time `json:"last_login_at" example:"2024-01-20T14:30:00Z"`
	LoginCount     int64      `json:"login_count" example:"25"`
	ActiveSessions int        `json:"active_sessions" example:"2"`
	CreatedAt      time.Time  `json:"created_at" example:"2024-01-01T08:00:00Z"`
}

// StatisticsResponse 统计信息响应
type StatisticsResponse struct {
	// 总体统计
	TotalUsers        int64 `json:"total_users" example:"1500"`
	NewUsersToday     int64 `json:"new_users_today" example:"12"`
	NewUsersThisWeek  int64 `json:"new_users_this_week" example:"85"`
	NewUsersThisMonth int64 `json:"new_users_this_month" example:"320"`

	// 验证状态统计
	VerifiedUsers    int64   `json:"verified_users" example:"1200"`
	UnverifiedUsers  int64   `json:"unverified_users" example:"300"`
	VerificationRate float64 `json:"verification_rate" example:"80.0"`

	// 用户状态统计
	ActiveUsers    int64            `json:"active_users" example:"1200"`
	InactiveUsers  int64            `json:"inactive_users" example:"250"`
	SuspendedUsers int64            `json:"suspended_users" example:"45"`
	DeletedUsers   int64            `json:"deleted_users" example:"5"`
	UsersByStatus  map[string]int64 `json:"users_by_status"`

	// 活跃度统计
	ActiveUsersToday     int64 `json:"active_users_today" example:"450"`
	ActiveUsersThisWeek  int64 `json:"active_users_this_week" example:"850"`
	ActiveUsersThisMonth int64 `json:"active_users_this_month" example:"1100"`

	// 时间趋势
	RegistrationTrend []DailyCountResponse `json:"registration_trend"`
	VerificationTrend []DailyCountResponse `json:"verification_trend"`

	// 生成时间
	GeneratedAt time.Time `json:"generated_at" example:"2024-01-22T15:30:00Z"`
}

// DailyCountResponse 每日计数响应
type DailyCountResponse struct {
	Date  string `json:"date" example:"2024-01-22"`
	Count int64  `json:"count" example:"15"`
}

// UserActivityResponse 用户活动响应
type UserActivityResponse struct {
	UserID string `json:"user_id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email  string `json:"email" example:"user@example.com"`
	Name   string `json:"name" example:"张三"`

	// 登录活动
	LastLoginAt         *time.Time `json:"last_login_at" example:"2024-01-20T14:30:00Z"`
	TotalLogins         int64      `json:"total_logins" example:"25"`
	FailedLoginAttempts int        `json:"failed_login_attempts" example:"0"`

	// 会话信息
	ActiveSessions []SessionSummary `json:"active_sessions"`
	LastActivityAt *time.Time       `json:"last_activity_at" example:"2024-01-22T10:15:00Z"`

	// 最近操作日志
	RecentLogs []UserActivityLogSummary `json:"recent_logs"`
}

// SessionSummary 会话摘要
type SessionSummary struct {
	ID           string    `json:"id" example:"sess_123456"`
	IPAddress    string    `json:"ip_address" example:"192.168.1.100"`
	UserAgent    *string   `json:"user_agent" example:"Mozilla/5.0..."`
	LastActivity time.Time `json:"last_activity" example:"2024-01-22T10:15:00Z"`
	CreatedAt    time.Time `json:"created_at" example:"2024-01-22T08:00:00Z"`
	ExpiresAt    time.Time `json:"expires_at" example:"2024-01-23T08:00:00Z"`
}

// UserActivityLogSummary 用户活动日志摘要
type UserActivityLogSummary struct {
	ID        string    `json:"id" example:"log_123456"`
	Action    string    `json:"action" example:"login"`
	IPAddress string    `json:"ip_address" example:"192.168.1.100"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-22T10:15:00Z"`
	Status    string    `json:"status" example:"success"`
}

// UserManagementLogResponse 用户管理日志响应
type UserManagementLogResponse struct {
	ID           string                  `json:"id" example:"log_123456"`
	AdminID      string                  `json:"admin_id" example:"admin_123"`
	AdminEmail   string                  `json:"admin_email" example:"admin@trusioo.com"`
	TargetUserID string                  `json:"target_user_id" example:"user_123"`
	TargetEmail  string                  `json:"target_email" example:"user@example.com"`
	Action       string                  `json:"action" example:"suspend"`
	Reason       *string                 `json:"reason" example:"违反用户协议"`
	Details      *map[string]interface{} `json:"details,omitempty"`
	IPAddress    string                  `json:"ip_address" example:"192.168.1.100"`
	UserAgent    *string                 `json:"user_agent" example:"Mozilla/5.0..."`
	CreatedAt    time.Time               `json:"created_at" example:"2024-01-22T10:15:00Z"`
}

// OperationResponse 操作响应（通用）
type OperationResponse struct {
	Success   bool        `json:"success" example:"true"`
	Message   string      `json:"message" example:"操作成功"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp" example:"2024-01-22T10:15:00Z"`
}

// ExportResponse 导出响应
type ExportResponse struct {
	FileName    string    `json:"file_name" example:"users_export_20240122.csv"`
	FileSize    int64     `json:"file_size" example:"2048576"`
	RecordCount int64     `json:"record_count" example:"1500"`
	DownloadURL string    `json:"download_url" example:"/api/v1/admin/user-management/downloads/users_export_20240122.csv"`
	ExpiresAt   time.Time `json:"expires_at" example:"2024-01-23T10:15:00Z"`
	GeneratedAt time.Time `json:"generated_at" example:"2024-01-22T10:15:00Z"`
}

// === 转换方法 ===

// ToUserDetailResponse 将用户管理模型转换为详情响应
func (u *UserManagementModel) ToUserDetailResponse(activeSessions int) *UserDetailResponse {
	return &UserDetailResponse{
		ID:              u.ID,
		Email:           u.Email,
		Name:            u.Name,
		Status:          u.Status,
		EmailVerified:   u.EmailVerified,
		EmailVerifiedAt: u.EmailVerifiedAt,
		LastLoginAt:     u.LastLoginAt,
		LoginCount:      u.LoginCount,
		FailedAttempts:  u.FailedAttempts,
		ActiveSessions:  activeSessions,
		SuspendedAt:     u.SuspendedAt,
		SuspendedReason: u.SuspendedReason,
		Notes:           u.Notes,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
		DeletedAt:       u.DeletedAt,
	}
}

// ToUserSummaryResponse 将用户管理模型转换为摘要响应
func (u *UserManagementModel) ToUserSummaryResponse(activeSessions int) *UserSummaryResponse {
	return &UserSummaryResponse{
		ID:             u.ID,
		Email:          u.Email,
		Name:           u.Name,
		Status:         u.Status,
		EmailVerified:  u.EmailVerified,
		LastLoginAt:    u.LastLoginAt,
		LoginCount:     u.LoginCount,
		ActiveSessions: activeSessions,
		CreatedAt:      u.CreatedAt,
	}
}

// ToSearchFilter 将请求转换为搜索过滤器
func (req *GetUsersRequest) ToSearchFilter() *SearchFilter {
	filter := &SearchFilter{}

	if req.Email != "" {
		filter.Email = &req.Email
	}
	if req.Name != "" {
		filter.Name = &req.Name
	}
	if req.Status != "" {
		status := user.UserStatus(req.Status)
		filter.Status = &status
	}
	if req.EmailVerified != nil {
		filter.EmailVerified = req.EmailVerified
	}

	// 解析时间范围
	if req.CreatedFrom != "" {
		if t, err := time.Parse("2006-01-02", req.CreatedFrom); err == nil {
			filter.CreatedFrom = &t
		}
	}
	if req.CreatedTo != "" {
		if t, err := time.Parse("2006-01-02", req.CreatedTo); err == nil {
			// 设置为当天结束时间
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filter.CreatedTo = &t
		}
	}
	if req.LastLoginFrom != "" {
		if t, err := time.Parse("2006-01-02", req.LastLoginFrom); err == nil {
			filter.LastLoginFrom = &t
		}
	}
	if req.LastLoginTo != "" {
		if t, err := time.Parse("2006-01-02", req.LastLoginTo); err == nil {
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filter.LastLoginTo = &t
		}
	}

	return filter
}

// ToPaginationParams 将请求转换为分页参数
func (req *GetUsersRequest) ToPaginationParams() PaginationParams {
	params := DefaultPaginationParams()

	if req.Page > 0 {
		params.Page = req.Page
	}
	if req.PageSize > 0 {
		params.PageSize = req.PageSize
	}
	if req.SortBy != "" {
		params.SortBy = req.SortBy
	}
	if req.SortDir != "" {
		params.SortDir = req.SortDir
	}

	return params
}
