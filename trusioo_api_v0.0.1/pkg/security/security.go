// Package security 提供安全增强功能
package security

import (
	"html"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// RateLimiter 限流器接口
type RateLimiter interface {
	Allow(key string) bool
	Remaining(key string) int
	Reset(key string) time.Time
}

// TokenBucketLimiter 令牌桶限流器
type TokenBucketLimiter struct {
	limiters map[string]*rate.Limiter
	mutex    sync.RWMutex
	rate     rate.Limit
	burst    int
	logger   *logrus.Logger
}

// NewTokenBucketLimiter 创建令牌桶限流器
func NewTokenBucketLimiter(rateLimit float64, burst int, logger *logrus.Logger) *TokenBucketLimiter {
	return &TokenBucketLimiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     rate.Limit(rateLimit),
		burst:    burst,
		logger:   logger,
	}
}

// Allow 检查是否允许请求
func (tbl *TokenBucketLimiter) Allow(key string) bool {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()

	limiter, exists := tbl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(tbl.rate, tbl.burst)
		tbl.limiters[key] = limiter
	}

	return limiter.Allow()
}

// Remaining 获取剩余令牌数
func (tbl *TokenBucketLimiter) Remaining(key string) int {
	tbl.mutex.RLock()
	defer tbl.mutex.RUnlock()

	limiter, exists := tbl.limiters[key]
	if !exists {
		return tbl.burst
	}

	return int(limiter.Tokens())
}

