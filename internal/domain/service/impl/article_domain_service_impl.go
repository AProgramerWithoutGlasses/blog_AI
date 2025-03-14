package service

import (
	"fmt"
	"siwuai/internal/domain/model/dto"
	"siwuai/internal/domain/service"
	"siwuai/internal/infrastructure/persistence"
)

type articleDomainService struct {
	repo persistence.ArticleRepository
}

func NewArticleService(repo persistence.ArticleRepository) service.ArticleDomainServiceInterface {
	return &articleDomainService{
		repo: repo,
	}
}

func (a *articleDomainService) VerifyHash(key string) (*dto.Article, error) {
	articleInfo, err := a.repo.VerifyHash(key)
	if err != nil {
		if err.Error() == "数据库中没有该 hash值" {

		} else {
			return nil, fmt.Errorf("(a *ArticleService) VerifyHash -> %v", err)
		}
	}
	return articleInfo.ConvertArticleEntityToDto(), nil
}
