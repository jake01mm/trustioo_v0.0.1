package auth

import "errors"

// ========== 通用验证错误 ==========
var (
	// 邮箱和密码相关
	ErrInvalidEmail     = errors.New("invalid email format")
	ErrPasswordTooShort = errors.New("password must be at least 6 characters")
	ErrPasswordRequired = errors.New("password is required")

	// 名称相关
	ErrNameTooShort        = errors.New("name must be at least 2 characters")
	ErrCompanyNameTooShort = errors.New("company name must be at least 2 characters")
	ErrContactNameTooShort = errors.New("contact name must be at least 2 characters")

	// 其他字段
	ErrPhoneRequired = errors.New("phone number is required")

	// 验证码相关
	ErrInvalidVerificationCode = errors.New("invalid verification code")
	ErrVerificationCodeExpired = errors.New("verification code has expired")
	ErrVerificationCodeUsed    = errors.New("verification code has already been used")
	ErrTooManyAttempts         = errors.New("too many verification attempts")
)

// ========== 状态错误 ==========
var (
	ErrInvalidUserStatus  = errors.New("invalid user status")
	ErrInvalidAdminStatus = errors.New("invalid admin status")
)

// ========== 认证和授权错误 ==========
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized access")
	ErrForbidden          = errors.New("forbidden access")
	ErrTokenExpired       = errors.New("token has expired")
	ErrTokenInvalid       = errors.New("invalid token")
)

// ========== 用户相关错误 ==========
var (
	ErrUserNotFound  = errors.New("user not found")
	ErrUserExists    = errors.New("user already exists")
	ErrUserInactive  = errors.New("user account is inactive")
	ErrUserSuspended = errors.New("user account is suspended")
)

// ========== 管理员相关错误 ==========
var (
	ErrAdminNotFound  = errors.New("admin not found")
	ErrAdminExists    = errors.New("admin already exists")
	ErrAdminInactive  = errors.New("admin account is inactive")
	ErrAdminSuspended = errors.New("admin account is suspended")
)

// ========== JWT相关错误 ==========
var (
	ErrJWTGenerationFailed = errors.New("failed to generate JWT token")
	ErrJWTValidationFailed = errors.New("failed to validate JWT token")
	ErrRefreshTokenExpired = errors.New("refresh token has expired")
	ErrRefreshTokenInvalid = errors.New("invalid refresh token")
)

// ========== 数据库相关错误 ==========
var (
	ErrDatabaseConnection = errors.New("database connection failed")
	ErrDatabaseQuery      = errors.New("database query failed")
	ErrDatabaseUpdate     = errors.New("database update failed")
	ErrDatabaseDelete     = errors.New("database delete failed")
)

// ========== 业务逻辑错误 ==========
var (
	ErrPermissionDenied      = errors.New("permission denied")
	ErrResourceNotFound      = errors.New("resource not found")
	ErrResourceAlreadyExists = errors.New("resource already exists")
	ErrInvalidOperation      = errors.New("invalid operation")
	ErrOperationNotAllowed   = errors.New("operation not allowed")
)
