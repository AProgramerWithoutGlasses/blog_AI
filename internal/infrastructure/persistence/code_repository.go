package persistence

import (
	"siwuai/internal/domain/model/entity"
)

// CodeRepository 定义了代码块的访问接口
type CodeRepository interface {
	GetCodeByHash(key string) (entity.Code, bool, error)
	SaveCode(code *entity.Code) (uint, error)
	SaveHistory(entity.History) error
	GetHistory(userId string) ([]entity.Code, error)
}
