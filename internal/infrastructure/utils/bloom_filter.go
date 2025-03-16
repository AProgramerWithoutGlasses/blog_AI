package utils

import (
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"gorm.io/gorm"
	"siwuai/internal/domain/model/entity"
)

// LoadBloomFilter 加载布隆过滤器，从数据库初始化
func LoadBloomFilter(db *gorm.DB) (bf *bloom.BloomFilter, err error) {
	bf = bloom.New(100000, 5) // 假设问题数量为100000，误判率为5%

	var codes []entity.Code
	if err = db.Find(&codes).Error; err != nil {
		fmt.Printf("加载布隆过滤器失败: %v\n", err)
		return
	}
	for _, code := range codes {
		// 使用问题的哈希值填充布隆过滤器
		bf.Add([]byte(code.Key))
	}
	fmt.Println("布隆过滤器加载完成")

	return
}
