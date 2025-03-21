package impl

import (
	"fmt"
	"siwuai/internal/app"
	"siwuai/internal/domain/model/dto"
	"siwuai/internal/domain/service"
	"siwuai/internal/infrastructure/persistence"
	"strconv"
)

type codeApp struct {
	repo              persistence.CodeRepository
	codeDomainService service.CodeDomainService
}

// NewCodeApp 构造函数
func NewCodeApp(r persistence.CodeRepository, ds service.CodeDomainService) app.CodeApp {
	return &codeApp{
		repo:              r,
		codeDomainService: ds,
	}
}

func (uc *codeApp) ExplainCode(req *dto.CodeReq) (code1 *dto.Code, err error) {
	code1, err = uc.codeDomainService.ExplainCode(req)
	if err != nil {
		fmt.Println("app.ExplainCode() uc.codeDomainService.ExplainCode() err: ", err)
		return
	}

	history, err := uc.repo.GetHistory(strconv.Itoa(int(req.UserId)))
	if err != nil {
		fmt.Println("app.ExplainCode() uc.repo.GetHistory() err: ", err)
		return
	}

	fmt.Printf("userId: %d 's histories is: %v\n", req.UserId, history)

	return
}
