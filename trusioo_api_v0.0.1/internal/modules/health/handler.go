package health

import (
	"context"
	"net/http"
	"time"

	"trusioo_api_v0.0.1/internal/infrastructure/database"
	"trusioo_api_v0.0.1/internal/infrastructure/redis"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Handler 健康检查处理器
type Handler struct {
	db     *database.Database
	redis  *redis.Client
	logger *logrus.Logger
}

// NewHandler 创建新的健康检查处理器
func NewHandler(db *database.Database, redis *redis.Client, logger *logrus.Logger) *Handler {
	return &Handler{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// HealthStatus 健康状态结构
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	Checks    map[string]CheckResult `json:"checks"`
}

var startTime = time.Now()

// OverallHealth 整体健康检查
func (h *Handler) OverallHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	checks := make(map[string]CheckResult)
	overallStatus := "healthy"

	// 检查数据库
	dbCheck := h.checkDatabase(ctx)
	checks["database"] = dbCheck
	if dbCheck.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	// 检查Redis
	redisCheck := h.checkRedis(ctx)
	checks["redis"] = redisCheck
	if redisCheck.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	// 检查API
	apiCheck := h.checkAPI(ctx)
	checks["api"] = apiCheck
	if apiCheck.Status != "healthy" {
		overallStatus = "unhealthy"
	}

	status := HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   "v0.0.1",
		Uptime:    time.Since(startTime).String(),
		Checks:    checks,
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, status)
}

// DatabaseHealth 数据库健康检查
func (h *Handler) DatabaseHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	check := h.checkDatabase(ctx)

	statusCode := http.StatusOK
	if check.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"service": "database",
		"check":   check,
	})
}

// RedisHealth Redis健康检查
func (h *Handler) RedisHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	check := h.checkRedis(ctx)

	statusCode := http.StatusOK
	if check.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"service": "redis",
		"check":   check,
	})
}

// APIHealth API健康检查
func (h *Handler) APIHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	check := h.checkAPI(ctx)

	statusCode := http.StatusOK
	if check.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"service": "api",
		"version": "v1",
		"check":   check,
	})
}

// Readiness 就绪检查
func (h *Handler) Readiness(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]CheckResult)
	ready := true

	// 检查数据库连接
	dbCheck := h.checkDatabase(ctx)
	checks["database"] = dbCheck
	if dbCheck.Status != "healthy" {
		ready = false
	}

	// 检查Redis连接
	redisCheck := h.checkRedis(ctx)
	checks["redis"] = redisCheck
	if redisCheck.Status != "healthy" {
		ready = false
	}

	status := "ready"
	statusCode := http.StatusOK
	if !ready {
		status = "not_ready"
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"status": status,
		"checks": checks,
	})
}

// Liveness 存活检查
func (h *Handler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "alive",
		"timestamp": time.Now().Format(time.RFC3339),
		"uptime":    time.Since(startTime).String(),
	})
}

// checkDatabase 检查数据库状态
func (h *Handler) checkDatabase(ctx context.Context) CheckResult {
	start := time.Now()

	if h.db == nil {
		return CheckResult{
			Status:  "unhealthy",
			Message: "Database connection not initialized",
			Error:   "database connection is nil",
		}
	}

	err := h.db.Health()
	latency := time.Since(start)

	if err != nil {
		h.logger.WithError(err).Error("Database health check failed")
		return CheckResult{
			Status:  "unhealthy",
			Message: "Database connection failed",
			Error:   err.Error(),
			Latency: latency.String(),
		}
	}

	// 获取数据库统计信息
	_ = h.db.GetStats()

	return CheckResult{
		Status:  "healthy",
		Message: "Database connection is healthy",
		Latency: latency.String(),
	}
}

// checkRedis 检查Redis状态
func (h *Handler) checkRedis(ctx context.Context) CheckResult {
	start := time.Now()

	if h.redis == nil {
		return CheckResult{
			Status:  "unhealthy",
			Message: "Redis connection not initialized",
			Error:   "redis connection is nil",
		}
	}

	err := h.redis.Health()
	latency := time.Since(start)

	if err != nil {
		h.logger.WithError(err).Error("Redis health check failed")
		return CheckResult{
			Status:  "unhealthy",
			Message: "Redis connection failed",
			Error:   err.Error(),
			Latency: latency.String(),
		}
	}

	return CheckResult{
		Status:  "healthy",
		Message: "Redis connection is healthy",
		Latency: latency.String(),
	}
}

// checkAPI 检查API状态
func (h *Handler) checkAPI(ctx context.Context) CheckResult {
	start := time.Now()

	// 这里可以添加更多的API健康检查逻辑
	// 比如检查关键服务的可用性等

	latency := time.Since(start)

	return CheckResult{
		Status:  "healthy",
		Message: "API is operational",
		Latency: latency.String(),
	}
}
