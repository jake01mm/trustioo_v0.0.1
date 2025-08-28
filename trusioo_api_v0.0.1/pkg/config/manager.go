// Package config 提供配置管理增强功能
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	configs     map[string]interface{}
	watchers    map[string]*ConfigWatcher
	validators  map[string]ConfigValidator
	mutex       sync.RWMutex
	logger      *logrus.Logger
	environment string
}

// ConfigWatcher 配置监听器
type ConfigWatcher struct {
	path     string
	callback func(string, interface{})
	lastMod  time.Time
	active   bool
}

// ConfigValidator 配置验证器
type ConfigValidator interface {
	Validate(value interface{}) error
}

// ValidationRule 验证规则
type ValidationRule struct {
	Required bool
	Type     string
	Min      interface{}
	Max      interface{}
	Pattern  string
	Custom   func(interface{}) error
}

// Validate 实现ConfigValidator接口
func (vr *ValidationRule) Validate(value interface{}) error {
	if vr.Required && value == nil {
		return fmt.Errorf("value is required")
	}

	if value == nil {
		return nil
	}

	// 类型验证
	if vr.Type != "" {
		if err := vr.validateType(value); err != nil {
			return err
		}
	}

	// 范围验证
	if vr.Min != nil || vr.Max != nil {
		if err := vr.validateRange(value); err != nil {
			return err
		}
	}

	// 自定义验证
	if vr.Custom != nil {
		return vr.Custom(value)
	}

	return nil
}

// validateType 验证类型
func (vr *ValidationRule) validateType(value interface{}) error {
	valueType := reflect.TypeOf(value).Kind().String()

	switch vr.Type {
	case "string":
		if valueType != "string" {
			return fmt.Errorf("expected string, got %s", valueType)
		}
	case "int":
		if valueType != "int" && valueType != "int64" && valueType != "int32" {
			return fmt.Errorf("expected int, got %s", valueType)
		}
	case "float":
		if valueType != "float64" && valueType != "float32" {
			return fmt.Errorf("expected float, got %s", valueType)
		}
	case "bool":
		if valueType != "bool" {
			return fmt.Errorf("expected bool, got %s", valueType)
		}
	}

	return nil
}

// validateRange 验证范围
func (vr *ValidationRule) validateRange(value interface{}) error {
	switch v := value.(type) {
	case string:
		if vr.Min != nil {
			if min, ok := vr.Min.(int); ok && len(v) < min {
				return fmt.Errorf("string length %d is less than minimum %d", len(v), min)
			}
		}
		if vr.Max != nil {
			if max, ok := vr.Max.(int); ok && len(v) > max {
				return fmt.Errorf("string length %d is greater than maximum %d", len(v), max)
			}
		}
	case int, int32, int64:
		val := reflect.ValueOf(v).Int()
		if vr.Min != nil {
			if min := reflect.ValueOf(vr.Min).Int(); val < min {
				return fmt.Errorf("value %d is less than minimum %d", val, min)
			}
		}
		if vr.Max != nil {
			if max := reflect.ValueOf(vr.Max).Int(); val > max {
				return fmt.Errorf("value %d is greater than maximum %d", val, max)
			}
		}
	case float32, float64:
		val := reflect.ValueOf(v).Float()
		if vr.Min != nil {
			if min := reflect.ValueOf(vr.Min).Float(); val < min {
				return fmt.Errorf("value %f is less than minimum %f", val, min)
			}
		}
		if vr.Max != nil {
			if max := reflect.ValueOf(vr.Max).Float(); val > max {
				return fmt.Errorf("value %f is greater than maximum %f", val, max)
			}
		}
	}

	return nil
}

// NewConfigManager 创建配置管理器
func NewConfigManager(environment string, logger *logrus.Logger) *ConfigManager {
	return &ConfigManager{
		configs:     make(map[string]interface{}),
		watchers:    make(map[string]*ConfigWatcher),
		validators:  make(map[string]ConfigValidator),
		logger:      logger,
		environment: environment,
	}
}

// LoadConfig 加载配置
func (cm *ConfigManager) LoadConfig(key string, configPath string) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config interface{}

	// 根据文件扩展名解析配置
	ext := filepath.Ext(configPath)
	switch ext {
	case ".json":
		if err := json.Unmarshal(data, &config); err != nil {
			return fmt.Errorf("failed to parse JSON config: %w", err)
		}
	case ".env":
		config = cm.parseEnvFile(string(data))
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	// 验证配置
	if validator, exists := cm.validators[key]; exists {
		if err := validator.Validate(config); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}
	}

	cm.configs[key] = config

	cm.logger.WithFields(logrus.Fields{
		"key":  key,
		"path": configPath,
	}).Info("Configuration loaded")

	return nil
}

// parseEnvFile 解析环境变量文件
func (cm *ConfigManager) parseEnvFile(content string) map[string]string {
	result := make(map[string]string)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// 移除引号
			if len(value) >= 2 {
				if (value[0] == '"' && value[len(value)-1] == '"') ||
					(value[0] == '\'' && value[len(value)-1] == '\'') {
					value = value[1 : len(value)-1]
				}
			}

			result[key] = value
		}
	}

	return result
}

// GetConfig 获取配置
func (cm *ConfigManager) GetConfig(key string) (interface{}, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	config, exists := cm.configs[key]
	return config, exists
}

