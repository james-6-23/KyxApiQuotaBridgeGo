package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// Redis Redis客户端封装
type Redis struct {
	client *redis.Client
	logger *logrus.Logger
}

// Config Redis配置
type Config struct {
	Host       string
	Port       int
	Password   string
	DB         int
	PoolSize   int
	MaxRetries int
}

// New 创建新的Redis客户端
func New(cfg *Config, logger *logrus.Logger) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolTimeout:  4 * time.Second,
		IdleTimeout:  5 * time.Minute,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"host": cfg.Host,
		"port": cfg.Port,
		"db":   cfg.DB,
	}).Info("Redis connected successfully")

	return &Redis{
		client: client,
		logger: logger,
	}, nil
}

// Close 关闭Redis连接
func (r *Redis) Close() error {
	r.logger.Info("Closing Redis connection")
	return r.client.Close()
}

// HealthCheck 健康检查
func (r *Redis) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}

	return nil
}

// Get 获取字符串值
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return val, nil
}

// Set 设置字符串值
func (r *Redis) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if err := r.client.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

// GetJSON 获取JSON值
func (r *Redis) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := r.Get(ctx, key)
	if err != nil {
		return err
	}
	if val == "" {
		return redis.Nil
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal json for key %s: %w", key, err)
	}
	return nil
}

// SetJSON 设置JSON值
func (r *Redis) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal json for key %s: %w", key, err)
	}
	return r.Set(ctx, key, data, expiration)
}

// Del 删除键
func (r *Redis) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}
	return nil
}

// Exists 检查键是否存在
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}
	return n > 0, nil
}

// Expire 设置键过期时间
func (r *Redis) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if err := r.client.Expire(ctx, key, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set expiration for key %s: %w", key, err)
	}
	return nil
}

// TTL 获取键的剩余生存时间
func (r *Redis) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get ttl for key %s: %w", key, err)
	}
	return ttl, nil
}

// Incr 递增
func (r *Redis) Incr(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return val, nil
}

// IncrBy 递增指定值
func (r *Redis) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	val, err := r.client.IncrBy(ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s by %d: %w", key, value, err)
	}
	return val, nil
}

// Decr 递减
func (r *Redis) Decr(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s: %w", key, err)
	}
	return val, nil
}

// DecrBy 递减指定值
func (r *Redis) DecrBy(ctx context.Context, key string, value int64) (int64, error) {
	val, err := r.client.DecrBy(ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s by %d: %w", key, value, err)
	}
	return val, nil
}

// SAdd 添加成员到集合
func (r *Redis) SAdd(ctx context.Context, key string, members ...interface{}) error {
	if err := r.client.SAdd(ctx, key, members...).Err(); err != nil {
		return fmt.Errorf("failed to add members to set %s: %w", key, err)
	}
	return nil
}

// SIsMember 检查是否是集合成员
func (r *Redis) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	result, err := r.client.SIsMember(ctx, key, member).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check membership in set %s: %w", key, err)
	}
	return result, nil
}

// SMembers 获取集合所有成员
func (r *Redis) SMembers(ctx context.Context, key string) ([]string, error) {
	members, err := r.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get members from set %s: %w", key, err)
	}
	return members, nil
}

// SCard 获取集合成员数量
func (r *Redis) SCard(ctx context.Context, key string) (int64, error) {
	count, err := r.client.SCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get cardinality of set %s: %w", key, err)
	}
	return count, nil
}

// HSet 设置哈希字段
func (r *Redis) HSet(ctx context.Context, key string, field string, value interface{}) error {
	if err := r.client.HSet(ctx, key, field, value).Err(); err != nil {
		return fmt.Errorf("failed to set hash field %s:%s: %w", key, field, err)
	}
	return nil
}

// HGet 获取哈希字段
func (r *Redis) HGet(ctx context.Context, key string, field string) (string, error) {
	val, err := r.client.HGet(ctx, key, field).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get hash field %s:%s: %w", key, field, err)
	}
	return val, nil
}

// HGetAll 获取哈希所有字段
func (r *Redis) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	result, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get all hash fields for key %s: %w", key, err)
	}
	return result, nil
}

// HDel 删除哈希字段
func (r *Redis) HDel(ctx context.Context, key string, fields ...string) error {
	if err := r.client.HDel(ctx, key, fields...).Err(); err != nil {
		return fmt.Errorf("failed to delete hash fields from key %s: %w", key, err)
	}
	return nil
}

// Pipeline 执行管道操作
func (r *Redis) Pipeline(ctx context.Context, fn func(redis.Pipeliner) error) error {
	pipe := r.client.Pipeline()
	if err := fn(pipe); err != nil {
		return err
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to execute pipeline: %w", err)
	}
	return nil
}

