package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

// Config 应用程序配置结构
type Config struct {
	App             AppConfig                `json:"app"`
	Database        DatabaseConfig           `json:"database"`
	Redis           RedisConfig              `json:"redis"`
	JWT             JWTConfig                `json:"jwt"`
	PasswordEncrypt PasswordEncryptionConfig `json:"password_encrypt"`
	Log             LogConfig                `json:"log"`
	Security        SecurityConfig           `json:"security"`
	Health          HealthConfig             `json:"health"`
}

// AppConfig 应用程序基础配置
type AppConfig struct {
	Name    string `json:"name" env:"APP_NAME" default:"trusioo_api"`
	Version string `json:"version" env:"APP_VERSION" default:"v0.0.1"`
	Env     string `json:"env" env:"APP_ENV" default:"development"`
	Host    string `json:"host" env:"APP_HOST" default:"0.0.0.0"`
	Port    string `json:"port" env:"APP_PORT" default:"8080"`
	Debug   bool   `json:"debug" env:"DEBUG" default:"true"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string        `json:"host" env:"DB_HOST" default:"localhost"`
	Port            string        `json:"port" env:"DB_PORT" default:"5432"`
	User            string        `json:"user" env:"DB_USER" default:"postgres"`
	Password        string        `json:"password" env:"DB_PASSWORD" default:"password"`
	Name            string        `json:"name" env:"DB_NAME" default:"trusioo_api"`
	SSLMode         string        `json:"ssl_mode" env:"DB_SSL_MODE" default:"disable"`
	MaxIdleConns    int           `json:"max_idle_conns" env:"DB_MAX_IDLE_CONNS" default:"10"`
	MaxOpenConns    int           `json:"max_open_conns" env:"DB_MAX_OPEN_CONNS" default:"100"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME" default:"60m"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host         string        `json:"host" env:"REDIS_HOST" default:"localhost"`
	Port         string        `json:"port" env:"REDIS_PORT" default:"6379"`
	Password     string        `json:"password" env:"REDIS_PASSWORD" default:""`
	DB           int           `json:"db" env:"REDIS_DB" default:"0"`
	PoolSize     int           `json:"pool_size" env:"REDIS_POOL_SIZE" default:"10"`
	MinIdleConns int           `json:"min_idle_conns" env:"REDIS_MIN_IDLE_CONNS" default:"5"`
	DialTimeout  time.Duration `json:"dial_timeout" env:"REDIS_DIAL_TIMEOUT" default:"5s"`
	ReadTimeout  time.Duration `json:"read_timeout" env:"REDIS_READ_TIMEOUT" default:"3s"`
	WriteTimeout time.Duration `json:"write_timeout" env:"REDIS_WRITE_TIMEOUT" default:"3s"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret                string        `json:"secret" env:"JWT_SECRET" default:"your-super-secret-jwt-key"`
	ExpireHours           int           `json:"expire_hours" env:"JWT_EXPIRE_HOURS" default:"24"`
	RefreshExpireHours    int           `json:"refresh_expire_hours" env:"JWT_REFRESH_EXPIRE_HOURS" default:"168"`
	ExpireDuration        time.Duration `json:"-"`
	RefreshExpireDuration time.Duration `json:"-"`
}

// PasswordEncryptionConfig 密码加密配置
type PasswordEncryptionConfig struct {
	Key    string `json:"key" env:"PASSWORD_ENCRYPTION_KEY" default:"your-ultra-secret-password-encryption-key"`
	Method string `json:"method" env:"PASSWORD_ENCRYPTION_METHOD" default:"hmac-bcrypt"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `json:"level" env:"LOG_LEVEL" default:"debug"`
	Format     string `json:"format" env:"LOG_FORMAT" default:"json"`
	Output     string `json:"output" env:"LOG_OUTPUT" default:"stdout"`
	FilePath   string `json:"file_path" env:"LOG_FILE_PATH" default:"./logs/app.log"`
	MaxSize    int    `json:"max_size" env:"LOG_MAX_SIZE" default:"100"`
	MaxBackups int    `json:"max_backups" env:"LOG_MAX_BACKUPS" default:"3"`
	MaxAge     int    `json:"max_age" env:"LOG_MAX_AGE" default:"30"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	CORSAllowedOrigins   []string `json:"cors_allowed_origins"`
	CORSAllowedMethods   []string `json:"cors_allowed_methods"`
	CORSAllowedHeaders   []string `json:"cors_allowed_headers"`
	CORSAllowCredentials bool     `json:"cors_allow_credentials" env:"CORS_ALLOW_CREDENTIALS" default:"true"`
	RateLimitRPM         int      `json:"rate_limit_rpm" env:"RATE_LIMIT_RPM" default:"100"`
	RateLimitWindow      int      `json:"rate_limit_window" env:"RATE_LIMIT_WINDOW" default:"1"`
}

// HealthConfig 健康检查配置
type HealthConfig struct {
	CheckInterval time.Duration `json:"check_interval" env:"HEALTH_CHECK_INTERVAL" default:"30s"`
	Timeout       time.Duration `json:"timeout" env:"HEALTH_CHECK_TIMEOUT" default:"5s"`
}

// Load 加载配置
func Load() (*Config, error) {
	// 加载.env文件
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found, using environment variables")
	}

	cfg := &Config{}

	// 加载应用配置
	cfg.App = AppConfig{
		Name:    getEnv("APP_NAME", "trusioo_api"),
		Version: getEnv("APP_VERSION", "v0.0.1"),
		Env:     getEnv("APP_ENV", "development"),
		Host:    getEnv("APP_HOST", "0.0.0.0"),
		Port:    getEnv("APP_PORT", "8080"),
		Debug:   getEnvAsBool("DEBUG", true),
	}

	// 加载数据库配置
	cfg.Database = DatabaseConfig{
		Host:            getEnv("DB_HOST", "localhost"),
		Port:            getEnv("DB_PORT", "5432"),
		User:            getEnv("DB_USER", "postgres"),
		Password:        getEnv("DB_PASSWORD", "password"),
		Name:            getEnv("DB_NAME", "trusioo_api"),
		SSLMode:         getEnv("DB_SSL_MODE", "disable"),
		MaxIdleConns:    getEnvAsInt("DB_MAX_IDLE_CONNS", 10),
		MaxOpenConns:    getEnvAsInt("DB_MAX_OPEN_CONNS", 100),
		ConnMaxLifetime: getEnvAsDuration("DB_CONN_MAX_LIFETIME", 60*time.Minute),
	}

	// 加载Redis配置
	cfg.Redis = RedisConfig{
		Host:         getEnv("REDIS_HOST", "localhost"),
		Port:         getEnv("REDIS_PORT", "6379"),
		Password:     getEnv("REDIS_PASSWORD", ""),
		DB:           getEnvAsInt("REDIS_DB", 0),
		PoolSize:     getEnvAsInt("REDIS_POOL_SIZE", 10),
		MinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5),
		DialTimeout:  getEnvAsDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
		ReadTimeout:  getEnvAsDuration("REDIS_READ_TIMEOUT", 3*time.Second),
		WriteTimeout: getEnvAsDuration("REDIS_WRITE_TIMEOUT", 3*time.Second),
	}

	// 加载JWT配置
	cfg.JWT = JWTConfig{
		Secret:             getEnv("JWT_SECRET", "your-super-secret-jwt-key"),
		ExpireHours:        getEnvAsInt("JWT_EXPIRE_HOURS", 24),
		RefreshExpireHours: getEnvAsInt("JWT_REFRESH_EXPIRE_HOURS", 168),
	}
	cfg.JWT.ExpireDuration = time.Duration(cfg.JWT.ExpireHours) * time.Hour
	cfg.JWT.RefreshExpireDuration = time.Duration(cfg.JWT.RefreshExpireHours) * time.Hour

	// 加载密码加密配置
	cfg.PasswordEncrypt = PasswordEncryptionConfig{
		Key:    getEnv("PASSWORD_ENCRYPTION_KEY", "your-ultra-secret-password-encryption-key"),
		Method: getEnv("PASSWORD_ENCRYPTION_METHOD", "hmac-bcrypt"),
	}

	// 加载日志配置
	cfg.Log = LogConfig{
		Level:      getEnv("LOG_LEVEL", "debug"),
		Format:     getEnv("LOG_FORMAT", "json"),
		Output:     getEnv("LOG_OUTPUT", "stdout"),
		FilePath:   getEnv("LOG_FILE_PATH", "./logs/app.log"),
		MaxSize:    getEnvAsInt("LOG_MAX_SIZE", 100),
		MaxBackups: getEnvAsInt("LOG_MAX_BACKUPS", 3),
		MaxAge:     getEnvAsInt("LOG_MAX_AGE", 30),
	}

	// 加载安全配置
	cfg.Security = SecurityConfig{
		CORSAllowedOrigins:   getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"*"}),
		CORSAllowedMethods:   getEnvAsSlice("CORS_ALLOWED_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		CORSAllowedHeaders:   getEnvAsSlice("CORS_ALLOWED_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}),
		CORSAllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
		RateLimitRPM:         getEnvAsInt("RATE_LIMIT_RPM", 100),
		RateLimitWindow:      getEnvAsInt("RATE_LIMIT_WINDOW", 1),
	}

	// 加载健康检查配置
	cfg.Health = HealthConfig{
		CheckInterval: getEnvAsDuration("HEALTH_CHECK_INTERVAL", 30*time.Second),
		Timeout:       getEnvAsDuration("HEALTH_CHECK_TIMEOUT", 5*time.Second),
	}

	return cfg, nil
}

// GetDSN 获取数据库连接字符串
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

// GetRedisAddr 获取Redis地址
func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

// IsProduction 判断是否为生产环境
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}

// IsDevelopment 判断是否为开发环境
func (c *Config) IsDevelopment() bool {
	return c.App.Env == "development"
}

// 辅助函数

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	strValue := getEnv(key, "")
	if value, err := strconv.ParseBool(strValue); err == nil {
		return value
	}
	return fallback
}

func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	strValue := getEnv(key, "")
	if value, err := time.ParseDuration(strValue); err == nil {
		return value
	}
	return fallback
}

func getEnvAsSlice(key string, fallback []string) []string {
	strValue := getEnv(key, "")
	if strValue == "" {
		return fallback
	}
	// 简单的逗号分割实现
	result := make([]string, 0)
	for _, item := range strings.Split(strValue, ",") {
		if trimmed := strings.TrimSpace(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}
