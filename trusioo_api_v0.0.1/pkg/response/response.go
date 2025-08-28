package response

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Response 统一API响应结构
type Response struct {
	Success   bool        `json:"success"`         // 请求是否成功
	Code      int         `json:"code"`            // 业务错误码
	Message   string      `json:"message"`         // 响应消息
	Data      interface{} `json:"data,omitempty"`  // 响应数据
	Error     interface{} `json:"error,omitempty"` // 错误详情
	RequestID string      `json:"request_id"`      // 请求ID（用于调用链追踪）
	Timestamp int64       `json:"timestamp"`       // 响应时间戳
}

// PaginatedResponse 分页响应结构
type PaginatedResponse struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Meta      Pagination  `json:"meta"` // 分页信息
	RequestID string      `json:"request_id"`
	Timestamp int64       `json:"timestamp"`
}

// Pagination 分页信息结构
type Pagination struct {
	Page         int   `json:"page"`          // 当前页码
	Limit        int   `json:"limit"`         // 每页数量
	Total        int64 `json:"total"`         // 总记录数
	TotalPages   int   `json:"total_pages"`   // 总页数
	HasPrevious  bool  `json:"has_previous"`  // 是否有上一页
	HasNext      bool  `json:"has_next"`      // 是否有下一页
	PreviousPage *int  `json:"previous_page"` // 上一页页码
	NextPage     *int  `json:"next_page"`     // 下一页页码
}

// 业务错误码常量
const (
	// 成功状态码
	CodeSuccess = 200

	// 客户端错误码 (400-499)
	CodeBadRequest       = 400 // 请求参数错误
	CodeUnauthorized     = 401 // 未授权
	CodeForbidden        = 403 // 禁止访问
	CodeNotFound         = 404 // 资源不存在
	CodeMethodNotAllowed = 405 // 方法不允许
	CodeConflict         = 409 // 资源冲突
	CodeValidationFailed = 422 // 数据验证失败
	CodeTooManyRequests  = 429 // 请求过多

	// 服务端错误码 (500-599)
	CodeInternalError      = 500 // 内部服务器错误
	CodeServiceUnavailable = 503 // 服务不可用
	CodeTimeout            = 504 // 请求超时

	// 业务错误码 (1000+)
	CodeEmailAlreadyExists = 1001 // 邮箱已存在
	CodeInvalidCredentials = 1002 // 凭证无效
	CodeAccountInactive    = 1003 // 账户未激活
	CodeAccountSuspended   = 1004 // 账户被暂停
	CodeTokenExpired       = 1005 // 令牌已过期
	CodeInvalidToken       = 1006 // 令牌无效
	CodeOperationFailed    = 1007 // 操作失败
)

// 响应消息映射
var codeMessages = map[int]string{
	CodeSuccess:            "操作成功",
	CodeBadRequest:         "请求参数错误",
	CodeUnauthorized:       "未授权访问",
	CodeForbidden:          "禁止访问",
	CodeNotFound:           "资源不存在",
	CodeMethodNotAllowed:   "方法不允许",
	CodeConflict:           "资源冲突",
	CodeValidationFailed:   "数据验证失败",
	CodeTooManyRequests:    "请求过多",
	CodeInternalError:      "内部服务器错误",
	CodeServiceUnavailable: "服务不可用",
	CodeTimeout:            "请求超时",
	CodeEmailAlreadyExists: "邮箱已存在",
	CodeInvalidCredentials: "用户名或密码错误",
	CodeAccountInactive:    "账户未激活",
	CodeAccountSuspended:   "账户被暂停",
	CodeTokenExpired:       "令牌已过期",
	CodeInvalidToken:       "令牌无效",
	CodeOperationFailed:    "操作失败",
}

// GetMessage 获取错误码对应的消息
func GetMessage(code int) string {
	if message, exists := codeMessages[code]; exists {
		return message
	}
	return "未知错误"
}

