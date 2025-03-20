package grpc

import (
	"context"
	"fmt"
	"siwuai/internal/app"
	appimpl "siwuai/internal/app/impl"
	"siwuai/internal/domain/model/dto"
	service "siwuai/internal/domain/service/impl"
	"siwuai/internal/infrastructure/config"
	pbToken "siwuai/proto/token"
)

type tokenGRPCHandler struct {
	pbToken.UnimplementedTokenServiceServer
	app              app.TokenApp
	secretKey        string
	generateTokenKey string
}

// NewTokenGRPCHandler 构造方法
func NewTokenGRPCHandler(cfg config.Config) pbToken.TokenServiceServer {
	service1 := service.NewTokenDomainService()
	app1 := appimpl.NewTokenApp(service1, cfg)
	return &tokenGRPCHandler{
		app:              app1,
		secretKey:        cfg.Token.SecretKey,
		generateTokenKey: cfg.Token.GenerateTokenKey,
	}
}

func (h *tokenGRPCHandler) GenerateToken(ctx context.Context, req *pbToken.TokenRequest) (resp *pbToken.TokenResponse, err error) {
	req1 := dto.TokenReq{
		GenerateTokenKey: req.GenerateTokenKey,
	}

	resp1, err := h.app.GenerateToken(&req1)
	if err != nil {
		fmt.Println("ExplainCode()", err)
		return
	}

	resp = &pbToken.TokenResponse{
		Token: resp1.Token,
	}

	return
}
