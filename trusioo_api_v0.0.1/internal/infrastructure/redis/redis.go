package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"trusioo_api_v0.0.1/internal/config"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Client Redis客户端结构
type Client struct {
	*redis.Client
	config *config.RedisConfig
	logger *logrus.Logger
}

// New 创建新的Redis客户端
func New(cfg *config.RedisConfig, logger *logrus.Logger) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connection established successfully")

	return &Client{
		Client: rdb,
		config: cfg,
		logger: logger,
	}, nil
}

// Close 关闭Redis连接
func (c *Client) Close() error {
	if c.Client != nil {
		if err := c.Client.Close(); err != nil {
			c.logger.WithError(err).Error("Error closing Redis connection")
			return err
		}
		c.logger.Info("Redis connection closed")
	}
	return nil
}

// Health 检查Redis健康状态
func (c *Client) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

// SetJSON 设置JSON对象到Redis
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := c.Set(ctx, key, data, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set JSON to Redis: %w", err)
	}

	return nil
}

// GetJSON 从Redis获取JSON对象
func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key %s not found", key)
		}
		return fmt.Errorf("failed to get JSON from Redis: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}

// SetHash 设置哈希字段
func (c *Client) SetHash(ctx context.Context, key string, field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal hash value: %w", err)
	}

	if err := c.HSet(ctx, key, field, data).Err(); err != nil {
		return fmt.Errorf("failed to set hash field: %w", err)
	}

	return nil
}

// GetHash 获取哈希字段
func (c *Client) GetHash(ctx context.Context, key string, field string, dest interface{}) error {
	data, err := c.HGet(ctx, key, field).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("hash field %s:%s not found", key, field)
		}
		return fmt.Errorf("failed to get hash field: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal hash value: %w", err)
	}

	return nil
}

// SetWithRetry 带重试机制的设置操作
func (c *Client) SetWithRetry(ctx context.Context, key string, value interface{}, expiration time.Duration, maxRetries int) error {
	var err error
	for i := 0; i <= maxRetries; i++ {
		if err = c.Set(ctx, key, value, expiration).Err(); err == nil {
			return nil
		}

		if i < maxRetries {
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
			c.logger.WithError(err).Warnf("Redis set retry %d/%d for key %s", i+1, maxRetries, key)
		}
	}
	return fmt.Errorf("failed to set after %d retries: %w", maxRetries, err)
}

// GetWithRetry 带重试机制的获取操作
func (c *Client) GetWithRetry(ctx context.Context, key string, maxRetries int) (string, error) {
	var result string
	var err error

	for i := 0; i <= maxRetries; i++ {
		result, err = c.Get(ctx, key).Result()
		if err == nil {
			return result, nil
		}

		if err == redis.Nil {
			return "", fmt.Errorf("key %s not found", key)
		}

		if i < maxRetries {
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
			c.logger.WithError(err).Warnf("Redis get retry %d/%d for key %s", i+1, maxRetries, key)
		}
	}
	return "", fmt.Errorf("failed to get after %d retries: %w", maxRetries, err)
}

// Lock 分布式锁
type Lock struct {
	client *Client
	key    string
	value  string
	ttl    time.Duration
}

// AcquireLock 获取分布式锁
func (c *Client) AcquireLock(ctx context.Context, key string, ttl time.Duration) (*Lock, error) {
	value := fmt.Sprintf("%d", time.Now().UnixNano())

	result, err := c.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !result {
		return nil, fmt.Errorf("lock already exists for key: %s", key)
	}

	return &Lock{
		client: c,
		key:    key,
		value:  value,
		ttl:    ttl,
	}, nil
}

// Release 释放锁
func (l *Lock) Release(ctx context.Context) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if result.(int64) == 0 {
		return fmt.Errorf("lock not found or not owned")
	}

	return nil
}

// Extend 延长锁的过期时间
func (l *Lock) Extend(ctx context.Context, newTTL time.Duration) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value, int(newTTL.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("failed to extend lock: %w", err)
	}

	if result.(int64) == 0 {
		return fmt.Errorf("lock not found or not owned")
	}

	l.ttl = newTTL
	return nil
}

// Cache 缓存管理器
type Cache struct {
	client *Client
	prefix string
}

// NewCache 创建缓存管理器
func (c *Client) NewCache(prefix string) *Cache {
	return &Cache{
		client: c,
		prefix: prefix,
	}
}

// buildKey 构建带前缀的键名
func (cache *Cache) buildKey(key string) string {
	if cache.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", cache.prefix, key)
}

// Set 设置缓存
func (cache *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return cache.client.SetJSON(ctx, cache.buildKey(key), value, expiration)
}

// Get 获取缓存
func (cache *Cache) Get(ctx context.Context, key string, dest interface{}) error {
	return cache.client.GetJSON(ctx, cache.buildKey(key), dest)
}

// Delete 删除缓存
func (cache *Cache) Delete(ctx context.Context, key string) error {
	return cache.client.Del(ctx, cache.buildKey(key)).Err()
}

// Exists 检查缓存是否存在
func (cache *Cache) Exists(ctx context.Context, key string) (bool, error) {
	result, err := cache.client.Exists(ctx, cache.buildKey(key)).Result()
	return result > 0, err
}

// TTL 获取缓存剩余时间
func (cache *Cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return cache.client.TTL(ctx, cache.buildKey(key)).Result()
}

// Expire 设置缓存过期时间
func (cache *Cache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return cache.client.Expire(ctx, cache.buildKey(key), expiration).Err()
}

// Clear 清除所有带前缀的缓存
func (cache *Cache) Clear(ctx context.Context) error {
	pattern := cache.buildKey("*")
	keys, err := cache.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return cache.client.Del(ctx, keys...).Err()
	}

	return nil
}
