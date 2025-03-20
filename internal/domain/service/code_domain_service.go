package service

import (
	"siwuai/internal/domain/model/dto"
)

type CodeDomainService interface {
	ExplainCode(req *dto.CodeReq) (*dto.Code, error)
	FetchAndSave(req *dto.CodeReq, key string) (*dto.Code, error)
	SaveToRedis(key string, code *dto.Code) (err error)
}
