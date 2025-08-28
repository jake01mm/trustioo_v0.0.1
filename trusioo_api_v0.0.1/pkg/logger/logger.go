package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// Logger 增强日志器
type Logger struct {
	*logrus.Logger
	config *Config
}

// Config 日志配置
type Config struct {
	Level      string `json:"level"`       // 日志级别
	Format     string `json:"format"`      // 日志格式 (json/text)
	Output     string `json:"output"`      // 输出方式 (stdout/file)
	FilePath   string `json:"file_path"`   // 文件路径
	MaxSize    int    `json:"max_size"`    // 最大文件大小(MB)
	MaxBackups int    `json:"max_backups"` // 最大备份数
	MaxAge     int    `json:"max_age"`     // 最大保存天数
	Compress   bool   `json:"compress"`    // 是否压缩
}

// Fields 日志字段类型
type Fields map[string]interface{}

// TraceContext 调用链上下文
type TraceContext struct {
	RequestID string `json:"request_id"` // 请求ID
	TraceID   string `json:"trace_id"`   // 调用链ID
	SpanID    string `json:"span_id"`    // 跨度ID
	UserID    string `json:"user_id"`    // 用户ID
}

// RequestLogFields 请求日志字段
type RequestLogFields struct {
	Method     string        `json:"method"`      // HTTP方法
	URL        string        `json:"url"`         // 请求URL
	Proto      string        `json:"proto"`       // 协议版本
	StatusCode int           `json:"status_code"` // 状态码
	Latency    time.Duration `json:"latency"`     // 响应时间
	ClientIP   string        `json:"client_ip"`   // 客户端IP
	UserAgent  string        `json:"user_agent"`  // 用户代理
	BodySize   int           `json:"body_size"`   // 响应体大小
}

// NewLogger 创建新的日志器
func NewLogger(config *Config) *Logger {
	logger := logrus.New()

	// 设置日志级别
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// 设置日志格式
	if config.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	// 设置输出
	if config.Output == "file" && config.FilePath != "" {
		// 这里可以集成文件轮转，暂时简化为文件输出
		file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			logger.SetOutput(os.Stdout)
		} else {
			logger.SetOutput(file)
		}
	} else {
		logger.SetOutput(os.Stdout)
	}

	return &Logger{
		Logger: logger,
		config: config,
	}
}

// WithContext 添加调用链上下文
func (l *Logger) WithContext(ctx context.Context) *logrus.Entry {
	fields := logrus.Fields{}

	if traceCtx := GetTraceContext(ctx); traceCtx != nil {
		if traceCtx.RequestID != "" {
			fields["request_id"] = traceCtx.RequestID
		}
		if traceCtx.TraceID != "" {
			fields["trace_id"] = traceCtx.TraceID
		}
		if traceCtx.SpanID != "" {
			fields["span_id"] = traceCtx.SpanID
		}
		if traceCtx.UserID != "" {
			fields["user_id"] = traceCtx.UserID
		}
	}

	return l.WithFields(fields)
}

// WithRequestContext 添加请求上下文
func (l *Logger) WithRequestContext(c *gin.Context) *logrus.Entry {
	fields := logrus.Fields{
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"client_ip":  c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
	}

	// 添加请求ID
	if requestID := c.GetString("X-Request-ID"); requestID != "" {
		fields["request_id"] = requestID
	}

	// 添加用户ID（如果已认证）
	if userID := c.GetString("user_id"); userID != "" {
		fields["user_id"] = userID
	}

	return l.WithFields(fields)
}

// Performance 性能日志
func (l *Logger) Performance(operation string, duration time.Duration, fields Fields) {
	logFields := logrus.Fields{
		"operation":   operation,
		"duration":    duration.String(),
		"duration_ms": duration.Milliseconds(),
		"type":        "performance",
	}

	for k, v := range fields {
		logFields[k] = v
	}

	l.WithFields(logFields).Info("Performance metric")
}

// Business 业务日志
func (l *Logger) Business(action string, result string, fields Fields) {
	logFields := logrus.Fields{
		"action": action,
		"result": result,
		"type":   "business",
	}

	for k, v := range fields {
		logFields[k] = v
	}

	l.WithFields(logFields).Info("Business action")
}

