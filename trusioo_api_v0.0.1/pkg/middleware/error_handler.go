// Package middleware 提供通用中间件函数
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"trusioo_api_v0.0.1/pkg/errors"
	"trusioo_api_v0.0.1/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ErrorHandler 全局错误处理中间件
func ErrorHandler(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 捕获panic
		defer func() {
			if err := recover(); err != nil {
				logger.WithFields(logrus.Fields{
					"panic":      err,
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"request_id": c.GetString("X-Request-ID"),
				}).Error("Panic recovered")

				// 返回内部服务器错误
				response.InternalError(c, "服务器内部错误")
				return
			}
		}()

		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			handleError(c, err, logger)
		}
	}
}

// handleError 处理错误
func handleError(c *gin.Context, err error, logger *logrus.Logger) {
	// 记录错误日志
	logError(c, err, logger)

	// 检查是否为自定义应用错误
	if appErr, ok := errors.GetAppError(err); ok {
		handleAppError(c, appErr)
		return
	}

	// 处理其他类型的错误
	response.InternalError(c, "服务器内部错误")
}

// handleAppError 处理自定义应用错误
func handleAppError(c *gin.Context, appErr *errors.AppError) {
	var httpStatus int

	// 根据错误码确定HTTP状态码
	switch appErr.Code {
	case errors.ErrCodeUnauthorized, errors.ErrCodeTokenExpired, errors.ErrCodeTokenInvalid:
		httpStatus = http.StatusUnauthorized
	case errors.ErrCodeForbidden, errors.ErrCodeInsufficientPermission:
		httpStatus = http.StatusForbidden
	case errors.ErrCodeRecordNotFound, errors.ErrCodeUserNotFound:
		httpStatus = http.StatusNotFound
	case errors.ErrCodeDuplicateKey, errors.ErrCodeUserExists:
		httpStatus = http.StatusConflict
	case errors.ErrCodeInvalidParameter, errors.ErrCodeMissingParameter, errors.ErrCodeInvalidFormat:
		httpStatus = http.StatusBadRequest
	case errors.ErrCodeRateLimitExceeded:
		httpStatus = http.StatusTooManyRequests
	case errors.ErrCodeServiceUnavailable:
		httpStatus = http.StatusServiceUnavailable
	case errors.ErrCodeTimeout:
		httpStatus = http.StatusGatewayTimeout
	default:
		httpStatus = http.StatusInternalServerError
	}

	// 构建错误响应
	errorDetails := map[string]interface{}{
		"type": "business_error",
	}

	if appErr.Details != nil {
		errorDetails["details"] = appErr.Details
	}

	if appErr.Context != nil {
		errorDetails["context"] = appErr.Context
	}

	response.ErrorWithMessage(c, httpStatus, appErr.Code, appErr.Message, errorDetails)
}

// logError 记录错误日志
func logError(c *gin.Context, err error, logger *logrus.Logger) {
	fields := logrus.Fields{
		"error":      err.Error(),
		"path":       c.Request.URL.Path,
		"method":     c.Request.Method,
		"request_id": c.GetString("X-Request-ID"),
		"user_agent": c.Request.UserAgent(),
		"client_ip":  c.ClientIP(),
	}

	// 如果是自定义错误，添加更多信息
	if appErr, ok := errors.GetAppError(err); ok {
		fields["error_code"] = appErr.Code
		fields["error_type"] = "app_error"

		if appErr.Context != nil {
			fields["error_context"] = appErr.Context
		}

		// 根据错误级别决定日志级别
		if appErr.Code >= 5000 {
			logger.WithFields(fields).Error("Business error occurred")
		} else {
			logger.WithFields(fields).Warn("Application error occurred")
		}
	} else {
		logger.WithFields(fields).Error("Unhandled error occurred")
	}
}

// === 以下是从 internal/middleware 移动过来的通用中间件 ===

const RequestIDHeader = "X-Request-ID"

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从请求头中获取请求ID
		requestID := c.GetHeader(RequestIDHeader)

		// 如果没有提供请求ID，则生成一个新的
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置请求ID到上下文
		c.Set("request_id", requestID)

		// 设置响应头
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID 从上下文中获取请求ID
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		return requestID.(string)
	}
	return ""
}

// Logger 日志中间件
func Logger(logger *logrus.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 获取请求ID
		requestID := ""
		if param.Request != nil {
			requestID = param.Request.Header.Get(RequestIDHeader)
		}

		// 自定义日志格式
		logger.WithFields(logrus.Fields{
			"request_id":  requestID,
			"status_code": param.StatusCode,
			"latency":     param.Latency,
			"client_ip":   param.ClientIP,
			"method":      param.Method,
			"path":        param.Path,
			"user_agent":  param.Request.UserAgent(),
			"error":       param.ErrorMessage,
			"body_size":   param.BodySize,
			"timestamp":   param.TimeStamp.Format(time.RFC3339),
		}).Info("HTTP Request")

		return ""
	})
}

// Recovery 恢复中间件，替代默认的Recovery
func Recovery(logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 获取堆栈信息
				stack := debug.Stack()
				requestID := GetRequestID(c)

				// 记录panic日志
				logger.WithFields(logrus.Fields{
					"request_id": requestID,
					"panic":      err,
					"stack":      string(stack),
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"client_ip":  c.ClientIP(),
				}).Error("Panic recovered")

				// 返回内部服务器错误
				if !c.Writer.Written() {
					response.InternalError(c, "服务器内部错误")
				}

				c.Abort()
			}
		}()

		c.Next()
	}
}

// Timeout 超时中间件
func Timeout(timeout time.Duration, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建带超时的上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 更新请求的上下文
		c.Request = c.Request.WithContext(ctx)

		// 监听取消信号
		finished := make(chan struct{}, 1)
		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			// 请求正常完成
			return
		case <-ctx.Done():
			// 请求超时
			requestID := GetRequestID(c)
			logger.WithFields(logrus.Fields{
				"request_id": requestID,
				"timeout":    timeout,
				"path":       c.Request.URL.Path,
				"method":     c.Request.Method,
			}).Warn("Request timeout")

			response.Error(c, http.StatusGatewayTimeout, response.CodeTimeout, fmt.Errorf("request timeout after %v", timeout))
			c.Abort()
			return
		}
	}
}