// getRequestID 从上下文中获取请求ID
func getRequestID(c *gin.Context) string {
	if requestID := c.GetString("X-Request-ID"); requestID != "" {
		return requestID
	}
	return ""
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	response := Response{
		Success:   true,
		Code:      CodeSuccess,
		Message:   GetMessage(CodeSuccess),
		Data:      data,
		RequestID: getRequestID(c),
		Timestamp: time.Now().Unix(),
	}
	c.JSON(http.StatusOK, response)
}

// SuccessWithMessage 带自定义消息的成功响应
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	response := Response{
		Success:   true,
		Code:      CodeSuccess,
		Message:   message,
		Data:      data,
		RequestID: getRequestID(c),
		Timestamp: time.Now().Unix(),
	}
	c.JSON(http.StatusOK, response)
}

// Error 错误响应
func Error(c *gin.Context, httpStatus int, code int, err interface{}) {
	response := Response{
		Success:   false,
		Code:      code,
		Message:   GetMessage(code),
		Error:     err,
		RequestID: getRequestID(c),
		Timestamp: time.Now().Unix(),
	}
	c.JSON(httpStatus, response)
	c.Abort()
}

// ErrorWithMessage 带自定义消息的错误响应
func ErrorWithMessage(c *gin.Context, httpStatus int, code int, message string, err interface{}) {
	response := Response{
		Success:   false,
		Code:      code,
		Message:   message,
		Error:     err,
		RequestID: getRequestID(c),
		Timestamp: time.Now().Unix(),
	}
	c.JSON(httpStatus, response)
	c.Abort()
}

// BadRequest 400错误响应
func BadRequest(c *gin.Context, err interface{}) {
	Error(c, http.StatusBadRequest, CodeBadRequest, err)
}

// Unauthorized 401错误响应
func Unauthorized(c *gin.Context, err interface{}) {
	Error(c, http.StatusUnauthorized, CodeUnauthorized, err)
}

// Forbidden 403错误响应
func Forbidden(c *gin.Context, err interface{}) {
	Error(c, http.StatusForbidden, CodeForbidden, err)
}

// NotFound 404错误响应
func NotFound(c *gin.Context, err interface{}) {
	Error(c, http.StatusNotFound, CodeNotFound, err)
}

// Conflict 409错误响应
func Conflict(c *gin.Context, err interface{}) {
	Error(c, http.StatusConflict, CodeConflict, err)
}

// ValidationFailed 422数据验证失败响应
func ValidationFailed(c *gin.Context, err interface{}) {
	Error(c, http.StatusUnprocessableEntity, CodeValidationFailed, err)
}

// InternalError 500内部错误响应
func InternalError(c *gin.Context, err interface{}) {
	Error(c, http.StatusInternalServerError, CodeInternalError, err)
}

// Paginated 分页响应
func Paginated(c *gin.Context, data interface{}, page, limit int, total int64) {
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	var previousPage, nextPage *int
	if page > 1 {
		prev := page - 1
		previousPage = &prev
	}
	if page < totalPages {
		next := page + 1
		nextPage = &next
	}

	pagination := Pagination{
		Page:         page,
		Limit:        limit,
		Total:        total,
		TotalPages:   totalPages,
		HasPrevious:  page > 1,
		HasNext:      page < totalPages,
		PreviousPage: previousPage,
		NextPage:     nextPage,
	}

	response := PaginatedResponse{
		Success:   true,
		Code:      CodeSuccess,
		Message:   GetMessage(CodeSuccess),
		Data:      data,
		Meta:      pagination,
		RequestID: getRequestID(c),
		Timestamp: time.Now().Unix(),
	}

	c.JSON(http.StatusOK, response)
}

// BusinessError 业务错误响应
func BusinessError(c *gin.Context, code int, err interface{}) {
	var httpStatus int
	switch {
	case code >= 400 && code < 500:
		httpStatus = code
	case code >= 1000:
		httpStatus = http.StatusBadRequest
	default:
		httpStatus = http.StatusInternalServerError
	}
	Error(c, httpStatus, code, err)
}
