package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"trusioo_api_v0.0.1/internal/config"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// Database 数据库连接结构
type Database struct {
	*sql.DB
	config *config.DatabaseConfig
	logger *logrus.Logger
}

// New 创建新的数据库连接
func New(cfg *config.DatabaseConfig, logger *logrus.Logger) (*Database, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// 配置连接池
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 测试连接
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established successfully")

	return &Database{
		DB:     db,
		config: cfg,
		logger: logger,
	}, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	if d.DB != nil {
		if err := d.DB.Close(); err != nil {
			d.logger.WithError(err).Error("Error closing database connection")
			return err
		}
		d.logger.Info("Database connection closed")
	}
	return nil
}

// Health 检查数据库健康状态
func (d *Database) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := d.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

// GetStats 获取数据库连接统计信息
func (d *Database) GetStats() sql.DBStats {
	return d.DB.Stats()
}

// Transaction 执行事务
func (d *Database) Transaction(fn func(*sql.Tx) error) error {
	tx, err := d.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				d.logger.WithError(rollbackErr).Error("Failed to rollback transaction during panic")
			}
			panic(p)
		} else if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				d.logger.WithError(rollbackErr).Error("Failed to rollback transaction")
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				d.logger.WithError(commitErr).Error("Failed to commit transaction")
				err = commitErr
			}
		}
	}()

	err = fn(tx)
	return err
}

// QueryBuilder SQL查询构建器辅助结构
type QueryBuilder struct {
	query string
	args  []interface{}
	db    *Database
}

// NewQueryBuilder 创建新的查询构建器
func (d *Database) NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		db:   d,
		args: make([]interface{}, 0),
	}
}

// Select 设置SELECT语句
func (qb *QueryBuilder) Select(columns string) *QueryBuilder {
	qb.query = "SELECT " + columns
	return qb
}

// From 设置FROM子句
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.query += " FROM " + table
	return qb
}

// Where 添加WHERE条件
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	if qb.query == "" {
		return qb
	}

	qb.query += " WHERE " + condition
	qb.args = append(qb.args, args...)
	return qb
}

// And 添加AND条件
func (qb *QueryBuilder) And(condition string, args ...interface{}) *QueryBuilder {
	qb.query += " AND " + condition
	qb.args = append(qb.args, args...)
	return qb
}

// Or 添加OR条件
func (qb *QueryBuilder) Or(condition string, args ...interface{}) *QueryBuilder {
	qb.query += " OR " + condition
	qb.args = append(qb.args, args...)
	return qb
}

// OrderBy 添加ORDER BY子句
func (qb *QueryBuilder) OrderBy(column string) *QueryBuilder {
	qb.query += " ORDER BY " + column
	return qb
}

// Limit 添加LIMIT子句
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.query += fmt.Sprintf(" LIMIT %d", limit)
	return qb
}

// Offset 添加OFFSET子句
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.query += fmt.Sprintf(" OFFSET %d", offset)
	return qb
}

// Build 构建最终的SQL查询
func (qb *QueryBuilder) Build() (string, []interface{}) {
	return qb.query, qb.args
}

// Execute 执行查询
func (qb *QueryBuilder) Execute() (*sql.Rows, error) {
	query, args := qb.Build()
	return qb.db.Query(query, args...)
}

// ExecuteRow 执行查询并返回单行
func (qb *QueryBuilder) ExecuteRow() *sql.Row {
	query, args := qb.Build()
	return qb.db.QueryRow(query, args...)
}

// Repository 基础仓储接口
type Repository interface {
	GetDB() *Database
}

// BaseRepository 基础仓储实现
type BaseRepository struct {
	db     *Database
	logger *logrus.Logger
}

// NewBaseRepository 创建基础仓储
func NewBaseRepository(db *Database, logger *logrus.Logger) *BaseRepository {
	return &BaseRepository{
		db:     db,
		logger: logger,
	}
}

// GetDB 获取数据库连接
func (r *BaseRepository) GetDB() *Database {
	return r.db
}

// Insert 通用插入方法
func (r *BaseRepository) Insert(table string, data map[string]interface{}) (sql.Result, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided for insert")
	}

	columns := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data))

	i := 1
	for column, value := range data {
		columns = append(columns, column)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		values = append(values, value)
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return r.db.Exec(query, values...)
}

// Update 通用更新方法
func (r *BaseRepository) Update(table string, data map[string]interface{}, where string, whereArgs ...interface{}) (sql.Result, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("no data provided for update")
	}

	setParts := make([]string, 0, len(data))
	values := make([]interface{}, 0, len(data)+len(whereArgs))

	i := 1
	for column, value := range data {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", column, i))
		values = append(values, value)
		i++
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s", table, strings.Join(setParts, ", "), where)
	values = append(values, whereArgs...)

	return r.db.Exec(query, values...)
}

// Delete 通用删除方法
func (r *BaseRepository) Delete(table string, where string, whereArgs ...interface{}) (sql.Result, error) {
	query := fmt.Sprintf("DELETE FROM %s WHERE %s", table, where)
	return r.db.Exec(query, whereArgs...)
}

// Exists 检查记录是否存在
func (r *BaseRepository) Exists(table string, where string, whereArgs ...interface{}) (bool, error) {
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s)", table, where)
	var exists bool
	err := r.db.QueryRow(query, whereArgs...).Scan(&exists)
	return exists, err
}

// Count 计算记录数量
func (r *BaseRepository) Count(table string, where string, whereArgs ...interface{}) (int64, error) {
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, where)
	var count int64
	err := r.db.QueryRow(query, whereArgs...).Scan(&count)
	return count, err
}
