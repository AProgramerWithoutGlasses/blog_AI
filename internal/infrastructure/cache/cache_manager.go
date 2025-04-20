package cache

import (
	"encoding/json"
	"fmt"
	"siwuai/internal/infrastructure/constant"
	"siwuai/internal/infrastructure/utils"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"siwuai/internal/domain/model/entity"
	"siwuai/internal/infrastructure/config"
	"siwuai/internal/infrastructure/redis_utils"
)

// 不同类型数据的缓存过期时间
const (
	DefaultExpiration = 24 * time.Hour     // 默认过期时间
	CodeExpiration    = 48 * time.Hour     // 代码解释缓存时间
	ArticleExpiration = 72 * time.Hour     // 文章缓存时间
	HotDataExpiration = 7 * 24 * time.Hour // 热点数据缓存时间
)

// CacheType 缓存类型
//type CacheType string
//
//const (
//	CodeCache    CacheType = "code"    // 代码缓存
//	ArticleCache CacheType = "article" // 文章缓存
//)

type CacheManagerInterface interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte, cacheType constant.CacheType)
	Delete(key string) error
	Close() error
}

// CacheManager 多级缓存管理器
type CacheManager struct {
	localCache  LocalCache               // 本地缓存（一级缓存）
	redisClient *redis_utils.RedisClient // Redis缓存（二级缓存）
	//bloomFilter *bloom.BloomFilter                // 布隆过滤器
	db     *gorm.DB                          // 数据库连接
	config config.Config                     // 配置信息
	mu     sync.RWMutex                      // 读写锁
	jct    constant.JudgingCacheType         // 缓存类型
	bfm    utils.BloomFilterManagerInterface // 布隆过滤器管理器
}

// NewCacheManager 创建一个新的缓存管理器
func NewCacheManager(localCache LocalCache, db *gorm.DB, redisClient *redis_utils.RedisClient, cfg config.Config, jct constant.JudgingCacheType, bfm utils.BloomFilterManagerInterface) *CacheManager {
	// 创建本地缓存，设置1小时的默认过期时间
	//localCache, err := NewBigCacheClient(1*time.Hour, 1024*1024, 1024) // 1MB最大条目大小，1024个分片
	//if err != nil {
	//	return nil, fmt.Errorf("创建本地缓存失败: %v", err)
	//}

	cm := &CacheManager{
		localCache:  localCache,
		redisClient: redisClient,
		//bloomFilter: bf,
		db:     db,
		config: cfg,
		jct:    jct,
		bfm:    bfm,
	}

	// 启动缓存预热
	go cm.WarmUpCache()

	zap.L().Info("多级缓存管理器初始化成功")
	return cm
}

// Get 从缓存获取数据，优先从本地缓存获取，然后是Redis
// 当查询文章信息时，key表示articleID
func (cm *CacheManager) Get(key string) ([]byte, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	//// 1. 检查布隆过滤器
	//if !cm.bloomFilter.Test([]byte(key)) {
	//	zap.L().Debug("布隆过滤器未命中", zap.String("key", key))
	//	return nil, nil
	//}F

	// 1. 检查布隆过滤器
	if !cm.bfm.Test([]byte(key)) {
		zap.L().Debug("布隆过滤器未命中", zap.String("key", key))
		return nil, nil
	}

	zap.L().Debug("布隆过滤器命中", zap.String("key", key))

	// 2. 检查本地缓存
	data, err := cm.localCache.Get(key)
	if err == nil && data != nil {
		zap.L().Debug("本地缓存命中", zap.String("key", key))
		return data, nil
	}

	// 3. 检查Redis缓存
	redisData, err := cm.redisClient.Get(key)
	if err != nil || redisData == "" {
		zap.L().Debug("Redis缓存未命中", zap.String("key", key))
		return nil, nil
	}

	// Redis缓存命中，同步到本地缓存
	zap.L().Debug("Redis缓存命中，同步到本地缓存", zap.String("key", key))
	data = []byte(redisData)
	_ = cm.localCache.Set(key, data, 0) // 忽略本地缓存设置错误

	return data, nil
}

