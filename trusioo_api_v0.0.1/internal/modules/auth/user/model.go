package user

import (
	"time"

	"trusioo_api_v0.0.1/internal/modules/auth"
)

// User 用户实体模型
// 核心领域对象，定义用户的基本属性和行为
type User struct {
	// 基础字段
	ID              string     `json:"id" db:"id"`                               // 用户唯一标识
	Email           string     `json:"email" db:"email"`                         // 邮箱地址
	Name            string     `json:"name" db:"name"`                           // 用户姓名
	Password        string     `json:"-" db:"password"`                          // 密码（不在JSON中显示）
	Status          string     `json:"status" db:"status"`                       // 状态：active, inactive, suspended
	EmailVerified   bool       `json:"email_verified" db:"email_verified"`       // 邮箱是否已验证
	EmailVerifiedAt *time.Time `json:"email_verified_at" db:"email_verified_at"` // 邮箱验证时间

	// 审计字段
	CreatedAt time.Time  `json:"created_at" db:"created_at"` // 创建时间
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"` // 更新时间
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"` // 软删除时间
}

// UserSession 用户会话模型（简化版）
type UserSession struct {
	ID             string                  `json:"id" db:"id"`
	SessionID      string                  `json:"session_id" db:"session_id"`
	UserID         string                  `json:"user_id" db:"user_id"`
	UserType       string                  `json:"user_type" db:"user_type"` // 固定为"user"
	RefreshTokenID *string                 `json:"refresh_token_id" db:"refresh_token_id"`
	IPAddress      string                  `json:"ip_address" db:"ip_address"`
	UserAgent      *string                 `json:"user_agent" db:"user_agent"`
	DeviceInfo     *map[string]interface{} `json:"device_info" db:"device_info"`
	LocationInfo   *map[string]interface{} `json:"location_info" db:"location_info"`
	IsActive       bool                    `json:"is_active" db:"is_active"`
	LastActivity   time.Time               `json:"last_activity" db:"last_activity"`
	ExpiresAt      time.Time               `json:"expires_at" db:"expires_at"`
	CreatedAt      time.Time               `json:"created_at" db:"created_at"`
}

// TableName 返回数据库表名
func (User) TableName() string {
	return "users"
}

// IsActive 检查用户是否处于激活状态
func (u *User) IsActive() bool {
	return u.Status == string(UserStatusActive) && u.DeletedAt == nil
}

// CanLogin 检查用户是否可以登录
func (u *User) CanLogin() bool {
	return u.IsActive() && u.Status != string(UserStatusSuspended)
}

