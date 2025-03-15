package app

import (
	"fmt"
	"grpc-ddd-demo/internal/app"
	"grpc-ddd-demo/internal/domain/model/dto"
	"grpc-ddd-demo/internal/domain/repository"
	"grpc-ddd-demo/internal/domain/service"
)

type codeUseCase struct {
	repo              repository.CodeRepository
	codeDomainService service.CodeDomainService
}

// NewCodeUseCase 构造函数
func NewCodeUseCase(r repository.CodeRepository, ds service.CodeDomainService) app.CodeUseCase {
	return &codeUseCase{
		repo:              r,
		codeDomainService: ds,
	}
}

func (uc *codeUseCase) ExplainCode(req *dto.CodeReq) (code1 *dto.Code, err error) {
	code1, err = uc.codeDomainService.ExplainCode(req)
	if err != nil {
		fmt.Println("app.ExplainCode() uc.codeDomainService.ExplainCode() err: ", err)
		return
	}
	return
}
