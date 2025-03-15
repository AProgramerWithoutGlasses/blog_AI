package app

import (
	"siwuai/internal/domain/model/dto"
)

type ArticleAppServiceInterface interface {
	GetArticleInfoFirst(content string, tags []string) (*dto.ArticleFirst, error)
	SaveArticleID(key string, articleID uint) error
	GetArticleInfo(articleID uint) (*dto.ArticleSecond, error)
	DelArticleInfo(articleID uint) error
}
