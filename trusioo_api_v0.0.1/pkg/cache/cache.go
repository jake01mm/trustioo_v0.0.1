// Package cache 提供缓存抽象层和策略管理
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Cache 缓存接口
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
	Clear(ctx context.Context) error
}

// RedisCache Redis缓存实现
type RedisCache struct {
	client *redis.Client
	logger *logrus.Logger
	prefix string
}

// NewRedisCache 创建Redis缓存实例
func NewRedisCache(client *redis.Client, prefix string, logger *logrus.Logger) *RedisCache {
	return &RedisCache{
		client: client,
		logger: logger,
		prefix: prefix,
	}
}

// Get 获取缓存值
func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	fullKey := r.getFullKey(key)
	result, err := r.client.Get(ctx, fullKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheNotFound
		}
		r.logger.WithFields(logrus.Fields{
			"key":   fullKey,
			"error": err,
		}).Error("Failed to get cache value")
		return nil, err
	}
	return result, nil
}

// Set 设置缓存值
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	fullKey := r.getFullKey(key)

	var data []byte
	var err error

	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		data, err = json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
	}

	err = r.client.Set(ctx, fullKey, data, expiration).Err()
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"key":        fullKey,
			"expiration": expiration,
			"error":      err,
		}).Error("Failed to set cache value")
		return err
	}

	return nil
}

// Delete 删除缓存
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := r.getFullKey(key)
	err := r.client.Del(ctx, fullKey).Err()
	if err != nil {
		r.logger.WithFields(logrus.Fields{
			"key":   fullKey,
			"error": err,
		}).Error("Failed to delete cache value")
		return err
	}
	return nil
}

// Exists 检查缓存是否存在
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := r.getFullKey(key)
	count, err := r.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// TTL 获取缓存过期时间
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := r.getFullKey(key)
	return r.client.TTL(ctx, fullKey).Result()
}

// Keys 获取匹配的键列表
func (r *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	fullPattern := r.getFullKey(pattern)
	keys, err := r.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return nil, err
	}

	// 移除前缀
	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = r.removePrefix(key)
	}

	return result, nil
}

// Clear 清空所有缓存
func (r *RedisCache) Clear(ctx context.Context) error {
	keys, err := r.Keys(ctx, "*")
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	// 添加前缀后删除
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.getFullKey(key)
	}

	return r.client.Del(ctx, fullKeys...).Err()
}

// getFullKey 获取完整键名
func (r *RedisCache) getFullKey(key string) string {
	if r.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", r.prefix, key)
}

// removePrefix 移除前缀
func (r *RedisCache) removePrefix(key string) string {
	if r.prefix == "" {
		return key
	}
	prefixLen := len(r.prefix) + 1 // +1 for ":"
	if len(key) <= prefixLen {
		return key
	}
	return key[prefixLen:]
}

// CacheManager 缓存管理器
type CacheManager struct {
	cache    Cache
	logger   *logrus.Logger
	strategy *CacheStrategy
}

// CacheStrategy 缓存策略
type CacheStrategy struct {
	DefaultTTL    time.Duration            `json:"default_ttl"`
	MaxTTL        time.Duration            `json:"max_ttl"`
	TypeTTLs      map[string]time.Duration `json:"type_ttls"`
	EnableMetrics bool                     `json:"enable_metrics"`

	// 缓存预热策略
	WarmupEnabled bool     `json:"warmup_enabled"`
	WarmupKeys    []string `json:"warmup_keys"`

	// 缓存失效策略
	EvictionPolicy string `json:"eviction_policy"` // LRU, LFU, TTL
	MaxMemory      int64  `json:"max_memory"`      // bytes
}

// NewCacheManager 创建缓存管理器
func NewCacheManager(cache Cache, strategy *CacheStrategy, logger *logrus.Logger) *CacheManager {
	if strategy == nil {
		strategy = DefaultStrategy()
	}

	return &CacheManager{
		cache:    cache,
		logger:   logger,
		strategy: strategy,
	}
}

// GetWithFallback 获取缓存，如果不存在则执行回调函数
func (cm *CacheManager) GetWithFallback(ctx context.Context, key string, fallback func() (interface{}, error)) ([]byte, error) {
	// 尝试从缓存获取
	data, err := cm.cache.Get(ctx, key)
	if err == nil {
		cm.logger.WithField("key", key).Debug("Cache hit")
		return data, nil
	}

	if err != ErrCacheNotFound {
		cm.logger.WithFields(logrus.Fields{
			"key":   key,
			"error": err,
		}).Warn("Cache get error, using fallback")
	} else {
		cm.logger.WithField("key", key).Debug("Cache miss")
	}

	// 执行回调函数
	value, err := fallback()
	if err != nil {
		return nil, fmt.Errorf("fallback function failed: %w", err)
	}

	// 将结果存入缓存
	ttl := cm.getTTLForKey(key)
	if err := cm.cache.Set(ctx, key, value, ttl); err != nil {
		cm.logger.WithFields(logrus.Fields{
			"key":   key,
			"error": err,
		}).Warn("Failed to set cache after fallback")
	}

	// 返回序列化后的数据
	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	default:
		return json.Marshal(value)
	}
}

