package app

import (
	"siwuai/internal/domain/model/dto"
)

// CodeUseCase 定义用户用例接口
type CodeUseCase interface {
	ExplainCode(req *dto.CodeReq) (*dto.Code, error)
}
