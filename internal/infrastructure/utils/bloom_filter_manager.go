package utils

import (
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"math"
	"siwuai/internal/domain/model/entity"
	"sync"
	"time"
)

type BloomFilterManagerInterface interface {
	Test(data []byte) bool
	Add(data []byte)
	GetBloomFilter() *bloom.BloomFilter
}

// BloomFilterManager 布隆过滤器管理器
// 负责布隆过滤器的动态参数调整和定期重建
type BloomFilterManager struct {
	db              *gorm.DB
	bloomFilter     *bloom.BloomFilter
	mutex           sync.RWMutex
	lastRebuildTime time.Time
	config          BloomFilterConfig
}

// BloomFilterConfig 布隆过滤器配置
type BloomFilterConfig struct {
	// 预估元素数量
	EstimatedElements uint
	// 期望的误判率 (0.0 到 1.0)
	FalsePositiveRate float64
	// 重建间隔时间
	RebuildInterval time.Duration
	// 重建阈值（当元素数量超过预估的百分比时触发重建）
	RebuildThreshold float64
}

// NewBloomFilterManager 创建一个新的布隆过滤器管理器
func NewBloomFilterManager(db *gorm.DB, config BloomFilterConfig) *BloomFilterManager {
	// 设置默认值
	if config.EstimatedElements == 0 {
		config.EstimatedElements = 100000
	}
	if config.FalsePositiveRate <= 0 || config.FalsePositiveRate >= 1 {
		config.FalsePositiveRate = 0.01
	}
	if config.RebuildInterval == 0 {
		config.RebuildInterval = 24 * time.Hour // 默认每天重建一次
	}
	if config.RebuildThreshold <= 0 || config.RebuildThreshold >= 1 {
		config.RebuildThreshold = 0.8 // 默认当元素数量达到预估的80%时重建
	}

	// 计算最优参数
	m, k := calculateOptimalParameters(config.EstimatedElements, config.FalsePositiveRate)

	// 创建布隆过滤器
	bf := bloom.New(m, k)

	manager := &BloomFilterManager{
		db:              db,
		bloomFilter:     bf,
		lastRebuildTime: time.Now(),
		config:          config,
	}

	// 初始化布隆过滤器
	manager.initBloomFilter()

	// 启动定期重建协程
	go manager.scheduleRebuild()

	return manager
}

// calculateOptimalParameters 计算布隆过滤器的最优参数
// n: 预估元素数量
// p: 期望的误判率
// 返回: m (位数组大小), k (哈希函数数量)
func calculateOptimalParameters(n uint, p float64) (uint, uint) {
	// 计算最优位数组大小 m
	m := uint(math.Ceil(-float64(n) * math.Log(p) / math.Pow(math.Log(2), 2)))

	// 计算最优哈希函数数量 k
	k := uint(math.Ceil(float64(m) / float64(n) * math.Log(2)))

	// 确保 k 至少为 1
	if k < 1 {
		k = 1
	}

	return m, k
}

// initBloomFilter 初始化布隆过滤器
func (b *BloomFilterManager) initBloomFilter() {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// 从数据库加载数据
	var codes []entity.Code
	if err := b.db.Find(&codes).Error; err != nil {
		zap.L().Error("加载布隆过滤器数据(code)失败", zap.Error(err))
		return
	}

	var articles []entity.Article
	if err := b.db.Find(&articles).Error; err != nil {
		zap.L().Error("加载布隆过滤器数据(article)失败", zap.Error(err))
	}

	// 清空布隆过滤器
	b.bloomFilter.ClearAll()

	// 填充布隆过滤器
	for _, code := range codes {
		b.bloomFilter.Add([]byte(code.Key))
	}

	for _, article := range articles {
		b.bloomFilter.Add([]byte(fmt.Sprintf("article:%d", article.ArticleID)))
	}

	// 更新重建时间
	b.lastRebuildTime = time.Now()

	msg := fmt.Sprintf("布隆过滤器初始化完成，共加载 %d 条记录", len(codes)+len(articles))
	fmt.Println(msg)
	zap.L().Info(msg)
}

// scheduleRebuild 定期重建布隆过滤器
func (b *BloomFilterManager) scheduleRebuild() {
	ticker := time.NewTicker(1 * time.Hour) // 每小时检查一次是否需要重建
	defer ticker.Stop()

	for {
		<-ticker.C

		// 检查是否需要重建
		if b.shouldRebuild() {
			zap.L().Info("开始重建布隆过滤器")
			b.rebuildBloomFilter()
			zap.L().Info("布隆过滤器重建完成")
		}
	}
}

// shouldRebuild 判断是否需要重建布隆过滤器
func (b *BloomFilterManager) shouldRebuild() bool {
	// 检查时间间隔
	timeThreshold := time.Since(b.lastRebuildTime) >= b.config.RebuildInterval

	// 检查元素数量
	b.mutex.RLock()
	approxCount := b.bloomFilter.ApproximatedSize()
	b.mutex.RUnlock()

	countThreshold := float64(approxCount) >= float64(b.config.EstimatedElements)*b.config.RebuildThreshold

	return timeThreshold || countThreshold
}

// rebuildBloomFilter 重建布隆过滤器
func (b *BloomFilterManager) rebuildBloomFilter() {
	// 获取当前数据库中的记录数量
	var codeCount int64
	if err := b.db.Model(&entity.Code{}).Count(&codeCount).Error; err != nil {
		zap.L().Error("获取数据库记录数量(code)失败", zap.Error(err))
		return
	}

	var articleCount int64
	if err := b.db.Model(&entity.Article{}).Count(&articleCount).Error; err != nil {
		zap.L().Error("获取数据库记录数量(article)失败", zap.Error(err))
		return
	}

	count := codeCount + articleCount

	// 如果数据库记录数量发生变化，重新计算参数
	if uint(count) > b.config.EstimatedElements {
		// 更新预估元素数量，增加50%的余量
		newEstimatedElements := uint(float64(count) * 1.5)

		// 计算新的最优参数
		m, k := calculateOptimalParameters(newEstimatedElements, b.config.FalsePositiveRate)

		// 创建新的布隆过滤器
		newBF := bloom.New(m, k)

		// 更新配置
		b.mutex.Lock()
		b.config.EstimatedElements = newEstimatedElements
		b.bloomFilter = newBF
		b.mutex.Unlock()

		zap.L().Info("布隆过滤器参数已更新",
			zap.Uint("新预估元素数量", newEstimatedElements),
			zap.Uint("位数组大小", m),
			zap.Uint("哈希函数数量", k))
	}

	// 重新初始化布隆过滤器
	b.initBloomFilter()
}

// Test 测试元素是否在布隆过滤器中
func (b *BloomFilterManager) Test(data []byte) bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.bloomFilter.Test(data)
}

// Add 添加元素到布隆过滤器
func (b *BloomFilterManager) Add(data []byte) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.bloomFilter.Add(data)

	// 检查是否需要重建
	if b.bloomFilter.ApproximatedSize() >= uint32(float64(b.config.EstimatedElements)*b.config.RebuildThreshold) {
		go b.rebuildBloomFilter() // 异步重建，不阻塞当前操作
	}
}

// GetBloomFilter 获取当前的布隆过滤器
func (b *BloomFilterManager) GetBloomFilter() *bloom.BloomFilter {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.bloomFilter
}
