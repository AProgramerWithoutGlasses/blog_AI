package codecase

import (
	"fmt"
	"grpc-ddd-demo/internal/domain/model/entity"
	"grpc-ddd-demo/internal/domain/repository"
	"grpc-ddd-demo/internal/domain/service"
	"grpc-ddd-demo/internal/infrastructure/code_infrastructure"
	"grpc-ddd-demo/proto/code"
)

// CodeUseCase 定义用户用例接口
type CodeUseCase interface {
	ExplainCode(req *code.CodeRequest) (entity.Code, error)
}

type codeUseCase struct {
	repo              repository.CodeRepository
	codeDomainService service.CodeDomainService
}

// NewCodeUseCase 构造函数
func NewCodeUseCase(r repository.CodeRepository, ds service.CodeDomainService) CodeUseCase {
	return &codeUseCase{
		repo:              r,
		codeDomainService: ds,
	}
}

func (uc *codeUseCase) ExplainCode(req *code.CodeRequest) (code1 entity.Code, err error) {
	// 获取hash值
	key, err := code_infrastructure.Hash(req.CodeQuestion)
	if err != nil {
		fmt.Println("codecase.ExplainCode() unique.Hash() err:", err)
		return
	}

	code1, ok, err := uc.repo.GetCodeByHash(key)
	if err != nil {
		fmt.Println("codecase.ExplainCode() repo.GetCodeByHash() err:", err)
		return
	}

	if !ok {
		code1, err = uc.codeDomainService.SaveCode(req, key)
		if err != nil {
			fmt.Println("codecase.ExplainCode() codeDomainService.SaveCode() err:", err)
			return
		}
	}

	history := entity.History{
		UserID: uint(req.UserId),
		CodeID: code1.ID,
	}
	err = uc.repo.SaveHistory(history)
	if err != nil {
		fmt.Println("codecase.ExplainCode() repo.SaveHistory() err:", err)
		return
	}

	return
}
