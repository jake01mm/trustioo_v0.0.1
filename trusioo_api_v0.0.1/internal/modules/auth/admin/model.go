package admin

import (
	"time"
)

// Admin 管理员实体模型
// 核心领域对象，定义管理员的基本属性和行为
type Admin struct {
	// 基础字段
	ID       string `json:"id" db:"id"`         // 管理员唯一标识
	Email    string `json:"email" db:"email"`   // 邮箱地址
	Name     string `json:"name" db:"name"`     // 管理员姓名
	Password string `json:"-" db:"password"`    // 密码（不在JSON中显示）
	Role     string `json:"role" db:"role"`     // 角色：super_admin, admin
	Active   bool   `json:"active" db:"active"` // 是否激活

	// 审计字段
	CreatedAt time.Time  `json:"created_at" db:"created_at"` // 创建时间
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"` // 更新时间
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"` // 软删除时间
}

// TableName 返回数据库表名
func (Admin) TableName() string {
	return "admins"
}

// IsActive 检查管理员是否处于激活状态
func (a *Admin) IsActive() bool {
	return a.Active && a.DeletedAt == nil
}

// IsSuperAdmin 检查是否为超级管理员
func (a *Admin) IsSuperAdmin() bool {
	return a.Role == "super_admin"
}

// GetPublicInfo 获取可公开的管理员信息（不包含敏感数据）
func (a *Admin) GetPublicInfo() *AdminPublicInfo {
	return &AdminPublicInfo{
		ID:        a.ID,
		Email:     a.Email,
		Name:      a.Name,
		Role:      a.Role,
		Active:    a.Active,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
	}
}

// AdminPublicInfo 管理员公开信息结构
type AdminPublicInfo struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AdminRole 管理员角色常量
type AdminRole string

const (
	AdminRoleSuperAdmin AdminRole = "super_admin" // 超级管理员
	AdminRoleAdmin      AdminRole = "admin"       // 普通管理员
)

// IsValidRole 验证角色是否有效
func (r AdminRole) IsValid() bool {
	switch r {
	case AdminRoleSuperAdmin, AdminRoleAdmin:
		return true
	default:
		return false
	}
}

// GetRoles 获取所有有效的管理员角色
func GetValidAdminRoles() []AdminRole {
	return []AdminRole{
		AdminRoleSuperAdmin,
		AdminRoleAdmin,
	}
}

// LoginLog 管理员登录日志模型
type LoginLog struct {
	ID            string                  `json:"id" db:"id"`
	AdminID       *string                 `json:"admin_id" db:"admin_id"`
	Email         string                  `json:"email" db:"email"`
	UserType      string                  `json:"user_type" db:"user_type"`       // 固定为"admin"
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
	UserType  string     `json:"user_type" db:"user_type"` // 固定为"admin"
	Token     string     `json:"token" db:"token"`
	IPAddress *string    `json:"ip_address" db:"ip_address"`
	Used      bool       `json:"used" db:"used"`
	UsedAt    *time.Time `json:"used_at" db:"used_at"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}
