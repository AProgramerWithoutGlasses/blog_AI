package grpc

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"grpc-ddd-demo/internal/app"
	"grpc-ddd-demo/internal/domain/model/dto"
	"grpc-ddd-demo/internal/domain/service"
	"grpc-ddd-demo/internal/infrastructure/persistence"
	"grpc-ddd-demo/proto/code"
)

type codeGRPCHandler struct {
	code.UnimplementedCodeServiceServer
	uc app.CodeUseCase
}

// NewCodeGRPCHandler 初始化 gRPC 处理器及其依赖，传入 MySQL 连接
func NewCodeGRPCHandler(db *gorm.DB) code.CodeServiceServer {
	repo := persistence.NewMySQLCodeRepository(db)
	ds := service.NewCodeDomainService(repo)
	uc := app.NewCodeUseCase(repo, ds)
	return &codeGRPCHandler{uc: uc}
}

func (h *codeGRPCHandler) ExplainCode(ctx context.Context, req *code.CodeRequest) (res *code.CodeResponse, err error) {
	req1 := dto.CodeReq{UserId: uint(req.UserId), Question: req.CodeQuestion, CodeType: req.CodeType}

	code1, err := h.uc.ExplainCode(&req1)
	if err != nil {
		fmt.Println("ExplainCode()", err)
		return
	}

	res = &code.CodeResponse{
		CodeExplain: code1.Explanation,
	}
	return
}
