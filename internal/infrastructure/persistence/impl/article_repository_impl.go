package impl

import (
	"fmt"
	"gorm.io/gorm"
	"siwuai/internal/domain/model/entity"
)

type ArticleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{
		db: db,
	}
}

func (a *ArticleRepository) VerifyHash(key string) (*entity.Article, error) {
	var articleInfo entity.Article
	result := a.db.Model(&entity.Article{}).Where("key = ?", key).Scan(&articleInfo)
	if result.Error != nil {
		return nil, fmt.Errorf("(a *ArticleRepository) VerifyHash -> %v", result.Error)
	} else if result.RowsAffected == 0 {
		return nil, fmt.Errorf("数据库中没有该 hash值")
	}
	return &articleInfo, nil
}
