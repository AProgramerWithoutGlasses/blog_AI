package redis_utils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"siwuai/internal/infrastructure/config"
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
		return fmt.Errorf("r.client.Set() err: %v", err)
	}
	return nil
}

// Del 删除缓存值
func (r *RedisClient) Del(key string) error {
	if err := r.client.Del(r.ctx, key).Err(); err != nil {
		return fmt.Errorf("r.client.Del() err: %v", err)
	}
	return nil
}

// TryLock 尝试获取分布式锁
func (r *RedisClient) TryLock(key string, expiration time.Duration) (bool, error) {
	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second) // 设置 5 秒超时
	defer cancel()

	// 使用 SETNX 尝试设置锁
	result, err := r.client.SetNX(ctx, "lock:"+key, "locked", expiration).Result()
	if err != nil {
		return false, fmt.Errorf("redis SetNX 失败: %v", err)
	}

	if result {
		fmt.Println("加锁成功: ", "lock:"+key)
	} else {
		fmt.Println("加锁失败，锁已存在，开始轮询: ", "lock "+key)
	}

	return result, nil
}

// Unlock 释放分布式锁
func (r *RedisClient) Unlock(key string) error {
	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second) // 设置 5 秒超时
	defer cancel()

	if err := r.client.Del(ctx, "lock:"+key).Err(); err != nil {
		return fmt.Errorf("r.client.Del(ctx, key) err: %v", err)
	}

	fmt.Println("解锁成功: ", "lock:"+key)
	return nil
}

// Close 关闭 Redis 连接
func (r *RedisClient) Close() error {
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("关闭 Redis 连接失败: %v", err)
	}
	return nil
}
