package impl

import (
	"fmt"
	"gorm.io/gorm"
	"siwuai/internal/domain/model/entity"
)

type articleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *articleRepository {
	return &articleRepository{
		db: db,
	}
}

// VerifyHash 验证hash值
func (a *articleRepository) VerifyHash(key string) (*entity.Article, error) {

	var articleInfo entity.Article
	result := a.db.Model(&entity.Article{}).Where("`key` = ?", key).Scan(&articleInfo)
	if result.Error != nil {
		return nil, fmt.Errorf("(a *ArticleRepository) VerifyHash -> %v", result.Error)
	} else if result.RowsAffected == 0 {
		return nil, fmt.Errorf("数据库中没有该 hash值")
	}
	return &articleInfo, nil
}

// SaveArticleInfo 保存文章的信息
func (a *articleRepository) SaveArticleInfo(article *entity.Article) error {

	tx := a.db.Begin()
	if tx.Error != nil {
		tx.Rollback()
		return fmt.Errorf("(a *articleRepository) SaveArticleInfo -> %v", tx.Error)
	}

	err := tx.Model(&entity.Article{}).Create(article).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("(a *articleRepository) SaveArticleInfo -> %v", err)
	}

	err = tx.Commit().Error
	if err != nil {
		return fmt.Errorf("(a *articleRepository) SaveArticleInfo -> %v", err)
	}

	return nil
}

// SaveArticleID 保存文章的ID
func (a *articleRepository) SaveArticleID(key string, articleID uint) error {

	tx := a.db.Begin()
	if tx.Error != nil {
		tx.Rollback()
		return fmt.Errorf("(a *articleRepository) SaveArticleID -> %v", tx.Error)
	}

	result := tx.Model(&entity.Article{}).Where("`key` = ?", key).Update("article_id", articleID)
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("(a *articleRepository) SaveArticleID -> %v", result.Error)
	} else if result.RowsAffected <= 0 {
		tx.Rollback()
		return fmt.Errorf("保存文章的ID失败")
	}

	err := tx.Commit().Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("(a *articleRepository) SaveArticleInfo -> %v", err)
	}

	return nil
}

// GetArticleInfo 查询文章信息
func (a *articleRepository) GetArticleInfo(articleID uint) (*entity.Article, error) {

	var articleInfo entity.Article
	result := a.db.Model(&entity.Article{}).Where("article_id = ?", articleID).Scan(&articleInfo)
	if result.Error != nil {
		return nil, fmt.Errorf("(a *articleRepository) GetArticleInfo -> %v", result.Error)
	} else if result.RowsAffected == 0 {
		return nil, fmt.Errorf("查询文章信息失败")
	}
	return &articleInfo, nil
}

// DelArticleInfo 删除文章信息
func (a *articleRepository) DelArticleInfo(articleID uint) error {

	tx := a.db.Begin()
	if tx.Error != nil {
		tx.Rollback()
		return fmt.Errorf("(a *articleRepository) DelArticleInfo -> %v", tx.Error)
	}

	result := tx.Where("article_id = ?", articleID).Delete(&entity.Article{})
	if result.Error != nil {
		tx.Rollback()
		return fmt.Errorf("(a *articleRepository) DelArticleInfo -> %v", result.Error)
	} else if result.RowsAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("数据库中没有该文章的信息")
	}

	err := tx.Commit().Error
	if err != nil {
		return fmt.Errorf("(a *articleRepository) DelArticleInfo -> %v", err)
	}

	return nil
}
