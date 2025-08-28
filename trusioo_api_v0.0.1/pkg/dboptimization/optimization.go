// Package dboptimization 提供数据库连接池优化和监控功能
package dboptimization

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// PoolMonitor 连接池监控器
type PoolMonitor struct {
	db     *sql.DB
	logger *logrus.Logger

	// Prometheus指标
	openConnections   prometheus.Gauge
	inUseConnections  prometheus.Gauge
	idleConnections   prometheus.Gauge
	waitCount         prometheus.Counter
	waitDuration      prometheus.Histogram
	maxIdleClosed     prometheus.Counter
	maxLifetimeClosed prometheus.Counter

	// 配置
	config *PoolConfig

	// 监控状态
	isMonitoring bool
	stopChan     chan struct{}
	mutex        sync.RWMutex
}

// PoolConfig 连接池配置
type PoolConfig struct {
	MaxOpenConns    int           `json:"max_open_conns"`     // 最大打开连接数
	MaxIdleConns    int           `json:"max_idle_conns"`     // 最大空闲连接数
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`  // 连接最大生命周期
	ConnMaxIdleTime time.Duration `json:"conn_max_idle_time"` // 连接最大空闲时间

	// 监控配置
	MonitorInterval time.Duration `json:"monitor_interval"` // 监控间隔
	EnableMetrics   bool          `json:"enable_metrics"`   // 启用Prometheus指标

	// 慢查询配置
	SlowQueryThreshold time.Duration `json:"slow_query_threshold"` // 慢查询阈值
	EnableSlowLog      bool          `json:"enable_slow_log"`      // 启用慢查询日志
}

// SlowQueryLogger 慢查询日志器
type SlowQueryLogger struct {
	logger    *logrus.Logger
	threshold time.Duration
	enabled   bool

	// 统计指标
	slowQueryCount prometheus.Counter
	queryDuration  prometheus.Histogram
}

// QueryStats 查询统计信息
type QueryStats struct {
	Query     string        `json:"query"`
	Duration  time.Duration `json:"duration"`
	Args      []interface{} `json:"args"`
	Error     error         `json:"error"`
	Timestamp time.Time     `json:"timestamp"`
}

// NewPoolMonitor 创建连接池监控器
func NewPoolMonitor(db *sql.DB, config *PoolConfig, logger *logrus.Logger) *PoolMonitor {
	if config == nil {
		config = DefaultPoolConfig()
	}

	pm := &PoolMonitor{
		db:       db,
		logger:   logger,
		config:   config,
		stopChan: make(chan struct{}),
	}

	// 配置连接池
	pm.configurePool()

	// 初始化Prometheus指标
	if config.EnableMetrics {
		pm.initMetrics()
	}

	return pm
}

// configurePool 配置连接池
func (pm *PoolMonitor) configurePool() {
	pm.db.SetMaxOpenConns(pm.config.MaxOpenConns)
	pm.db.SetMaxIdleConns(pm.config.MaxIdleConns)
	pm.db.SetConnMaxLifetime(pm.config.ConnMaxLifetime)
	pm.db.SetConnMaxIdleTime(pm.config.ConnMaxIdleTime)

	pm.logger.WithFields(logrus.Fields{
		"max_open_conns":     pm.config.MaxOpenConns,
		"max_idle_conns":     pm.config.MaxIdleConns,
		"conn_max_lifetime":  pm.config.ConnMaxLifetime,
		"conn_max_idle_time": pm.config.ConnMaxIdleTime,
	}).Info("Database connection pool configured")
}

// initMetrics 初始化Prometheus指标
func (pm *PoolMonitor) initMetrics() {
	pm.openConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "database",
		Name:      "open_connections",
		Help:      "The number of open connections to the database",
	})

	pm.inUseConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "database",
		Name:      "in_use_connections",
		Help:      "The number of connections currently in use",
	})

	pm.idleConnections = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "database",
		Name:      "idle_connections",
		Help:      "The number of idle connections",
	})

	pm.waitCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "database",
		Name:      "wait_count_total",
		Help:      "The total number of connections waited for",
	})

	pm.waitDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "database",
		Name:      "wait_duration_seconds",
		Help:      "The time spent waiting for a connection",
		Buckets:   prometheus.DefBuckets,
	})

	pm.maxIdleClosed = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "database",
		Name:      "max_idle_closed_total",
		Help:      "The total number of connections closed due to SetMaxIdleConns",
	})

	pm.maxLifetimeClosed = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "database",
		Name:      "max_lifetime_closed_total",
		Help:      "The total number of connections closed due to SetConnMaxLifetime",
	})

	// 注册指标
	prometheus.MustRegister(
		pm.openConnections,
		pm.inUseConnections,
		pm.idleConnections,
		pm.waitCount,
		pm.waitDuration,
		pm.maxIdleClosed,
		pm.maxLifetimeClosed,
	)
}

// StartMonitoring 开始监控
func (pm *PoolMonitor) StartMonitoring() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if pm.isMonitoring {
		return
	}

	pm.isMonitoring = true
	go pm.monitorLoop()

	pm.logger.WithField("interval", pm.config.MonitorInterval).Info("Database pool monitoring started")
}

// StopMonitoring 停止监控
func (pm *PoolMonitor) StopMonitoring() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if !pm.isMonitoring {
		return
	}

	pm.isMonitoring = false
	close(pm.stopChan)

	pm.logger.Info("Database pool monitoring stopped")
}

// monitorLoop 监控循环
func (pm *PoolMonitor) monitorLoop() {
	ticker := time.NewTicker(pm.config.MonitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pm.collectMetrics()
		case <-pm.stopChan:
			return
		}
	}
}

// collectMetrics 收集指标
func (pm *PoolMonitor) collectMetrics() {
	stats := pm.db.Stats()

	// 更新Prometheus指标
	if pm.config.EnableMetrics {
		pm.openConnections.Set(float64(stats.OpenConnections))
		pm.inUseConnections.Set(float64(stats.InUse))
		pm.idleConnections.Set(float64(stats.Idle))
		pm.waitCount.Add(float64(stats.WaitCount))
		pm.waitDuration.Observe(stats.WaitDuration.Seconds())
		pm.maxIdleClosed.Add(float64(stats.MaxIdleClosed))
		pm.maxLifetimeClosed.Add(float64(stats.MaxLifetimeClosed))
	}

	// 记录日志
	pm.logger.WithFields(logrus.Fields{
		"open_connections":    stats.OpenConnections,
		"in_use_connections":  stats.InUse,
		"idle_connections":    stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration,
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}).Debug("Database connection pool stats")

	// 检查连接池健康状态
	pm.checkPoolHealth(stats)
}

// checkPoolHealth 检查连接池健康状态
func (pm *PoolMonitor) checkPoolHealth(stats sql.DBStats) {
	// 检查连接池是否接近饱和
	if stats.OpenConnections >= pm.config.MaxOpenConns*9/10 {
		pm.logger.WithFields(logrus.Fields{
			"open_connections": stats.OpenConnections,
			"max_open_conns":   pm.config.MaxOpenConns,
		}).Warn("Database connection pool is near saturation")
	}

	// 检查等待时间是否过长
	if stats.WaitDuration > time.Second {
		pm.logger.WithFields(logrus.Fields{
			"wait_duration": stats.WaitDuration,
			"wait_count":    stats.WaitCount,
		}).Warn("Database connection wait time is high")
	}

	// 检查空闲连接是否过多
	if stats.Idle > pm.config.MaxIdleConns*8/10 {
		pm.logger.WithFields(logrus.Fields{
			"idle_connections": stats.Idle,
			"max_idle_conns":   pm.config.MaxIdleConns,
		}).Info("High number of idle database connections")
	}
}

// GetStats 获取连接池统计信息
func (pm *PoolMonitor) GetStats() sql.DBStats {
	return pm.db.Stats()
}

// GetHealth 获取连接池健康状态
func (pm *PoolMonitor) GetHealth() map[string]interface{} {
	stats := pm.db.Stats()

	health := map[string]interface{}{
		"open_connections":    stats.OpenConnections,
		"in_use_connections":  stats.InUse,
		"idle_connections":    stats.Idle,
		"max_open_conns":      pm.config.MaxOpenConns,
		"max_idle_conns":      pm.config.MaxIdleConns,
		"wait_count":          stats.WaitCount,
		"wait_duration_ms":    stats.WaitDuration.Milliseconds(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}

	// 计算健康状态
	utilizationRate := float64(stats.OpenConnections) / float64(pm.config.MaxOpenConns)

	var status string
	if utilizationRate < 0.5 {
		status = "healthy"
	} else if utilizationRate < 0.8 {
		status = "warning"
	} else {
		status = "critical"
	}

	health["status"] = status
	health["utilization_rate"] = utilizationRate

	return health
}

// NewSlowQueryLogger 创建慢查询日志器
func NewSlowQueryLogger(threshold time.Duration, logger *logrus.Logger) *SlowQueryLogger {
	sql := &SlowQueryLogger{
		logger:    logger,
		threshold: threshold,
		enabled:   true,
	}

	// 初始化指标
	sql.slowQueryCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "database",
		Name:      "slow_query_total",
		Help:      "The total number of slow queries",
	})

	sql.queryDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: "database",
		Name:      "query_duration_seconds",
		Help:      "The query execution duration",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10},
	})

	prometheus.MustRegister(sql.slowQueryCount, sql.queryDuration)

	return sql
}

// LogQuery 记录查询日志
func (sql *SlowQueryLogger) LogQuery(stats QueryStats) {
	// 记录查询耗时指标
	sql.queryDuration.Observe(stats.Duration.Seconds())

	// 如果查询时间超过阈值，记录慢查询日志
	if sql.enabled && stats.Duration >= sql.threshold {
		sql.slowQueryCount.Inc()

		fields := logrus.Fields{
			"query":     stats.Query,
			"duration":  stats.Duration,
			"timestamp": stats.Timestamp,
		}

		if stats.Error != nil {
			fields["error"] = stats.Error.Error()
			sql.logger.WithFields(fields).Error("Slow query with error")
		} else {
			sql.logger.WithFields(fields).Warn("Slow query detected")
		}
	}
}

// Enable 启用慢查询日志
func (sql *SlowQueryLogger) Enable() {
	sql.enabled = true
}

// Disable 禁用慢查询日志
func (sql *SlowQueryLogger) Disable() {
	sql.enabled = false
}

// IsEnabled 检查是否启用
func (sql *SlowQueryLogger) IsEnabled() bool {
	return sql.enabled
}

// SetThreshold 设置慢查询阈值
func (sql *SlowQueryLogger) SetThreshold(threshold time.Duration) {
	sql.threshold = threshold
	sql.logger.WithField("threshold", threshold).Info("Slow query threshold updated")
}

// QueryWrapper 查询包装器，用于自动记录慢查询
type QueryWrapper struct {
	db         *sql.DB
	slowLogger *SlowQueryLogger
	logger     *logrus.Logger
}

// NewQueryWrapper 创建查询包装器
func NewQueryWrapper(db *sql.DB, slowLogger *SlowQueryLogger, logger *logrus.Logger) *QueryWrapper {
	return &QueryWrapper{
		db:         db,
		slowLogger: slowLogger,
		logger:     logger,
	}
}

// QueryContext 执行查询并记录日志
func (qw *QueryWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := qw.db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	stats := QueryStats{
		Query:     query,
		Duration:  duration,
		Args:      args,
		Error:     err,
		Timestamp: start,
	}

	qw.slowLogger.LogQuery(stats)

	return rows, err
}

// QueryRowContext 执行单行查询并记录日志
func (qw *QueryWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := qw.db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	stats := QueryStats{
		Query:     query,
		Duration:  duration,
		Args:      args,
		Error:     nil, // QueryRow不返回错误
		Timestamp: start,
	}

	qw.slowLogger.LogQuery(stats)

	return row
}

// ExecContext 执行SQL并记录日志
func (qw *QueryWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	result, err := qw.db.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	stats := QueryStats{
		Query:     query,
		Duration:  duration,
		Args:      args,
		Error:     err,
		Timestamp: start,
	}

	qw.slowLogger.LogQuery(stats)

	return result, err
}

// PrepareContext 准备SQL语句并记录日志
func (qw *QueryWrapper) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	start := time.Now()
	stmt, err := qw.db.PrepareContext(ctx, query)
	duration := time.Since(start)

	if duration >= qw.slowLogger.threshold {
		qw.logger.WithFields(logrus.Fields{
			"query":    query,
			"duration": duration,
			"type":     "prepare",
		}).Warn("Slow statement preparation")
	}

	return stmt, err
}

// DefaultPoolConfig 返回默认连接池配置
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxOpenConns:       25,               // 最大25个打开连接
		MaxIdleConns:       10,               // 最大10个空闲连接
		ConnMaxLifetime:    time.Hour,        // 连接最大生命周期1小时
		ConnMaxIdleTime:    time.Minute * 15, // 连接最大空闲时间15分钟
		MonitorInterval:    time.Second * 30, // 监控间隔30秒
		EnableMetrics:      true,             // 启用指标
		SlowQueryThreshold: time.Second,      // 慢查询阈值1秒
		EnableSlowLog:      true,             // 启用慢查询日志
	}
}

// OptimizePoolConfig 根据环境优化连接池配置
func OptimizePoolConfig(environment string, maxConcurrentUsers int) *PoolConfig {
	config := DefaultPoolConfig()

	switch environment {
	case "development":
		config.MaxOpenConns = 5
		config.MaxIdleConns = 2
		config.MonitorInterval = time.Minute
		config.SlowQueryThreshold = time.Millisecond * 500

	case "testing":
		config.MaxOpenConns = 3
		config.MaxIdleConns = 1
		config.ConnMaxLifetime = time.Minute * 10
		config.SlowQueryThreshold = time.Millisecond * 100

	case "production":
		// 根据并发用户数动态调整
		config.MaxOpenConns = maxConcurrentUsers/2 + 10
		if config.MaxOpenConns > 100 {
			config.MaxOpenConns = 100
		}
		config.MaxIdleConns = config.MaxOpenConns / 3
		config.ConnMaxLifetime = time.Hour * 2
		config.SlowQueryThreshold = time.Millisecond * 2000

	case "staging":
		config.MaxOpenConns = 15
		config.MaxIdleConns = 5
		config.SlowQueryThreshold = time.Millisecond * 1000
	}

	return config
}
