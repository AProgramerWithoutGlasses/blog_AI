package service

import (
	"fmt"
	"siwuai/internal/domain/model/entity"
	"siwuai/internal/infrastructure/code_infrastructure"
	"siwuai/internal/infrastructure/persistence"
	"siwuai/proto/code"
)

// UserDomainService 定义领域服务接口
type CodeDomainService interface {
	ValidateUser(user *entity.Code) error
	SaveCode(req *code.CodeRequest, key string) (code entity.Code, err error)
}

type codeDomainService struct {
	repo persistence.CodeRepository
}

// NewUserDomainService 创建领域服务实例
func NewCodeDomainService(repo persistence.CodeRepository) CodeDomainService {
	return &codeDomainService{repo: repo}
}

// ValidateUser 校验用户信息
func (s *codeDomainService) ValidateUser(code *entity.Code) error {
	return nil
}

// SaveCode 调用llm生成回复并且将回复存入表中
func (s *codeDomainService) SaveCode(req *code.CodeRequest, key string) (entity.Code, error) {
	explain, err := code_infrastructure.Generate(req.CodeQuestion)
	if err != nil {
		fmt.Println("codecase.ExplainCode() llm.Generate() err:", err)
		return entity.Code{}, err
	}

	code1 := entity.Code{
		Key:         key,
		Explanation: explain,
		Question:    req.CodeQuestion,
	}
	code1.ID, err = s.repo.SaveCode(code1)
	if err != nil {
		fmt.Println("codecase ExplainCode() repo.SaveCode() err:", err)
		return entity.Code{}, err
	}

	return code1, nil

}