// Security 安全日志
func (l *Logger) Security(event string, level string, fields Fields) {
	logFields := logrus.Fields{
		"event":          event,
		"security_level": level,
		"type":           "security",
	}

	for k, v := range fields {
		logFields[k] = v
	}

	entry := l.WithFields(logFields)
	switch level {
	case "critical":
		entry.Error("Security event")
	case "warning":
		entry.Warn("Security event")
	default:
		entry.Info("Security event")
	}
}

// Database 数据库日志
func (l *Logger) Database(query string, duration time.Duration, error error, fields Fields) {
	logFields := logrus.Fields{
		"query":       query,
		"duration":    duration.String(),
		"duration_ms": duration.Milliseconds(),
		"type":        "database",
	}

	for k, v := range fields {
		logFields[k] = v
	}

	entry := l.WithFields(logFields)
	if error != nil {
		logFields["error"] = error.Error()
		entry.Error("Database query failed")
	} else if duration > time.Second {
		entry.Warn("Slow database query")
	} else {
		entry.Debug("Database query executed")
	}
}

// 调用链追踪相关

// traceContextKey 调用链上下文键
type traceContextKey struct{}

// SetTraceContext 设置调用链上下文
func SetTraceContext(ctx context.Context, traceCtx *TraceContext) context.Context {
	return context.WithValue(ctx, traceContextKey{}, traceCtx)
}

// GetTraceContext 获取调用链上下文
func GetTraceContext(ctx context.Context) *TraceContext {
	if traceCtx, ok := ctx.Value(traceContextKey{}).(*TraceContext); ok {
		return traceCtx
	}
	return nil
}

// NewTraceContext 创建新的调用链上下文
func NewTraceContext() *TraceContext {
	return &TraceContext{
		RequestID: uuid.New().String(),
		TraceID:   uuid.New().String(),
		SpanID:    uuid.New().String(),
	}
}

// 中间件

// RequestID 请求ID中间件
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		c.Set("X-Request-ID", requestID)
		c.Header("X-Request-ID", requestID)

		// 设置调用链上下文
		traceCtx := NewTraceContext()
		traceCtx.RequestID = requestID

		ctx := SetTraceContext(c.Request.Context(), traceCtx)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequestLogger 请求日志中间件
func RequestLoggerMiddleware(logger *Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		fields := RequestLogFields{
			Method:     c.Request.Method,
			URL:        path,
			Proto:      c.Request.Proto,
			StatusCode: c.Writer.Status(),
			Latency:    latency,
			ClientIP:   c.ClientIP(),
			UserAgent:  c.Request.UserAgent(),
			BodySize:   c.Writer.Size(),
		}

		if raw != "" {
			fields.URL = path + "?" + raw
		}

		logFields := logrus.Fields{
			"method":      fields.Method,
			"url":         fields.URL,
			"status_code": fields.StatusCode,
			"latency":     fields.Latency.String(),
			"latency_ms":  fields.Latency.Milliseconds(),
			"client_ip":   fields.ClientIP,
			"user_agent":  fields.UserAgent,
			"body_size":   fields.BodySize,
			"type":        "request",
		}

		// 添加请求ID
		if requestID := c.GetString("X-Request-ID"); requestID != "" {
			logFields["request_id"] = requestID
		}

		// 添加错误信息
		if len(c.Errors) > 0 {
			logFields["errors"] = c.Errors.String()
		}

		entry := logger.WithFields(logFields)

		// 根据状态码确定日志级别
		switch {
		case fields.StatusCode >= 500:
			entry.Error("Request completed with server error")
		case fields.StatusCode >= 400:
			entry.Warn("Request completed with client error")
		case fields.Latency > 5*time.Second:
			entry.Warn("Request completed with high latency")
		default:
			entry.Info("Request completed")
		}
	}
}

// 全局日志器
var defaultLogger *Logger

// Init 初始化全局日志器
func Init(config *Config) {
	defaultLogger = NewLogger(config)
}

// GetLogger 获取全局日志器
func GetLogger() *Logger {
	if defaultLogger == nil {
		// 使用默认配置
		Init(&Config{
			Level:  "info",
			Format: "json",
			Output: "stdout",
		})
	}
	return defaultLogger
}

// SetOutput 设置输出
func SetOutput(out io.Writer) {
	if defaultLogger != nil {
		defaultLogger.SetOutput(out)
	}
}
