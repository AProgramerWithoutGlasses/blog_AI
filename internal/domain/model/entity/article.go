package entity

import (
	"gorm.io/gorm"
	"siwuai/internal/domain/model/dto"
)

type Article struct {
	gorm.Model
	Key       string `json:"key"`        // 用于标识文章的状态(是否被修改)
	ArticleID uint   `json:"article_id"` // 文章ID
	Abstract  string `json:"abstract"`   // 发布文章时，提取的文章摘要
	Summary   string `json:"summary"`    // 发布文章时，提取的文章总结
}

func (*Article) TableName() string {
	return ""
}

func (a *Article) ConvertArticleEntityToDto() *dto.Article {
	return &dto.Article{
		Key:       a.Key,
		ArticleID: a.ArticleID,
		Abstract:  a.Abstract,
		Summary:   a.Summary,
	}
}

func ConvertArticleDtoToEntity(article *dto.Article) *Article {
	return &Article{
		Key:       article.Key,
		ArticleID: article.ArticleID,
		Abstract:  article.Abstract,
		Summary:   article.Summary,
	}
}
