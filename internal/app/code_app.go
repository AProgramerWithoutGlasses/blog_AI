package app

import (
	"siwuai/internal/domain/model/dto"
)

// CodeApp 定义用户用例接口
type CodeApp interface {
	ExplainCode(req *dto.CodeReq) (*dto.Code, error)
}
