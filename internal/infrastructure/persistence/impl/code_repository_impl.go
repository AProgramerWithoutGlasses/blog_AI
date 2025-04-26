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
			fmt.Println("该错误已手动忽略:  r.db.Where().First(&code) err: ", err)
			err = nil
		}
		err = fmt.Errorf("r.db.Where(`key` = ?, key).First(&code) err: %v", err)
		return
	}
	return code, true, err
}

func (r *mysqlCodeRepository) SaveCode(code *entity.Code) (codeId uint, err error) {
	// 记录调用前的 ID
	originalID := code.ID

	// 使用 FirstOrCreate，基于 Key 字段判断是否已存在
	result := r.db.Where(entity.Code{Key: code.Key}).FirstOrCreate(code)
	if result.Error != nil {
		err = fmt.Errorf("r.db.FirstOrCreate() err: %v", result.Error)
		return 0, result.Error
	}

	codeId = code.ID
	// 判断记录是否新创建
	if result.RowsAffected == 1 {
		fmt.Printf("校验数据一致性: 成功将记录缓存到mysql: %s ——— %#v\n", code.Key, code)
	} else if originalID == 0 && codeId != 0 {
		// 如果 RowsAffected 不准确，可以用 ID 判断
		fmt.Printf("校验数据一致性: 记录已存在于mysql: %s ——— %#v\n", code.Key, code)
	} else {
		fmt.Printf("校验数据一致性: 记录已存在于mysql: %s ——— %#v\n", code.Key, code)
	}

	return codeId, nil
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
	err = r.db.Create(&history).Error
	if err != nil {
		err = fmt.Errorf("r.db.Create() err: %v", err)
	}
	return
}
