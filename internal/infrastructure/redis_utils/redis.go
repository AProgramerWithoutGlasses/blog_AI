package redis_utils

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	"siwuai/internal/infrastructure/config"
	"time"
)

// RedisClient 封装 Redis 客户端和配置
type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// Config Redis 配置
type Config struct {
	Addr     string        // Redis 地址，例如 "localhost:6379"
	Password string        // Redis 密码（可选）
	DB       int           // 数据库编号，默认 0
	Timeout  time.Duration // 操作超时时间
}

// NewRedisClient 初始化 Redis 客户端
func NewRedisClient(cfg config.Config) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx := context.Background()
	// 测试连接
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("连接 Redis 失败: %v", err)
	}

	fmt.Println("redis 连接成功")
	return &RedisClient{
		client: client,
		ctx:    ctx,
	}, nil
}

// Get 获取缓存值
func (r *RedisClient) Get(key string) (string, error) {
	val, err := r.client.Get(r.ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil // 缓存未命中，返回空字符串和 nil 错误
	}
	if err != nil {
		return "", fmt.Errorf("redis Get 失败: %v", err)
	}
	return val, nil
}

// Set 设置缓存值
func (r *RedisClient) Set(key, value string, expiration time.Duration) error {
	if err := r.client.Set(r.ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("redis Set 失败: %v", err)
	}
	return nil
}

// Close 关闭 Redis 连接
func (r *RedisClient) Close() error {
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("关闭 Redis 连接失败: %v", err)
	}
	return nil
}
