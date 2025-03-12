package grpc

import (
	"context"
	"gorm.io/gorm"
	"grpc-ddd-demo/internal/app/codecase"
	"grpc-ddd-demo/internal/domain/service"
	"grpc-ddd-demo/internal/infrastructure/persistence"
	"grpc-ddd-demo/proto/code"
)

type codeGRPCHandler struct {
	code.UnimplementedCodeServiceServer
	uc codecase.CodeUseCase
}

// NewCodeGRPCHandler 初始化 gRPC 处理器及其依赖，传入 MySQL 连接
func NewCodeGRPCHandler(db *gorm.DB) code.CodeServiceServer {
	repo := persistence.NewMySQLCodeRepository(db)
	ds := service.NewCodeDomainService()
	uc := codecase.NewCodeUseCase(repo, ds)
	return &codeGRPCHandler{uc: uc}
}

func (h *codeGRPCHandler) ExplainCode(ctx context.Context, req *code.CodeRequest) (res *code.CodeResponse, err error) {
	code1, err := h.uc.ExplainCode(req)

	res = &code.CodeResponse{
		CodeExplain: code1.Explanation,
	}
	return
}
