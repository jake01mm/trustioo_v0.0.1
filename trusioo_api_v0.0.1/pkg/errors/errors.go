package errors

import (
	"fmt"
	"runtime"
)

// AppError 自定义应用错误类型
type AppError struct {
	Code    int                    `json:"code"`              // 业务错误码
	Message string                 `json:"message"`           // 错误消息
	Details any            `json:"details,omitempty"` // 错误详情
	Cause   error                  `json:"-"`                 // 原始错误
	Stack   string                 `json:"stack,omitempty"`   // 错误堆栈（开发环境）
	Context map[string]any `json:"context,omitempty"` // 错误上下文
}

// Error 实现error接口
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("Code: %d, Message: %s, Cause: %s", e.Code, e.Message, e.Cause.Error())
	}
	return fmt.Sprintf("Code: %d, Message: %s", e.Code, e.Message)
}

// Unwrap 实现错误链
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithContext 添加错误上下文
func (e *AppError) WithContext(key string, value any) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// WithDetails 添加错误详情
func (e *AppError) WithDetails(details any) *AppError {
	e.Details = details
	return e
}

// 错误码定义
const (
	// 通用错误码
	ErrCodeUnknown          = 1000
	ErrCodeInvalidParameter = 1001
	ErrCodeMissingParameter = 1002
	ErrCodeInvalidFormat    = 1003

	// 认证错误码
	ErrCodeUnauthorized       = 2001
	ErrCodeForbidden          = 2002
	ErrCodeTokenExpired       = 2003
	ErrCodeTokenInvalid       = 2004
	ErrCodeInvalidCredentials = 2005

	// 用户相关错误码
	ErrCodeUserNotFound  = 3001
	ErrCodeUserExists    = 3002
	ErrCodeUserInactive  = 3003
	ErrCodeUserSuspended = 3004

	// 数据库错误码
	ErrCodeDatabaseError       = 4001
	ErrCodeRecordNotFound      = 4002
	ErrCodeDuplicateKey        = 4003
	ErrCodeConstraintViolation = 4004

	// 业务逻辑错误码
	ErrCodeOperationFailed        = 5001
	ErrCodeInsufficientPermission = 5002
	ErrCodeResourceLocked         = 5003
	ErrCodeQuotaExceeded          = 5004

	// 外部服务错误码
	ErrCodeServiceUnavailable = 6001
	ErrCodeTimeout            = 6002
	ErrCodeRateLimitExceeded  = 6003
)

// 错误消息映射
var errorMessages = map[int]string{
	// 通用错误
	ErrCodeUnknown:          "未知错误",
	ErrCodeInvalidParameter: "参数无效",
	ErrCodeMissingParameter: "缺少必需参数",
	ErrCodeInvalidFormat:    "格式错误",

	// 认证错误
	ErrCodeUnauthorized:       "未授权访问",
	ErrCodeForbidden:          "禁止访问",
	ErrCodeTokenExpired:       "令牌已过期",
	ErrCodeTokenInvalid:       "令牌无效",
	ErrCodeInvalidCredentials: "用户名或密码错误",

	// 用户相关错误
	ErrCodeUserNotFound:  "用户不存在",
	ErrCodeUserExists:    "用户已存在",
	ErrCodeUserInactive:  "用户账户未激活",
	ErrCodeUserSuspended: "用户账户已被暂停",

	// 数据库错误
	ErrCodeDatabaseError:       "数据库错误",
	ErrCodeRecordNotFound:      "记录不存在",
	ErrCodeDuplicateKey:        "记录已存在",
	ErrCodeConstraintViolation: "数据约束违反",

	// 业务逻辑错误
	ErrCodeOperationFailed:        "操作失败",
	ErrCodeInsufficientPermission: "权限不足",
	ErrCodeResourceLocked:         "资源被锁定",
	ErrCodeQuotaExceeded:          "配额已超限",

	// 外部服务错误
	ErrCodeServiceUnavailable: "服务不可用",
	ErrCodeTimeout:            "请求超时",
	ErrCodeRateLimitExceeded:  "请求频率超限",
}

// GetErrorMessage 获取错误消息
func GetErrorMessage(code int) string {
	if message, exists := errorMessages[code]; exists {
		return message
	}
	return "未知错误"
}

// getStack 获取调用堆栈
func getStack() string {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// New 创建新的应用错误
func New(code int, message string) *AppError {
	if message == "" {
		message = GetErrorMessage(code)
	}

	return &AppError{
		Code:    code,
		Message: message,
		Stack:   getStack(),
	}
}

// Wrap 包装现有错误
func Wrap(err error, code int, message string) *AppError {
	if message == "" {
		message = GetErrorMessage(code)
	}

	return &AppError{
		Code:    code,
		Message: message,
		Cause:   err,
		Stack:   getStack(),
	}
}

// WrapWithContext 包装错误并添加上下文
func WrapWithContext(err error, code int, message string, context map[string]any) *AppError {
	appErr := Wrap(err, code, message)
	appErr.Context = context
	return appErr
}

// 预定义的常用错误

// NewUnknownError 未知错误
func NewUnknownError(details any) *AppError {
	return New(ErrCodeUnknown, "").WithDetails(details)
}

// NewValidationError 验证错误
func NewValidationError(details any) *AppError {
	return New(ErrCodeInvalidParameter, "").WithDetails(details)
}

// NewUnauthorizedError 未授权错误
func NewUnauthorizedError() *AppError {
	return New(ErrCodeUnauthorized, "")
}

// NewForbiddenError 禁止访问错误
func NewForbiddenError() *AppError {
	return New(ErrCodeForbidden, "")
}

// NewNotFoundError 资源不存在错误
func NewNotFoundError(resource string) *AppError {
	return New(ErrCodeRecordNotFound, fmt.Sprintf("%s不存在", resource))
}

// NewDuplicateError 重复资源错误
func NewDuplicateError(resource string) *AppError {
	return New(ErrCodeDuplicateKey, fmt.Sprintf("%s已存在", resource))
}

// NewDatabaseError 数据库错误
func NewDatabaseError(err error) *AppError {
	return Wrap(err, ErrCodeDatabaseError, "")
}

// NewBusinessError 业务逻辑错误
func NewBusinessError(code int, message string) *AppError {
	return New(code, message)
}

// 错误类型检查函数

// IsAppError 检查是否为应用错误
func IsAppError(err error) bool {
	_, ok := err.(*AppError)
	return ok
}

// GetAppError 获取应用错误
func GetAppError(err error) (*AppError, bool) {
	appErr, ok := err.(*AppError)
	return appErr, ok
}

// IsErrorCode 检查错误码
func IsErrorCode(err error, code int) bool {
	if appErr, ok := GetAppError(err); ok {
		return appErr.Code == code
	}
	return false
}
