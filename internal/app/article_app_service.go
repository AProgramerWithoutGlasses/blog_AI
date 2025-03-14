package app

import "siwuai/internal/domain/model/entity"

type ArticleAppServiceInterface interface {
	GetArticleInfoFirst(content string, tags []string) (*entity.Article, error)
}
