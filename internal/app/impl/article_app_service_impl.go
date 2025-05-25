package impl

import (
	"fmt"
	"siwuai/internal/app"
	"siwuai/internal/domain/model/dto"
	"siwuai/internal/domain/model/entity"
	"siwuai/internal/domain/service"
	"siwuai/internal/infrastructure/persistence"
	"siwuai/internal/infrastructure/utils"
	"strconv"
)

type articleAppService struct {
	repo service.ArticleDomainServiceInterface
	code persistence.CodeRepository
}

func NewArticleAppService(repo service.ArticleDomainServiceInterface, code persistence.CodeRepository) app.ArticleAppServiceInterface {
	return &articleAppService{
		repo: repo,
		code: code,
	}
}

// GetArticleInfoFirst 第一次获取文章的摘要、总结、标签
func (a *articleAppService) GetArticleInfoFirst(content string, tags []string, articleID uint) (*dto.ArticleFirst, error) {
	// 根据文章的内容生成 hash值
	hashValue, err := utils.Hash(content)
	if err != nil {
		return nil, fmt.Errorf("(r *ArticleRepository) GetArticleInfoFirst -> %v", err)
	}

	articleInfo, err := a.repo.VerifyHash(hashValue)
	if err != nil {
		if err.Error() == "数据库中没有该 hash值" {
			// 封装数据
			ap := &dto.ArticlePrompt{
				Content:   content,
				Tags:      tags,
				ArticleID: articleID,
			}
			// 调用AI，提炼文章的摘要、总结、标签
			articleFirst, err := a.repo.AskAI(hashValue, ap)
			if err != nil {
				return nil, fmt.Errorf("(r *ArticleRepository) GetArticleInfoFirst -> %v", err)
			}
			return articleFirst, nil
		} else {
			return nil, fmt.Errorf("(r *ArticleRepository) GetArticleInfoFirst -> %v", err)
		}
	}

	// 如果hash存在，直接返回数据
	return articleInfo, nil
}

// SaveArticleID 保存文章的ID
func (a *articleAppService) SaveArticleID(key string, articleID uint) error {
	err := a.repo.SaveArticleID(key, articleID)
	if err != nil {
		return fmt.Errorf("(a *articleAppService) SaveArticleID -> %v", err)
	}
	return nil
}

// GetArticleInfo 非首次获取文章的信息
func (a *articleAppService) GetArticleInfo(articleID uint, userID uint) (*dto.ArticleSecond, []entity.Code, error) {
	articleSecond, err := a.repo.GetArticleInfo(articleID)
	if err != nil {
		return nil, nil, err
	}

	// 代码解释
	code, err := a.code.GetHistory(strconv.Itoa(int(userID)))
	if err != nil {
		return nil, nil, err
	}
	//for _, v := range code {
	//	value := &dto.CodeExplanation{
	//		Question: v.Question,
	//		Explanation: v.Explanation,
	//	}
	//	articleSecond.Codes = append(articleSecond.Codes, value)
	//}

	return articleSecond, code, nil
}

// DelArticleInfo 删除文章信息
func (a *articleAppService) DelArticleInfo(articleID uint) error {
	err := a.repo.DelArticleInfo(articleID)
	return err
}
