package user_management

import (
	"time"

	"trusioo_api_v0.0.1/internal/modules/auth/user"
)

// UserManagementModel 用户管理模型 - 扩展基础用户模型的管理功能
type UserManagementModel struct {
	*user.User
	// 管理相关的额外字段
	LastLoginAt     *time.Time `json:"last_login_at" db:"last_login_at"`
	LoginCount      int64      `json:"login_count" db:"login_count"`
	FailedAttempts  int        `json:"failed_attempts" db:"failed_attempts"`
	SuspendedAt     *time.Time `json:"suspended_at" db:"suspended_at"`
	SuspendedReason *string    `json:"suspended_reason" db:"suspended_reason"`
	Notes           *string    `json:"notes" db:"notes"` // 管理员备注
}

// UserStatistics 用户统计数据结构
type UserStatistics struct {
	// 总体统计
	TotalUsers        int64 `json:"total_users"`
	NewUsersToday     int64 `json:"new_users_today"`
	NewUsersThisWeek  int64 `json:"new_users_this_week"`
	NewUsersThisMonth int64 `json:"new_users_this_month"`

	// 验证状态统计
	VerifiedUsers    int64   `json:"verified_users"`
	UnverifiedUsers  int64   `json:"unverified_users"`
	VerificationRate float64 `json:"verification_rate"` // 验证率百分比

	// 用户状态统计
	ActiveUsers    int64                     `json:"active_users"`
	InactiveUsers  int64                     `json:"inactive_users"`
	SuspendedUsers int64                     `json:"suspended_users"`
	DeletedUsers   int64                     `json:"deleted_users"`
	UsersByStatus  map[user.UserStatus]int64 `json:"users_by_status"`

	// 时间趋势统计
	RegistrationTrend []DailyCount `json:"registration_trend"`
	VerificationTrend []DailyCount `json:"verification_trend"`

	// 活跃度统计
	ActiveUsersToday     int64 `json:"active_users_today"`
	ActiveUsersThisWeek  int64 `json:"active_users_this_week"`
	ActiveUsersThisMonth int64 `json:"active_users_this_month"`

	// 生成时间
	GeneratedAt time.Time `json:"generated_at"`
}

// DailyCount 每日计数结构
type DailyCount struct {
	Date  string `json:"date"` // YYYY-MM-DD 格式
	Count int64  `json:"count"`
}

// UserActivitySummary 用户活动摘要
type UserActivitySummary struct {
	UserID              string     `json:"user_id"`
	Email               string     `json:"email"`
	Name                string     `json:"name"`
	LastLoginAt         *time.Time `json:"last_login_at"`
	TotalLogins         int64      `json:"total_logins"`
	FailedLoginAttempts int        `json:"failed_login_attempts"`
	ActiveSessions      int        `json:"active_sessions"`
	LastActivityAt      *time.Time `json:"last_activity_at"`
}

// UserManagementAction 用户管理操作类型
type UserManagementAction string

const (
	ActionActivate      UserManagementAction = "activate"
	ActionDeactivate    UserManagementAction = "deactivate"
	ActionSuspend       UserManagementAction = "suspend"
	ActionUnsuspend     UserManagementAction = "unsuspend"
	ActionDelete        UserManagementAction = "delete"
	ActionResetPassword UserManagementAction = "reset_password"
	ActionForceLogout   UserManagementAction = "force_logout"
	ActionUpdateEmail   UserManagementAction = "update_email"
	ActionVerifyEmail   UserManagementAction = "verify_email"
)

// IsValid 验证操作类型是否有效
func (a UserManagementAction) IsValid() bool {
	switch a {
	case ActionActivate, ActionDeactivate, ActionSuspend, ActionUnsuspend,
		ActionDelete, ActionResetPassword, ActionForceLogout, ActionUpdateEmail, ActionVerifyEmail:
		return true
	default:
		return false
	}
}

// String 返回操作类型字符串
func (a UserManagementAction) String() string {
	return string(a)
}