// SetWithStrategy 使用策略设置缓存
func (cm *CacheManager) SetWithStrategy(ctx context.Context, key string, value interface{}) error {
	ttl := cm.getTTLForKey(key)
	return cm.cache.Set(ctx, key, value, ttl)
}

// DeletePattern 删除匹配模式的缓存
func (cm *CacheManager) DeletePattern(ctx context.Context, pattern string) error {
	keys, err := cm.cache.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := cm.cache.Delete(ctx, key); err != nil {
			cm.logger.WithFields(logrus.Fields{
				"key":   key,
				"error": err,
			}).Warn("Failed to delete cache key")
		}
	}

	return nil
}

// Warmup 缓存预热
func (cm *CacheManager) Warmup(ctx context.Context, warmupFunc func(key string) (interface{}, error)) error {
	if !cm.strategy.WarmupEnabled {
		return nil
	}

	cm.logger.Info("Starting cache warmup")

	for _, key := range cm.strategy.WarmupKeys {
		value, err := warmupFunc(key)
		if err != nil {
			cm.logger.WithFields(logrus.Fields{
				"key":   key,
				"error": err,
			}).Warn("Failed to warmup cache key")
			continue
		}

		if err := cm.SetWithStrategy(ctx, key, value); err != nil {
			cm.logger.WithFields(logrus.Fields{
				"key":   key,
				"error": err,
			}).Warn("Failed to set warmup cache")
		} else {
			cm.logger.WithField("key", key).Debug("Cache warmed up")
		}
	}

	cm.logger.Info("Cache warmup completed")
	return nil
}

// getTTLForKey 根据键获取TTL
func (cm *CacheManager) getTTLForKey(key string) time.Duration {
	// 检查特定类型的TTL
	for keyType, ttl := range cm.strategy.TypeTTLs {
		if len(key) > len(keyType) && key[:len(keyType)] == keyType {
			return ttl
		}
	}

	return cm.strategy.DefaultTTL
}

// CacheOptions 缓存选项
type CacheOptions struct {
	TTL       time.Duration
	Tags      []string
	Namespace string
}

// TaggedCache 带标签的缓存
type TaggedCache struct {
	manager *CacheManager
	logger  *logrus.Logger
}

// NewTaggedCache 创建带标签的缓存
func NewTaggedCache(manager *CacheManager, logger *logrus.Logger) *TaggedCache {
	return &TaggedCache{
		manager: manager,
		logger:  logger,
	}
}

// SetWithTags 设置带标签的缓存
func (tc *TaggedCache) SetWithTags(ctx context.Context, key string, value interface{}, opts *CacheOptions) error {
	if opts == nil {
		opts = &CacheOptions{TTL: tc.manager.strategy.DefaultTTL}
	}

	// 设置主缓存
	err := tc.manager.cache.Set(ctx, key, value, opts.TTL)
	if err != nil {
		return err
	}

	// 设置标签映射
	for _, tag := range opts.Tags {
		tagKey := fmt.Sprintf("tag:%s:%s", tag, key)
		if err := tc.manager.cache.Set(ctx, tagKey, "1", opts.TTL); err != nil {
			tc.logger.WithFields(logrus.Fields{
				"tag":   tag,
				"key":   key,
				"error": err,
			}).Warn("Failed to set tag mapping")
		}
	}

	return nil
}

// InvalidateByTag 根据标签失效缓存
func (tc *TaggedCache) InvalidateByTag(ctx context.Context, tag string) error {
	pattern := fmt.Sprintf("tag:%s:*", tag)
	tagKeys, err := tc.manager.cache.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	for _, tagKey := range tagKeys {
		// 提取原始键名
		parts := strings.Split(tagKey, ":")
		if len(parts) >= 3 {
			originalKey := strings.Join(parts[2:], ":")

			// 删除原始缓存
			if err := tc.manager.cache.Delete(ctx, originalKey); err != nil {
				tc.logger.WithFields(logrus.Fields{
					"key":   originalKey,
					"error": err,
				}).Warn("Failed to delete cache by tag")
			}

			// 删除标签映射
			if err := tc.manager.cache.Delete(ctx, tagKey); err != nil {
				tc.logger.WithFields(logrus.Fields{
					"tag_key": tagKey,
					"error":   err,
				}).Warn("Failed to delete tag mapping")
			}
		}
	}

	return nil
}

// 错误定义
var (
	ErrCacheNotFound = fmt.Errorf("cache not found")
	ErrCacheExpired  = fmt.Errorf("cache expired")
)

// DefaultStrategy 返回默认缓存策略
func DefaultStrategy() *CacheStrategy {
	return &CacheStrategy{
		DefaultTTL:     15 * time.Minute,
		MaxTTL:         24 * time.Hour,
		EnableMetrics:  true,
		WarmupEnabled:  false,
		EvictionPolicy: "LRU",
		TypeTTLs: map[string]time.Duration{
			"user:":    30 * time.Minute, // 用户信息缓存30分钟
			"session:": 2 * time.Hour,    // 会话缓存2小时
			"config:":  1 * time.Hour,    // 配置缓存1小时
			"auth:":    15 * time.Minute, // 认证缓存15分钟
		},
	}
}
