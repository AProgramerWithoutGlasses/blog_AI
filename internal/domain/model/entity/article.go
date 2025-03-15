package entity

import (
	"gorm.io/gorm"
	"siwuai/internal/domain/model/dto"
)

type Article struct {
	gorm.Model
	Key       string `gorm:"key"`        // 用于标识文章的状态(是否被修改)
	ArticleID uint   `gorm:"article_id"` // 文章ID
	Abstract  string `gorm:"abstract"`   // 发布文章时，提取的文章摘要
	Summary   string `gorm:"summary"`    // 发布文章时，提取的文章总结
}

//func (*ArticleFirst) TableName() string {
//	return ""
//}

func (a *Article) ConvertArticleEntityToDtoFirst() *dto.ArticleFirst {
	return &dto.ArticleFirst{
		Abstract: a.Abstract,
		Summary:  a.Summary,
	}
}

func (a *Article) ConvertArticleEntityToDtoSecond() *dto.ArticleSecond {
	return &dto.ArticleSecond{
		Abstract: a.Abstract,
		Summary:  a.Summary,
	}
}

func ConvertArticleDtoToEntity(article *dto.ArticleFirst) *Article {
	return &Article{
		Key:      article.Key,
		Abstract: article.Abstract,
		Summary:  article.Summary,
	}
}
