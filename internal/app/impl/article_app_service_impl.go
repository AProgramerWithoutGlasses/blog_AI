package impl

import (
	"fmt"
	"siwuai/internal/app"
	"siwuai/internal/domain/model/entity"
	"siwuai/internal/domain/service"
	"siwuai/internal/infrastructure/utils"
)

type articleAppService struct {
	repo service.ArticleDomainServiceInterface
}

func NewArticleAppService(repo service.ArticleDomainServiceInterface) app.ArticleAppServiceInterface {
	return &articleAppService{
		repo: repo,
	}
}

func (a *articleAppService) GetArticleInfoFirst(content string, tags []string) (*entity.Article, error) {
	// 根据文章的内容生成 hash值
	hashValue, err := utils.Hash(content)
	if err != nil {
		return nil, fmt.Errorf("(r *ArticleRepository) GetArticleInfoFirst -> %v", err)
	}

	articleInfo, err := a.repo.VerifyHash(hashValue)
	if err.Error() == "数据库中没有该 hash值" {
		fmt.Println(articleInfo)
	} else if err != nil {
		return nil, fmt.Errorf("(r *ArticleRepository) GetArticleInfoFirst -> %v", err)
	}

	// 封装数据
	return nil, nil
}
