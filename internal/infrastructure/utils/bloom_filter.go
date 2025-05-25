package utils

import (
	"fmt"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

// 全局布隆过滤器管理器实例
//var bloomFilterManager *BloomFilterManager

// LoadBloomFilter 加载布隆过滤器，从数据库初始化
// 使用布隆过滤器管理器实现动态参数调整和定期重建
func LoadBloomFilter(db *gorm.DB) (bfm *BloomFilterManager, err error) {
	// 创建布隆过滤器配置
	config := BloomFilterConfig{
		EstimatedElements: 100000,         // 初始预估元素数量
		FalsePositiveRate: 0.01,           // 期望的误判率为1%
		RebuildInterval:   24 * time.Hour, // 每24小时重建一次
		RebuildThreshold:  0.8,            // 当元素数量达到预估的80%时触发重建
	}

	//// 创建布隆过滤器管理器
	//bloomFilterManager = NewBloomFilterManager(db, config)
	//
	//// 获取布隆过滤器实例
	//bf = bloomFilterManager.GetBloomFilter()
	bfm = NewBloomFilterManager(db, config)

	msg := fmt.Sprintf("BloomFilter 加载完成，使用动态参数调整和定期重建机制")
	fmt.Println(msg)
	zap.L().Info(msg)

	return
}

//// GetBloomFilterManager 获取布隆过滤器管理器实例
//func GetBloomFilterManager() *BloomFilterManager {
//	return bloomFilterManager
//}
