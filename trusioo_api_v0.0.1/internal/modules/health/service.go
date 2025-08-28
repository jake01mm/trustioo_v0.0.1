package health

import (
	"context"
	"time"

	"trusioo_api_v0.0.1/internal/infrastructure/database"
	"trusioo_api_v0.0.1/internal/infrastructure/redis"

	"github.com/sirupsen/logrus"
)

// CheckResult 检查结果结构
type CheckResult struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

// Service 健康检查服务
type Service struct {
	db     *database.Database
	redis  *redis.Client
	logger *logrus.Logger
}

// NewService 创建新的健康检查服务
func NewService(db *database.Database, redis *redis.Client, logger *logrus.Logger) *Service {
	return &Service{
		db:     db,
		redis:  redis,
		logger: logger,
	}
}

// CheckAllServices 检查所有服务状态
func (s *Service) CheckAllServices(ctx context.Context) (map[string]CheckResult, error) {
	checks := make(map[string]CheckResult)

	// 并发检查各个服务
	dbChan := make(chan CheckResult, 1)
	redisChan := make(chan CheckResult, 1)

	// 检查数据库
	go func() {
		dbChan <- s.checkDatabaseService(ctx)
	}()

	// 检查Redis
	go func() {
		redisChan <- s.checkRedisService(ctx)
	}()

	// 收集结果
	checks["database"] = <-dbChan
	checks["redis"] = <-redisChan

	return checks, nil
}

// CheckDatabase 检查数据库服务
func (s *Service) CheckDatabase(ctx context.Context) CheckResult {
	return s.checkDatabaseService(ctx)
}

// CheckRedis 检查Redis服务
func (s *Service) CheckRedis(ctx context.Context) CheckResult {
	return s.checkRedisService(ctx)
}

// checkDatabaseService 内部数据库检查方法
func (s *Service) checkDatabaseService(ctx context.Context) CheckResult {
	start := time.Now()

	if s.db == nil {
		return CheckResult{
			Status:  "unhealthy",
			Message: "Database service not available",
			Error:   "database connection is nil",
		}
	}

	// 创建带超时的上下文
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 检查连接
	err := s.db.PingContext(checkCtx)
	latency := time.Since(start)

	if err != nil {
		s.logger.WithError(err).Error("Database health check failed")
		return CheckResult{
			Status:  "unhealthy",
			Message: "Database ping failed",
			Error:   err.Error(),
			Latency: latency.String(),
		}
	}

	// 获取连接池状态
	stats := s.db.GetStats()
	if stats.OpenConnections == 0 {
		return CheckResult{
			Status:  "unhealthy",
			Message: "No database connections available",
			Latency: latency.String(),
		}
	}

	return CheckResult{
		Status:  "healthy",
		Message: "Database service is healthy",
		Latency: latency.String(),
	}
}

// checkRedisService 内部Redis检查方法
func (s *Service) checkRedisService(ctx context.Context) CheckResult {
	start := time.Now()

	if s.redis == nil {
		return CheckResult{
			Status:  "unhealthy",
			Message: "Redis service not available",
			Error:   "redis connection is nil",
		}
	}

	// 创建带超时的上下文
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 执行ping命令
	err := s.redis.Ping(checkCtx).Err()
	latency := time.Since(start)

	if err != nil {
		s.logger.WithError(err).Error("Redis health check failed")
		return CheckResult{
			Status:  "unhealthy",
			Message: "Redis ping failed",
			Error:   err.Error(),
			Latency: latency.String(),
		}
	}

	// 测试基本的set/get操作
	testKey := "health_check_test"
	testValue := "ok"

	if err := s.redis.Set(checkCtx, testKey, testValue, time.Minute).Err(); err != nil {
		return CheckResult{
			Status:  "unhealthy",
			Message: "Redis set operation failed",
			Error:   err.Error(),
			Latency: latency.String(),
		}
	}

	if val, err := s.redis.Get(checkCtx, testKey).Result(); err != nil || val != testValue {
		return CheckResult{
			Status:  "unhealthy",
			Message: "Redis get operation failed",
			Error:   err.Error(),
			Latency: latency.String(),
		}
	}

	// 清理测试键
	s.redis.Del(checkCtx, testKey)

	return CheckResult{
		Status:  "healthy",
		Message: "Redis service is healthy",
		Latency: latency.String(),
	}
}

// IsHealthy 检查整体服务是否健康
func (s *Service) IsHealthy(ctx context.Context) bool {
	checks, err := s.CheckAllServices(ctx)
	if err != nil {
		return false
	}

	for _, check := range checks {
		if check.Status != "healthy" {
			return false
		}
	}

	return true
}
