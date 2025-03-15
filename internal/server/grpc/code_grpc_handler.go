package grpc

import (
	"context"
	"fmt"
	"gorm.io/gorm"
	"siwuai/internal/app"
	appimpl "siwuai/internal/app/impl"
	"siwuai/internal/domain/model/dto"
	serviceimpl "siwuai/internal/domain/service/impl"
	persistenceimpl "siwuai/internal/infrastructure/persistence/impl"
	"siwuai/proto/code"
)

type codeGRPCHandler struct {
	code.UnimplementedCodeServiceServer
	uc app.CodeApp
}

// NewCodeGRPCHandler 初始化 gRPC 处理器及其依赖，传入 MySQL 连接
func NewCodeGRPCHandler(db *gorm.DB) code.CodeServiceServer {
	repo := persistenceimpl.NewMySQLCodeRepository(db)
	ds := serviceimpl.NewCodeDomainService(repo)
	uc := appimpl.NewCodeApp(repo, ds)
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