// GetString 获取字符串配置
func (cm *ConfigManager) GetString(key string, defaultValue ...string) string {
	if config, exists := cm.GetConfig(key); exists {
		if str, ok := config.(string); ok {
			return str
		}
		if m, ok := config.(map[string]string); ok {
			if len(defaultValue) > 0 {
				if val, exists := m[defaultValue[0]]; exists {
					return val
				}
			}
		}
	}

	// 尝试从环境变量获取
	if envVal := os.Getenv(key); envVal != "" {
		return envVal
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return ""
}

// GetInt 获取整数配置
func (cm *ConfigManager) GetInt(key string, defaultValue ...int) int {
	if config, exists := cm.GetConfig(key); exists {
		switch v := config.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case string:
			if val, err := strconv.Atoi(v); err == nil {
				return val
			}
		}
	}

	// 尝试从环境变量获取
	if envVal := os.Getenv(key); envVal != "" {
		if val, err := strconv.Atoi(envVal); err == nil {
			return val
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return 0
}

// GetBool 获取布尔配置
func (cm *ConfigManager) GetBool(key string, defaultValue ...bool) bool {
	if config, exists := cm.GetConfig(key); exists {
		switch v := config.(type) {
		case bool:
			return v
		case string:
			return v == "true" || v == "1" || v == "yes"
		}
	}

	// 尝试从环境变量获取
	if envVal := os.Getenv(key); envVal != "" {
		return envVal == "true" || envVal == "1" || envVal == "yes"
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return false
}

// GetFloat 获取浮点数配置
func (cm *ConfigManager) GetFloat(key string, defaultValue ...float64) float64 {
	if config, exists := cm.GetConfig(key); exists {
		switch v := config.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case string:
			if val, err := strconv.ParseFloat(v, 64); err == nil {
				return val
			}
		}
	}

	// 尝试从环境变量获取
	if envVal := os.Getenv(key); envVal != "" {
		if val, err := strconv.ParseFloat(envVal, 64); err == nil {
			return val
		}
	}

	if len(defaultValue) > 0 {
		return defaultValue[0]
	}

	return 0.0
}

// SetValidator 设置配置验证器
func (cm *ConfigManager) SetValidator(key string, validator ConfigValidator) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.validators[key] = validator
}

// Watch 监听配置文件变化
func (cm *ConfigManager) Watch(key string, configPath string, callback func(string, interface{})) error {
	// 获取文件信息
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	}

	watcher := &ConfigWatcher{
		path:     configPath,
		callback: callback,
		lastMod:  fileInfo.ModTime(),
		active:   true,
	}

	cm.mutex.Lock()
	cm.watchers[key] = watcher
	cm.mutex.Unlock()

	// 启动监听协程
	go cm.watchFile(key, watcher)

	cm.logger.WithFields(logrus.Fields{
		"key":  key,
		"path": configPath,
	}).Info("Started watching config file")

	return nil
}

// watchFile 监听文件变化
func (cm *ConfigManager) watchFile(key string, watcher *ConfigWatcher) {
	ticker := time.NewTicker(5 * time.Second) // 每5秒检查一次
	defer ticker.Stop()

	for range ticker.C {
		if !watcher.active {
			break
		}

		fileInfo, err := os.Stat(watcher.path)
		if err != nil {
			cm.logger.WithFields(logrus.Fields{
				"key":   key,
				"path":  watcher.path,
				"error": err,
			}).Warn("Failed to check config file")
			continue
		}

		if fileInfo.ModTime().After(watcher.lastMod) {
			cm.logger.WithFields(logrus.Fields{
				"key":  key,
				"path": watcher.path,
			}).Info("Config file changed, reloading")

			// 重新加载配置
			if err := cm.LoadConfig(key, watcher.path); err != nil {
				cm.logger.WithFields(logrus.Fields{
					"key":   key,
					"path":  watcher.path,
					"error": err,
				}).Error("Failed to reload config")
				continue
			}

			// 获取新配置
			if config, exists := cm.GetConfig(key); exists {
				watcher.callback(key, config)
			}

			watcher.lastMod = fileInfo.ModTime()
		}
	}
}

// StopWatch 停止监听配置文件
func (cm *ConfigManager) StopWatch(key string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if watcher, exists := cm.watchers[key]; exists {
		watcher.active = false
		delete(cm.watchers, key)

		cm.logger.WithField("key", key).Info("Stopped watching config file")
	}
}

// Reload 重新加载所有配置
func (cm *ConfigManager) Reload() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	for key, watcher := range cm.watchers {
		if err := cm.LoadConfig(key, watcher.path); err != nil {
			cm.logger.WithFields(logrus.Fields{
				"key":   key,
				"error": err,
			}).Error("Failed to reload config")
			return err
		}
	}

	cm.logger.Info("All configurations reloaded")
	return nil
}

// GetEnvironment 获取当前环境
func (cm *ConfigManager) GetEnvironment() string {
	return cm.environment
}

// SetEnvironment 设置环境
func (cm *ConfigManager) SetEnvironment(env string) {
	cm.environment = env
	cm.logger.WithField("environment", env).Info("Environment changed")
}

// IsProduction 检查是否为生产环境
func (cm *ConfigManager) IsProduction() bool {
	return cm.environment == "production" || cm.environment == "prod"
}

// IsDevelopment 检查是否为开发环境
func (cm *ConfigManager) IsDevelopment() bool {
	return cm.environment == "development" || cm.environment == "dev"
}

// GetConfigKeys 获取所有配置键
func (cm *ConfigManager) GetConfigKeys() []string {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	keys := make([]string, 0, len(cm.configs))
	for key := range cm.configs {
		keys = append(keys, key)
	}

	return keys
}

// RemoveConfig 移除配置
func (cm *ConfigManager) RemoveConfig(key string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.configs, key)
	delete(cm.validators, key)

	if watcher, exists := cm.watchers[key]; exists {
		watcher.active = false
		delete(cm.watchers, key)
	}

	cm.logger.WithField("key", key).Info("Configuration removed")
}
