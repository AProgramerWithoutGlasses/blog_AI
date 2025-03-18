package impl

import (
	"fmt"
	"siwuai/internal/app"
	"siwuai/internal/domain/model/dto"
	"siwuai/internal/domain/service"
	"siwuai/internal/infrastructure/config"
)

type tokenApp struct {
	tokenDomainService service.TokenDomainService
	secretKey          string
	generateTokenKey   string
}

// NewTokenApp 构造函数
func NewTokenApp(service1 service.TokenDomainService, cfg config.Config) app.TokenApp {
	return &tokenApp{
		tokenDomainService: service1,
		secretKey:          cfg.Token.SecretKey,
		generateTokenKey:   cfg.Token.GenerateTokenKey,
	}
}

func (app *tokenApp) GenerateToken(req *dto.TokenReq) (resp *dto.TokenResp, err error) {
	token, err := app.tokenDomainService.GenerateToken(req, app.generateTokenKey, app.secretKey)
	if err != nil {
		fmt.Println("app.tokenDomainService.GenerateToken() err: ", err)
		return
	}

	resp = &dto.TokenResp{}
	resp.Token = token
	return
}