// Keys 获取匹配的键列表（谨慎使用）
func (r *Redis) Keys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get keys with pattern %s: %w", pattern, err)
	}
	return keys, nil
}

// FlushDB 清空当前数据库（危险操作）
func (r *Redis) FlushDB(ctx context.Context) error {
	if err := r.client.FlushDB(ctx).Err(); err != nil {
		return fmt.Errorf("failed to flush database: %w", err)
	}
	r.logger.Warn("Redis database flushed")
	return nil
}

// GetClient 获取原始Redis客户端（用于高级操作）
func (r *Redis) GetClient() *redis.Client {
	return r.client
}

// GetStats 获取Redis统计信息
func (r *Redis) GetStats(ctx context.Context) (map[string]interface{}, error) {
	poolStats := r.client.PoolStats()
	info := r.client.Info(ctx, "stats", "memory")

	stats := map[string]interface{}{
		"pool": map[string]interface{}{
			"hits":        poolStats.Hits,
			"misses":      poolStats.Misses,
			"timeouts":    poolStats.Timeouts,
			"total_conns": poolStats.TotalConns,
			"idle_conns":  poolStats.IdleConns,
			"stale_conns": poolStats.StaleConns,
		},
	}

	if info.Err() == nil {
		stats["info"] = info.Val()
	}

	return stats, nil
}

// SetNX 仅当键不存在时设置
func (r *Redis) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	result, err := r.client.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to setnx key %s: %w", key, err)
	}
	return result, nil
}

// GetSet 设置新值并返回旧值
func (r *Redis) GetSet(ctx context.Context, key string, value interface{}) (string, error) {
	oldVal, err := r.client.GetSet(ctx, key, value).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to getset key %s: %w", key, err)
	}
	return oldVal, nil
}

// MGet 批量获取
func (r *Redis) MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	if len(keys) == 0 {
		return []interface{}{}, nil
	}
	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to mget keys: %w", err)
	}
	return vals, nil
}

// MSet 批量设置
func (r *Redis) MSet(ctx context.Context, pairs ...interface{}) error {
	if len(pairs) == 0 {
		return nil
	}
	if err := r.client.MSet(ctx, pairs...).Err(); err != nil {
		return fmt.Errorf("failed to mset: %w", err)
	}
	return nil
}

// LPush 从左侧插入列表
func (r *Redis) LPush(ctx context.Context, key string, values ...interface{}) error {
	if err := r.client.LPush(ctx, key, values...).Err(); err != nil {
		return fmt.Errorf("failed to lpush to key %s: %w", key, err)
	}
	return nil
}

// RPush 从右侧插入列表
func (r *Redis) RPush(ctx context.Context, key string, values ...interface{}) error {
	if err := r.client.RPush(ctx, key, values...).Err(); err != nil {
		return fmt.Errorf("failed to rpush to key %s: %w", key, err)
	}
	return nil
}

// LPop 从左侧弹出列表元素
func (r *Redis) LPop(ctx context.Context, key string) (string, error) {
	val, err := r.client.LPop(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to lpop from key %s: %w", key, err)
	}
	return val, nil
}

// RPop 从右侧弹出列表元素
func (r *Redis) RPop(ctx context.Context, key string) (string, error) {
	val, err := r.client.RPop(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to rpop from key %s: %w", key, err)
	}
	return val, nil
}

// LRange 获取列表范围内的元素
func (r *Redis) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	vals, err := r.client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to lrange key %s: %w", key, err)
	}
	return vals, nil
}

// LLen 获取列表长度
func (r *Redis) LLen(ctx context.Context, key string) (int64, error) {
	length, err := r.client.LLen(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get length of key %s: %w", key, err)
	}
	return length, nil
}

// IsNil 检查错误是否是 redis.Nil
func IsNil(err error) bool {
	return err == redis.Nil
}

// Ping 检查Redis连接
func (r *Redis) Ping(ctx context.Context) error {
	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	return nil
}

// BloomAdd 向布隆过滤器添加元素（使用Set实现简单版本）
// 注意：这不是真正的布隆过滤器，如果需要真正的布隆过滤器，请安装RedisBloom模块
func (r *Redis) BloomAdd(ctx context.Context, key string, member string) error {
	if err := r.client.SAdd(ctx, key, member).Err(); err != nil {
		return fmt.Errorf("failed to add to bloom filter %s: %w", key, err)
	}
	return nil
}

// BloomExists 检查元素是否可能存在于布隆过滤器（使用Set实现简单版本）
// 注意：这不是真正的布隆过滤器，如果需要真正的布隆过滤器，请安装RedisBloom模块
func (r *Redis) BloomExists(ctx context.Context, key string, member string) (bool, error) {
	result, err := r.client.SIsMember(ctx, key, member).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check bloom filter %s: %w", key, err)
	}
	return result, nil
}
