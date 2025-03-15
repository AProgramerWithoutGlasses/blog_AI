package service

import (
	"siwuai/internal/domain/model/dto"
)

type ArticleDomainServiceInterface interface {
	VerifyHash(key string) (*dto.ArticleFirst, error)
	AskAI(key string, content string) (*dto.ArticleFirst, error)
	SaveArticleID(key string, articleID uint) error
	GetArticleInfo(articleID uint) (*dto.ArticleSecond, error)
	DelArticleInfo(articleID uint) error
}