// GetPublicInfo 获取可公开的用户信息（不包含敏感数据）
func (u *User) GetPublicInfo() *UserPublicInfo {
	return &UserPublicInfo{
		ID:              u.ID,
		Email:           u.Email,
		Name:            u.Name,
		Status:          u.Status,
		EmailVerified:   u.EmailVerified,
		EmailVerifiedAt: u.EmailVerifiedAt,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
}

// UserPublicInfo 用户公开信息结构
type UserPublicInfo struct {
	ID              string     `json:"id"`
	Email           string     `json:"email"`
	Name            string     `json:"name"`
	Status          string     `json:"status"`
	EmailVerified   bool       `json:"email_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// UserStatus 用户状态常量
type UserStatus string

const (
	UserStatusActive    UserStatus = "active"    // 激活状态
	UserStatusInactive  UserStatus = "inactive"  // 未激活状态
	UserStatusSuspended UserStatus = "suspended" // 暂停状态
)

// IsValid 验证状态是否有效
func (s UserStatus) IsValid() bool {
	switch s {
	case UserStatusActive, UserStatusInactive, UserStatusSuspended:
		return true
	default:
		return false
	}
}

// String 返回状态字符串
func (s UserStatus) String() string {
	return string(s)
}

// GetValidUserStatuses 获取所有有效的用户状态
func GetValidUserStatuses() []UserStatus {
	return []UserStatus{
		UserStatusActive,
		UserStatusInactive,
		UserStatusSuspended,
	}
}

// SetStatus 设置用户状态
func (u *User) SetStatus(status UserStatus) error {
	if !status.IsValid() {
		return auth.ErrInvalidUserStatus
	}
	u.Status = status.String()
	return nil
}

// Activate 激活用户
func (u *User) Activate() {
	u.Status = UserStatusActive.String()
}

// Deactivate 停用用户
func (u *User) Deactivate() {
	u.Status = UserStatusInactive.String()
}

// Suspend 暂停用户
func (u *User) Suspend() {
	u.Status = UserStatusSuspended.String()
}

// VerifyEmail 验证邮箱
func (u *User) VerifyEmail() {
	u.EmailVerified = true
	now := time.Now()
	u.EmailVerifiedAt = &now
}

// IsEmailVerified 检查邮箱是否已验证
func (u *User) IsEmailVerified() bool {
	return u.EmailVerified
}

// UserSession 相关常量
const (
	// 默认会话过期时间
	DefaultSessionDuration  = 24 * time.Hour     // 24小时
	ExtendedSessionDuration = 7 * 24 * time.Hour // 7天（记住我）
)

// IsValid 检查会话是否有效
func (s *UserSession) IsValid() bool {
	return s.IsActive && !s.IsExpired()
}

// IsExpired 检查会话是否已过期
func (s *UserSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Deactivate 停用会话
func (s *UserSession) Deactivate() {
	s.IsActive = false
	s.LastActivity = time.Now()
}

// GetDefaultExpirationTime 获取默认过期时间
func GetDefaultExpirationTime(rememberMe bool) time.Time {
	if rememberMe {
		return time.Now().Add(ExtendedSessionDuration)
	}
	return time.Now().Add(DefaultSessionDuration)
}

// LoginLog 用户登录日志模型（简化版）
type LoginLog struct {
	ID            string                  `json:"id" db:"id"`
	UserID        *string                 `json:"user_id" db:"user_id"`
	Email         string                  `json:"email" db:"email"`
	UserType      string                  `json:"user_type" db:"user_type"`       // 固定为"user"
	LoginStatus   string                  `json:"login_status" db:"login_status"` // success, failed, blocked
	FailureReason *string                 `json:"failure_reason" db:"failure_reason"`
	IPAddress     string                  `json:"ip_address" db:"ip_address"`
	UserAgent     *string                 `json:"user_agent" db:"user_agent"`
	DeviceInfo    *map[string]interface{} `json:"device_info" db:"device_info"`
	LocationInfo  *map[string]interface{} `json:"location_info" db:"location_info"`
	SessionID     *string                 `json:"session_id" db:"session_id"`
	RiskScore     int                     `json:"risk_score" db:"risk_score"`
	CreatedAt     time.Time               `json:"created_at" db:"created_at"`
}

// LoginLog 相关常量
const (
	LoginStatusSuccess = "success"
	LoginStatusFailed  = "failed"
	LoginStatusBlocked = "blocked"

	FailureReasonInvalidCredentials = "invalid_credentials"
	FailureReasonInvalidCode        = "invalid_verification_code"
	FailureReasonTooManyAttempts    = "too_many_attempts"
	FailureReasonAccountLocked      = "account_locked"
	FailureReasonAccountInactive    = "account_inactive"
)

// PasswordReset 密码重置记录模型
type PasswordReset struct {
	ID        string     `json:"id" db:"id"`
	Email     string     `json:"email" db:"email"`
	UserType  string     `json:"user_type" db:"user_type"` // 固定为"user"
	Token     string     `json:"token" db:"token"`
	IPAddress *string    `json:"ip_address" db:"ip_address"`
	Used      bool       `json:"used" db:"used"`
	UsedAt    *time.Time `json:"used_at" db:"used_at"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}