// Reset 重置限流器
func (tbl *TokenBucketLimiter) Reset(key string) time.Time {
	tbl.mutex.Lock()
	defer tbl.mutex.Unlock()

	delete(tbl.limiters, key)
	return time.Now().Add(time.Hour) // 1小时后重置
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Rate         float64                   `json:"rate"`          // 每秒请求数
	Burst        int                       `json:"burst"`         // 突发请求数
	KeyGenerator func(*gin.Context) string `json:"-"`             // 键生成函数
	SkipPaths    []string                  `json:"skip_paths"`    // 跳过的路径
	SkipIPs      []string                  `json:"skip_ips"`      // 跳过的IP
	ErrorMessage string                    `json:"error_message"` // 错误消息
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(limiter RateLimiter, config *RateLimitConfig, logger *logrus.Logger) gin.HandlerFunc {
	if config == nil {
		config = &RateLimitConfig{
			Rate:         10, // 10 requests per second
			Burst:        20, // burst of 20
			ErrorMessage: "Rate limit exceeded",
			KeyGenerator: func(c *gin.Context) string {
				return c.ClientIP()
			},
		}
	}

	return func(c *gin.Context) {
		// 检查跳过路径
		for _, path := range config.SkipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// 检查跳过IP
		clientIP := c.ClientIP()
		for _, ip := range config.SkipIPs {
			if clientIP == ip {
				c.Next()
				return
			}
		}

		// 生成限流键
		key := config.KeyGenerator(c)

		// 检查限流
		if !limiter.Allow(key) {
			remaining := limiter.Remaining(key)
			resetTime := limiter.Reset(key)

			logger.WithFields(logrus.Fields{
				"ip":        clientIP,
				"key":       key,
				"path":      c.Request.URL.Path,
				"remaining": remaining,
				"reset":     resetTime,
			}).Warn("Rate limit exceeded")

			c.Header("X-Rate-Limit-Remaining", strconv.Itoa(remaining))
			c.Header("X-Rate-Limit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"success":    false,
				"message":    config.ErrorMessage,
				"error_code": "RATE_LIMIT_EXCEEDED",
				"remaining":  remaining,
				"reset_at":   resetTime.Unix(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SQLInjectionProtector SQL注入防护器
type SQLInjectionProtector struct {
	patterns []*regexp.Regexp
	logger   *logrus.Logger
}

// NewSQLInjectionProtector 创建SQL注入防护器
func NewSQLInjectionProtector(logger *logrus.Logger) *SQLInjectionProtector {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)(\bunion\s+select\b)`),
		regexp.MustCompile(`(?i)(\bselect\s+.*\bfrom\b)`),
		regexp.MustCompile(`(?i)(\binsert\s+into\b)`),
		regexp.MustCompile(`(?i)(\bupdate\s+.*\bset\b)`),
		regexp.MustCompile(`(?i)(\bdelete\s+from\b)`),
		regexp.MustCompile(`(?i)(\bdrop\s+table\b)`),
		regexp.MustCompile(`(?i)(\bcreate\s+table\b)`),
		regexp.MustCompile(`(?i)(\balter\s+table\b)`),
		regexp.MustCompile(`(?i)(\bexec\s*\()`),
		regexp.MustCompile(`(?i)(\bexecute\s*\()`),
		regexp.MustCompile(`(?i)(--|\#|\/\*|\*\/)`),
		regexp.MustCompile(`(?i)(\bor\s+1\s*=\s*1\b)`),
		regexp.MustCompile(`(?i)(\band\s+1\s*=\s*1\b)`),
		regexp.MustCompile(`(?i)(\'\s*or\s*\'\s*=\s*\')`),
		regexp.MustCompile(`(?i)(\"\s*or\s*\"\s*=\s*\")`),
	}

	return &SQLInjectionProtector{
		patterns: patterns,
		logger:   logger,
	}
}

// IsSQLInjection 检测是否为SQL注入
func (sip *SQLInjectionProtector) IsSQLInjection(input string) bool {
	for _, pattern := range sip.patterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// SanitizeSQL 清理SQL注入
func (sip *SQLInjectionProtector) SanitizeSQL(input string) string {
	sanitized := input
	for _, pattern := range sip.patterns {
		sanitized = pattern.ReplaceAllString(sanitized, "")
	}
	return strings.TrimSpace(sanitized)
}

// SQLInjectionMiddleware SQL注入防护中间件
func SQLInjectionMiddleware(protector *SQLInjectionProtector, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查查询参数
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if protector.IsSQLInjection(value) {
					logger.WithFields(logrus.Fields{
						"ip":        c.ClientIP(),
						"path":      c.Request.URL.Path,
						"parameter": key,
						"value":     value,
					}).Warn("SQL injection attempt detected")

					c.JSON(http.StatusBadRequest, gin.H{
						"success":    false,
						"message":    "Invalid input detected",
						"error_code": "INVALID_INPUT",
					})
					c.Abort()
					return
				}
			}
		}

		// 检查表单数据
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			if err := c.Request.ParseForm(); err == nil {
				for key, values := range c.Request.PostForm {
					for _, value := range values {
						if protector.IsSQLInjection(value) {
							logger.WithFields(logrus.Fields{
								"ip":        c.ClientIP(),
								"path":      c.Request.URL.Path,
								"parameter": key,
								"value":     value,
							}).Warn("SQL injection attempt in form data")

							c.JSON(http.StatusBadRequest, gin.H{
								"success":    false,
								"message":    "Invalid input detected",
								"error_code": "INVALID_INPUT",
							})
							c.Abort()
							return
						}
					}
				}
			}
		}

		c.Next()
	}
}

// XSSProtector XSS防护器
type XSSProtector struct {
	patterns []*regexp.Regexp
	logger   *logrus.Logger
}

// NewXSSProtector 创建XSS防护器
func NewXSSProtector(logger *logrus.Logger) *XSSProtector {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
		regexp.MustCompile(`(?i)<iframe[^>]*>.*?</iframe>`),
		regexp.MustCompile(`(?i)<object[^>]*>.*?</object>`),
		regexp.MustCompile(`(?i)<embed[^>]*>`),
		regexp.MustCompile(`(?i)<link[^>]*>`),
		regexp.MustCompile(`(?i)<meta[^>]*>`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)vbscript:`),
		regexp.MustCompile(`(?i)onload\s*=`),
		regexp.MustCompile(`(?i)onerror\s*=`),
		regexp.MustCompile(`(?i)onclick\s*=`),
		regexp.MustCompile(`(?i)onmouseover\s*=`),
		regexp.MustCompile(`(?i)onfocus\s*=`),
		regexp.MustCompile(`(?i)onblur\s*=`),
		regexp.MustCompile(`(?i)onchange\s*=`),
		regexp.MustCompile(`(?i)onsubmit\s*=`),
	}

	return &XSSProtector{
		patterns: patterns,
		logger:   logger,
	}
}

// IsXSS 检测是否为XSS攻击
func (xp *XSSProtector) IsXSS(input string) bool {
	for _, pattern := range xp.patterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// SanitizeXSS 清理XSS攻击代码
func (xp *XSSProtector) SanitizeXSS(input string) string {
	sanitized := html.EscapeString(input)
	for _, pattern := range xp.patterns {
		sanitized = pattern.ReplaceAllString(sanitized, "")
	}
	return sanitized
}

// XSSMiddleware XSS防护中间件
func XSSMiddleware(protector *XSSProtector, logger *logrus.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查查询参数
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if protector.IsXSS(value) {
					logger.WithFields(logrus.Fields{
						"ip":        c.ClientIP(),
						"path":      c.Request.URL.Path,
						"parameter": key,
						"value":     value,
					}).Warn("XSS attempt detected")

					c.JSON(http.StatusBadRequest, gin.H{
						"success":    false,
						"message":    "Invalid input detected",
						"error_code": "INVALID_INPUT",
					})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// SecurityHeaders 安全头中间件
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// XSS防护
		c.Header("X-XSS-Protection", "1; mode=block")

		// 内容类型嗅探防护
		c.Header("X-Content-Type-Options", "nosniff")

		// 点击劫持防护
		c.Header("X-Frame-Options", "DENY")

		// HSTS (仅在HTTPS下)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// 内容安全策略
		csp := "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'"
		c.Header("Content-Security-Policy", csp)

		// 引用者策略
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// 权限策略
		c.Header("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		c.Next()
	}
}

// CORS 跨域资源共享中间件
func CORSMiddleware(allowOrigins []string, allowMethods []string, allowHeaders []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 检查允许的源
		allowed := false
		for _, allowOrigin := range allowOrigins {
			if allowOrigin == "*" || allowOrigin == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", strings.Join(allowMethods, ", "))
		c.Header("Access-Control-Allow-Headers", strings.Join(allowHeaders, ", "))
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	RateLimit *RateLimitConfig `json:"rate_limit"`
	CORS      *CORSConfig      `json:"cors"`
	Headers   *HeadersConfig   `json:"headers"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins []string `json:"allow_origins"`
	AllowMethods []string `json:"allow_methods"`
	AllowHeaders []string `json:"allow_headers"`
}

// HeadersConfig 安全头配置
type HeadersConfig struct {
	EnableHSTS               bool   `json:"enable_hsts"`
	EnableCSP                bool   `json:"enable_csp"`
	CustomCSP                string `json:"custom_csp"`
	EnableFrameOptions       bool   `json:"enable_frame_options"`
	EnableContentTypeOptions bool   `json:"enable_content_type_options"`
}

// DefaultSecurityConfig 返回默认安全配置
func DefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		RateLimit: &RateLimitConfig{
			Rate:         10,
			Burst:        20,
			ErrorMessage: "Rate limit exceeded",
		},
		CORS: &CORSConfig{
			AllowOrigins: []string{"http://localhost:3000", "http://localhost:8080"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		},
		Headers: &HeadersConfig{
			EnableHSTS:               true,
			EnableCSP:                true,
			EnableFrameOptions:       true,
			EnableContentTypeOptions: true,
		},
	}
}
