package entity

import "gorm.io/gorm"

type Article struct {
	gorm.Model
	Key       int    `json:"key"`        // 用于标识文章的状态(是否被修改)
	ArticleID int    `json:"article_id"` // 文章ID
	Abstract  string `json:"abstract"`   // 发布文章时，提取的文章摘要
	Summary   string `json:"summary"`    // 发布文章时，提取的文章总结
}
