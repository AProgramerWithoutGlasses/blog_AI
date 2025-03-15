package service

import (
	"siwuai/internal/domain/model/dto"
	"siwuai/internal/domain/model/entity"
)

// UserDomainService 定义领域服务接口
type CodeDomainService interface {
	ValidateUser(user *entity.Code) error
	SaveCode(req *dto.CodeReq, key string) (code *dto.Code, err error)
	ExplainCode(req *dto.CodeReq) (code1 *dto.Code, err error)
}
