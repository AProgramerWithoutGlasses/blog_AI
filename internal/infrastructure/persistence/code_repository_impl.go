package persistence

import (
	"errors"
	"gorm.io/gorm"
	"grpc-ddd-demo/internal/domain/model/entity"
	"grpc-ddd-demo/internal/domain/repository"
)

type mysqlCodeRepository struct {
	db *gorm.DB
}

// NewMySQLCodeRepository 返回基于 MySQL 的仓储实现
func NewMySQLCodeRepository(db *gorm.DB) repository.CodeRepository {
	return &mysqlCodeRepository{db: db}
}

func (r *mysqlCodeRepository) GetCodeByHash(key string) (code entity.Code, ok bool, err error) {
	err = r.db.First(&code, "key = ?", key).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = nil
		}
		return
	}

	return code, true, err
}

func (r *mysqlCodeRepository) SaveCode(code entity.Code) (uint, error) {
	err := r.db.Create(&code).Error
	return code.ID, err
}

// GetHistory 根据history表中某个userid的后10条记录中的codeId去查询Code表中信息
func (r *mysqlCodeRepository) GetHistory(userId string) (history []entity.Code, err error) {
	err = r.db.Where("id IN (?)",
		r.db.Model(&entity.History{}).
			Select("code_id").
			Where("user_id = ?", userId).
			Order("created_at DESC").
			Limit(10),
	).Find(&history).Error

	return
}

func (r *mysqlCodeRepository) SaveHistory(entity.History) (err error) {
	return r.db.Create(&entity.History{}).Error
}
