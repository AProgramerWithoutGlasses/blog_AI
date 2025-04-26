package app

import (
	"siwuai/internal/domain/model/dto"
	"siwuai/internal/domain/model/entity"
)

type ArticleAppServiceInterface interface {
	GetArticleInfoFirst(content string, tags []string) (*dto.ArticleFirst, error)
	SaveArticleID(key string, articleID uint) error
	GetArticleInfo(articleID uint, userID uint) (*dto.ArticleSecond, []entity.Code, error)
	DelArticleInfo(articleID uint) error
}
