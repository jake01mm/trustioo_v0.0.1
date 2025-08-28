// Package metrics 提供监控指标收集和性能分析功能
package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
	"github.com/sirupsen/logrus"
)

// Metrics 指标收集器
type Metrics struct {
	// HTTP请求相关指标
	httpRequestsTotal    *prometheus.CounterVec
	httpRequestDuration  *prometheus.HistogramVec
	httpRequestSize      *prometheus.HistogramVec
	httpResponseSize     *prometheus.HistogramVec
	httpRequestsInFlight prometheus.Gauge

	// 应用程序指标
	activeConnections   prometheus.Gauge
	databaseConnections *prometheus.GaugeVec
	redisConnections    *prometheus.GaugeVec

	// 业务指标
	userRegistrations *prometheus.CounterVec
	userLogins        *prometheus.CounterVec
	apiErrors         *prometheus.CounterVec

	// 系统资源指标
	memoryUsage prometheus.Gauge
	cpuUsage    prometheus.Gauge

	logger *logrus.Logger
}

// Config 监控配置
type Config struct {
	Namespace            string            `json:"namespace"`
	Subsystem            string            `json:"subsystem"`
	Labels               map[string]string `json:"labels"`
	EnableGoMetrics      bool              `json:"enable_go_metrics"`
	EnableProcessMetrics bool              `json:"enable_process_metrics"`
}

// NewMetrics 创建新的指标收集器
func NewMetrics(config *Config, logger *logrus.Logger) *Metrics {
	if config == nil {
		config = &Config{
			Namespace:            "trusioo",
			Subsystem:            "api",
			EnableGoMetrics:      true,
			EnableProcessMetrics: true,
		}
	}

	m := &Metrics{
		logger: logger,
	}

	// 初始化HTTP请求指标
	m.httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status_code"},
	)

	m.httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latency distributions",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	m.httpRequestSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "http_request_size_bytes",
			Help:      "HTTP request size in bytes",
			Buckets:   []float64{1, 10, 100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "endpoint"},
	)

	m.httpResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "http_response_size_bytes",
			Help:      "HTTP response size in bytes",
			Buckets:   []float64{1, 10, 100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "endpoint"},
	)

	m.httpRequestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "http_requests_in_flight",
			Help:      "Number of HTTP requests currently being processed",
		},
	)

	// 初始化连接指标
	m.activeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "active_connections",
			Help:      "Number of active connections",
		},
	)

	m.databaseConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "database_connections",
			Help:      "Number of database connections",
		},
		[]string{"database", "state"}, // state: idle, active, total
	)

	m.redisConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "redis_connections",
			Help:      "Number of Redis connections",
		},
		[]string{"state"}, // state: idle, active, total
	)

	// 初始化业务指标
	m.userRegistrations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "user_registrations_total",
			Help:      "Total number of user registrations",
		},
		[]string{"user_type"}, // admin, user
	)

	m.userLogins = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "user_logins_total",
			Help:      "Total number of user logins",
		},
		[]string{"user_type", "status"}, // status: success, failure
	)

	m.apiErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "api_errors_total",
			Help:      "Total number of API errors",
		},
		[]string{"error_type", "endpoint"},
	)

	// 初始化系统资源指标
	m.memoryUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "memory_usage_bytes",
			Help:      "Current memory usage in bytes",
		},
	)

	m.cpuUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: config.Namespace,
			Subsystem: config.Subsystem,
			Name:      "cpu_usage_percent",
			Help:      "Current CPU usage percentage",
		},
	)

	// 注册所有指标
	m.registerMetrics(config)

	return m
}

// registerMetrics 注册所有指标到Prometheus
func (m *Metrics) registerMetrics(config *Config) {
	prometheus.MustRegister(
		m.httpRequestsTotal,
		m.httpRequestDuration,
		m.httpRequestSize,
		m.httpResponseSize,
		m.httpRequestsInFlight,
		m.activeConnections,
		m.databaseConnections,
		m.redisConnections,
		m.userRegistrations,
		m.userLogins,
		m.apiErrors,
		m.memoryUsage,
		m.cpuUsage,
	)

	// 注册Go运行时指标
	if config.EnableGoMetrics {
		prometheus.MustRegister(prometheus.NewGoCollector())
	}

	// 注册进程指标
	if config.EnableProcessMetrics {
		prometheus.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	}
}

