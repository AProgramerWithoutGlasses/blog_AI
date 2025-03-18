package impl

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"siwuai/internal/domain/model/entity"
	"siwuai/internal/infrastructure/persistence"
)

type mysqlCodeRepository struct {
	db *gorm.DB
}

// NewMySQLCodeRepository 返回基于 MySQL 的仓储实现
func NewMySQLCodeRepository(db *gorm.DB) persistence.CodeRepository {
	return &mysqlCodeRepository{db: db}
}

func (r *mysqlCodeRepository) GetCodeByHash(key string) (code entity.Code, ok bool, err error) {
	err = r.db.Where("`key` = ?", key).First(&code).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		}
		return
	}
	return code, true, err
}

func (r *mysqlCodeRepository) SaveCode(code *entity.Code) (uint, error) {
	err := r.db.Create(&code).Error
	if err != nil {
		fmt.Println("r.db.Create(&code) err: ", err)
	}
	return code.ID, err
}

// GetHistory 根据history表中某个userid的后10条记录中的codeId去查询Code表中信息
func (r *mysqlCodeRepository) GetHistory(userId string) (history []entity.Code, err error) {
	// 构造子查询，查询指定 userId 的最新 10 条历史记录中的 code_id
	subQuery := r.db.Model(&entity.History{}).
		Select("code_id").
		Where("user_id = ?", userId).
		Order("created_at DESC").
		Limit(10)

	// 使用 JOIN 将 Code 表和子查询关联
	err = r.db.
		Joins("JOIN (?) AS h ON h.code_id = sw_ai_codes.id", subQuery).
		Find(&history).Error

	return
}

func (r *mysqlCodeRepository) SaveHistory(history entity.History) (err error) {
	return r.db.Create(&history).Error

}
