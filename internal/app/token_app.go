package app

import (
	"siwuai/internal/domain/model/dto"
)

// TokenApp 定义Token接口
type TokenApp interface {
	GenerateToken(req *dto.TokenReq) (resp *dto.TokenResp, err error)
}
