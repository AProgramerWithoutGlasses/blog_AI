package service

import (
	"fmt"
	"grpc-ddd-demo/internal/domain/model/entity"
	"grpc-ddd-demo/internal/domain/repository"
	"grpc-ddd-demo/internal/infrastructure/code_infrastructure"
	"grpc-ddd-demo/proto/code"
)

// UserDomainService 定义领域服务接口
type CodeDomainService interface {
	ValidateUser(user *entity.Code) error
	SaveCode(req *code.CodeRequest, key string) (code entity.Code, err error)
}

type codeDomainService struct {
	repo repository.CodeRepository
}

// NewUserDomainService 创建领域服务实例
func NewCodeDomainService() CodeDomainService {
	return &codeDomainService{}
}

// ValidateUser 校验用户信息
func (s *codeDomainService) ValidateUser(code *entity.Code) error {
	return nil
}

// SaveCode 调用llm生成回复并且将回复存入表中
func (s *codeDomainService) SaveCode(req *code.CodeRequest, key string) (code1 entity.Code, err error) {
	explain, err := code_infrastructure.Generate(req.CodeQuestion)
	if err != nil {
		fmt.Println("codecase.ExplainCode() llm.Generate() err:", err)
		return
	}

	code1 = entity.Code{
		Key:         key,
		Explanation: explain,
		Question:    req.CodeQuestion,
	}
	_, err = s.repo.SaveCode(code1)
	if err != nil {
		fmt.Println("codecase ExplainCode() repo.SaveCode() err:", err)
		return
	}

	return
}
