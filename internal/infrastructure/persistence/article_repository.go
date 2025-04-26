package persistence

import "siwuai/internal/domain/model/entity"

type ArticleRepositoryInterface interface {
	VerifyHash(key string) (*entity.Article, error)
	SaveArticleInfo(article *entity.Article) error
	SaveArticleID(key string, articleID uint) error
	GetArticleInfo(articleID uint) (*entity.Article, error)
	DelArticleInfo(articleID uint) error
}