// UserManagementLog 用户管理操作日志
type UserManagementLog struct {
	ID           string                  `json:"id" db:"id"`
	AdminID      string                  `json:"admin_id" db:"admin_id"`
	AdminEmail   string                  `json:"admin_email" db:"admin_email"`
	TargetUserID string                  `json:"target_user_id" db:"target_user_id"`
	TargetEmail  string                  `json:"target_email" db:"target_email"`
	Action       UserManagementAction    `json:"action" db:"action"`
	Reason       *string                 `json:"reason" db:"reason"`
	Details      *map[string]interface{} `json:"details" db:"details"` // JSON格式的详细信息
	IPAddress    string                  `json:"ip_address" db:"ip_address"`
	UserAgent    *string                 `json:"user_agent" db:"user_agent"`
	CreatedAt    time.Time               `json:"created_at" db:"created_at"`
}

// SearchFilter 用户搜索过滤器
type SearchFilter struct {
	Email         *string          `json:"email"`
	Name          *string          `json:"name"`
	Status        *user.UserStatus `json:"status"`
	EmailVerified *bool            `json:"email_verified"`
	CreatedFrom   *time.Time       `json:"created_from"`
	CreatedTo     *time.Time       `json:"created_to"`
	LastLoginFrom *time.Time       `json:"last_login_from"`
	LastLoginTo   *time.Time       `json:"last_login_to"`
}

// PaginationParams 分页参数
type PaginationParams struct {
	Page     int    `json:"page" binding:"min=1"`
	PageSize int    `json:"page_size" binding:"min=1,max=100"`
	SortBy   string `json:"sort_by"`
	SortDir  string `json:"sort_dir" binding:"oneof=asc desc"`
}

// DefaultPaginationParams 默认分页参数
func DefaultPaginationParams() PaginationParams {
	return PaginationParams{
		Page:     1,
		PageSize: 20,
		SortBy:   "created_at",
		SortDir:  "desc",
	}
}

// GetOffset 计算偏移量
func (p *PaginationParams) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}

// GetLimit 获取限制数量
func (p *PaginationParams) GetLimit() int {
	return p.PageSize
}

// PaginatedResult 分页结果
type PaginatedResult struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
	HasNext    bool        `json:"has_next"`
	HasPrev    bool        `json:"has_prev"`
}

// NewPaginatedResult 创建分页结果
func NewPaginatedResult(data interface{}, total int64, params PaginationParams) *PaginatedResult {
	totalPages := int((total + int64(params.PageSize) - 1) / int64(params.PageSize))

	return &PaginatedResult{
		Data:       data,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
		HasNext:    params.Page < totalPages,
		HasPrev:    params.Page > 1,
	}
}

// UserExportData 用户导出数据结构
type UserExportData struct {
	ID              string  `json:"id" csv:"用户ID"`
	Email           string  `json:"email" csv:"邮箱"`
	Name            string  `json:"name" csv:"姓名"`
	Status          string  `json:"status" csv:"状态"`
	EmailVerified   string  `json:"email_verified" csv:"邮箱已验证"`
	EmailVerifiedAt *string `json:"email_verified_at" csv:"邮箱验证时间"`
	LastLoginAt     *string `json:"last_login_at" csv:"最后登录时间"`
	LoginCount      int64   `json:"login_count" csv:"登录次数"`
	CreatedAt       string  `json:"created_at" csv:"注册时间"`
	UpdatedAt       string  `json:"updated_at" csv:"更新时间"`
}

// ToExportData 将用户数据转换为导出格式
func (u *UserManagementModel) ToExportData() *UserExportData {
	data := &UserExportData{
		ID:            u.ID,
		Email:         u.Email,
		Name:          u.Name,
		Status:        u.Status,
		EmailVerified: boolToString(u.EmailVerified),
		LoginCount:    u.LoginCount,
		CreatedAt:     u.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     u.UpdatedAt.Format("2006-01-02 15:04:05"),
	}

	if u.EmailVerifiedAt != nil {
		verifiedAt := u.EmailVerifiedAt.Format("2006-01-02 15:04:05")
		data.EmailVerifiedAt = &verifiedAt
	}

	if u.LastLoginAt != nil {
		lastLogin := u.LastLoginAt.Format("2006-01-02 15:04:05")
		data.LastLoginAt = &lastLogin
	}

	return data
}

// boolToString 将布尔值转换为字符串
func boolToString(b bool) string {
	if b {
		return "是"
	}
	return "否"
}