// Set 设置缓存，同时设置本地缓存和Redis缓存
func (cm *CacheManager) Set(key string, value []byte, cacheType constant.CacheType) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 根据缓存类型设置不同的过期时间
	expiration := cm.getExpirationByType(cacheType)

	// 1. 设置本地缓存
	if err := cm.localCache.Set(key, value, 0); err != nil {
		zap.L().Error("设置本地缓存失败", zap.Error(err), zap.String("key", key))
		// 本地缓存失败不影响Redis缓存
	}

	// 2. 设置Redis缓存
	if err := cm.redisClient.Set(key, string(value), expiration); err != nil {
		zap.L().Error("设置Redis缓存失败", zap.Error(err), zap.String("key", key))
		//return fmt.Errorf("设置Redis缓存失败: %v", err)
	}

	// 3. 更新布隆过滤器
	//cm.bloomFilter.Add([]byte(key))
	cm.bfm.Add([]byte(key))

	zap.L().Debug("缓存设置成功",
		zap.String("key", key),
		zap.String("type", string(cacheType)),
		zap.Duration("expiration", expiration),
		zap.String("data", string(value)),
	)
}

// Delete 删除缓存
func (cm *CacheManager) Delete(key string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 1. 删除本地缓存
	if err := cm.localCache.Delete(key); err != nil {
		zap.L().Error("删除本地缓存失败", zap.Error(err), zap.String("key", key))
		// 本地缓存删除失败不影响Redis缓存删除
	}

	// 2. 删除Redis缓存
	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	//defer cancel()

	if err := cm.redisClient.Del(key); err != nil {
		return fmt.Errorf("删除Redis缓存失败: %v", err)
	}

	return nil
}

// WarmUpCache 缓存预热，加载热点数据到缓存
func (cm *CacheManager) WarmUpCache() {
	zap.L().Info("开始缓存预热...")

	// 1. 加载热门代码解释
	//cm.warmUpCodes(cm.jct.GetCodeFlag())

	// 2. 加载热门文章
	cm.warmUpArticles(cm.jct.GetArticleFlag())

	zap.L().Info("缓存预热完成")
}

// warmUpCodes 预热代码解释缓存
func (cm *CacheManager) warmUpCodes(cacheType constant.CacheType) {
	// 查询访问量最高的前100条代码记录
	var codes []entity.Code
	if err := cm.db.Model(&entity.Code{}).Order("visit_count DESC").Limit(100).Find(&codes).Error; err != nil {
		zap.L().Error("加载热门代码记录失败", zap.Error(err))
		return
	}

	for _, code := range codes {
		// 序列化数据
		data, err := json.Marshal(code)
		if err != nil {
			zap.L().Error("序列化代码记录失败", zap.Error(err), zap.Uint("id", code.ID))
			continue
		}

		// 设置到缓存
		cm.Set(code.Key, data, cacheType)

		//err != nil {
		//	zap.L().Error("预热代码缓存失败", zap.Error(err), zap.String("key", code.Key))
		//}
	}

	zap.L().Info("代码缓存预热完成", zap.Int("count", len(codes)))
}

// warmUpArticles 预热文章缓存
func (cm *CacheManager) warmUpArticles(cacheType constant.CacheType) {
	// 查询访问量最高的前50篇文章
	var articles []entity.Article
	if err := cm.db.Model(&entity.Article{}).Order("visit_count DESC").Limit(50).Find(&articles).Error; err != nil {
		zap.L().Error("加载热门文章记录失败", zap.Error(err))
		return
	}

	for _, article := range articles {
		// 序列化数据
		data, err := json.Marshal(article)
		if err != nil {
			zap.L().Error("序列化文章记录失败", zap.Error(err), zap.Uint("id", article.ID))
			continue
		}

		// 设置到缓存
		cm.Set(fmt.Sprintf("article:%d", article.ArticleID), data, cacheType)

		//err != nil {
		//	zap.L().Error("预热文章缓存失败", zap.Error(err), zap.Uint("id", article.ID))
		//}
	}

	zap.L().Info("文章缓存预热完成", zap.Int("count", len(articles)))
}

// getExpirationByType 根据缓存类型获取过期时间
func (cm *CacheManager) getExpirationByType(cacheType constant.CacheType) time.Duration {
	switch cacheType {
	case cm.jct.GetCodeFlag():
		return CodeExpiration
	case cm.jct.GetArticleFlag():
		return ArticleExpiration
	default:
		return DefaultExpiration
	}
}

// Close 关闭缓存管理器
func (cm *CacheManager) Close() error {
	// 关闭本地缓存
	if err := cm.localCache.Close(); err != nil {
		zap.L().Error("关闭本地缓存失败", zap.Error(err))
	}

	return nil
}