// HTTPMetricsMiddleware HTTP请求指标中间件
func (m *Metrics) HTTPMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 增加正在处理的请求数
		m.httpRequestsInFlight.Inc()
		defer m.httpRequestsInFlight.Dec()

		// 记录请求大小
		if c.Request.ContentLength > 0 {
			m.httpRequestSize.WithLabelValues(
				c.Request.Method,
				c.FullPath(),
			).Observe(float64(c.Request.ContentLength))
		}

		// 处理请求
		c.Next()

		// 计算请求处理时间
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())

		// 记录HTTP请求指标
		m.httpRequestsTotal.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
			statusCode,
		).Inc()

		m.httpRequestDuration.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(duration)

		// 记录响应大小
		m.httpResponseSize.WithLabelValues(
			c.Request.Method,
			c.FullPath(),
		).Observe(float64(c.Writer.Size()))

		// 记录错误
		if c.Writer.Status() >= 400 {
			errorType := "client_error"
			if c.Writer.Status() >= 500 {
				errorType = "server_error"
			}
			m.apiErrors.WithLabelValues(errorType, c.FullPath()).Inc()
		}
	}
}

// RecordUserRegistration 记录用户注册
func (m *Metrics) RecordUserRegistration(userType string) {
	m.userRegistrations.WithLabelValues(userType).Inc()
	m.logger.WithField("user_type", userType).Debug("User registration recorded")
}

// RecordUserLogin 记录用户登录
func (m *Metrics) RecordUserLogin(userType, status string) {
	m.userLogins.WithLabelValues(userType, status).Inc()
	m.logger.WithFields(logrus.Fields{
		"user_type": userType,
		"status":    status,
	}).Debug("User login recorded")
}

// RecordAPIError 记录API错误
func (m *Metrics) RecordAPIError(errorType, endpoint string) {
	m.apiErrors.WithLabelValues(errorType, endpoint).Inc()
	m.logger.WithFields(logrus.Fields{
		"error_type": errorType,
		"endpoint":   endpoint,
	}).Debug("API error recorded")
}

// UpdateDatabaseConnections 更新数据库连接指标
func (m *Metrics) UpdateDatabaseConnections(database, state string, count int) {
	m.databaseConnections.WithLabelValues(database, state).Set(float64(count))
}

// UpdateRedisConnections 更新Redis连接指标
func (m *Metrics) UpdateRedisConnections(state string, count int) {
	m.redisConnections.WithLabelValues(state).Set(float64(count))
}

// UpdateSystemMetrics 更新系统资源指标
func (m *Metrics) UpdateSystemMetrics(memoryBytes int64, cpuPercent float64) {
	m.memoryUsage.Set(float64(memoryBytes))
	m.cpuUsage.Set(cpuPercent)
}

// SetupMetricsRoute 设置指标路由
func (m *Metrics) SetupMetricsRoute(router *gin.Engine) {
	// Prometheus指标端点
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 健康检查端点（用于监控系统）
	router.GET("/health/metrics", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"metrics": gin.H{
				"http_requests_in_flight": m.getHttpRequestsInFlight(),
				"active_connections":      m.getActiveConnections(),
			},
		})
	})
}

// getHttpRequestsInFlight 获取当前正在处理的请求数
func (m *Metrics) getHttpRequestsInFlight() float64 {
	metric := &dto.Metric{}
	m.httpRequestsInFlight.Write(metric)
	return metric.GetGauge().GetValue()
}

// getActiveConnections 获取活跃连接数
func (m *Metrics) getActiveConnections() float64 {
	metric := &dto.Metric{}
	m.activeConnections.Write(metric)
	return metric.GetGauge().GetValue()
}

// PerformanceProfiler 性能分析器
type PerformanceProfiler struct {
	startTime time.Time
	operation string
	logger    *logrus.Logger
}

// NewPerformanceProfiler 创建性能分析器
func NewPerformanceProfiler(operation string, logger *logrus.Logger) *PerformanceProfiler {
	return &PerformanceProfiler{
		startTime: time.Now(),
		operation: operation,
		logger:    logger,
	}
}

// Stop 停止性能分析并记录结果
func (p *PerformanceProfiler) Stop() time.Duration {
	duration := time.Since(p.startTime)

	p.logger.WithFields(logrus.Fields{
		"operation":   p.operation,
		"duration":    duration.String(),
		"duration_ms": duration.Milliseconds(),
	}).Info("Performance profile completed")

	return duration
}

// DefaultConfig 返回默认监控配置
func DefaultConfig() *Config {
	return &Config{
		Namespace:            "trusioo",
		Subsystem:            "api",
		EnableGoMetrics:      true,
		EnableProcessMetrics: true,
		Labels: map[string]string{
			"version": "v0.0.1",
			"env":     "development",
		},
	}
}
