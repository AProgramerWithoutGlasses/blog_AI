package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/allegro/bigcache/v3"
	"go.uber.org/zap"
)

// LocalCache 本地内存缓存接口
type LocalCache interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, expiration time.Duration) error
	Delete(key string) error
	Close() error
}

// BigCacheClient 使用BigCache实现的本地缓存
type BigCacheClient struct {
	cache *bigcache.BigCache
}

// NewBigCacheClient 创建一个新的BigCache客户端
// evictionTime time.Duration：缓存条目的全局淘汰时间（条目保留的时长）。
// maxEntrySize int：单个缓存条目的最大字节大小。
// shards int：缓存的分片数，影响并发性能和内存使用。
func NewBigCacheClient(evictionTime time.Duration, maxEntrySize int, shards int) *BigCacheClient {
	config := bigcache.DefaultConfig(evictionTime)
	config.MaxEntriesInWindow = 50000  // 设置清理窗口的最大条目数，限制内存使用
	config.MaxEntrySize = maxEntrySize // 设置单个条目最大大小来自函数参数）
	config.Shards = shards             // 设置分片数（来自函数参数），分片数越多并发性能越好，但内存开销增加
	config.Verbose = false             // 禁用 bigcache 的详细日志，减少日志输出
	//config.CleanWindow = 5 * time.Minute // 设置清理窗口时间

	cache, err := bigcache.New(context.Background(), config)
	if err != nil {
		zap.L().Error("创建BigCache失败", zap.Error(err))
		//return nil, fmt.Errorf("创建BigCache失败: %v", err)
	}

	zap.L().Info("本地缓存(BigCache)初始化成功")
	return &BigCacheClient{cache: cache}
}

// Get 从本地缓存获取值
func (c *BigCacheClient) Get(key string) ([]byte, error) {
	val, err := c.cache.Get(key)
	if err == bigcache.ErrEntryNotFound {
		return nil, nil // 缓存未命中，返回nil
	}
	if err != nil {
		return nil, fmt.Errorf("BigCache Get失败: %v", err)
	}
	return val, nil
}

// Set 设置本地缓存值
func (c *BigCacheClient) Set(key string, value []byte, _ time.Duration) error {
	// BigCache不支持单独设置每个key的过期时间，过期时间在创建时全局设置
	if err := c.cache.Set(key, value); err != nil {
		return fmt.Errorf("BigCache Set失败: %v", err)
	}
	return nil
}

// Delete 删除本地缓存中的键
func (c *BigCacheClient) Delete(key string) error {
	return c.cache.Delete(key)
}

// Close 关闭本地缓存
func (c *BigCacheClient) Close() error {
	return c.cache.Close()
}
