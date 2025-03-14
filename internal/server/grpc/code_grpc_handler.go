package grpc

import (
	"context"
	"gorm.io/gorm"
	"siwuai/internal/app"
	"siwuai/internal/domain/service"
	"siwuai/internal/infrastructure/persistence/impl"
	"siwuai/proto/code"
)

type codeGRPCHandler struct {
	code.UnimplementedCodeServiceServer
	uc app.CodeUseCase
}

// NewCodeGRPCHandler 初始化 gRPC 处理器及其依赖，传入 MySQL 连接
func NewCodeGRPCHandler(db *gorm.DB) code.CodeServiceServer {
	repo := impl.NewMySQLCodeRepository(db)
	ds := service.NewCodeDomainService(repo)
	uc := app.NewCodeUseCase(repo, ds)
	return &codeGRPCHandler{uc: uc}
}

func (h *codeGRPCHandler) ExplainCode(ctx context.Context, req *code.CodeRequest) (res *code.CodeResponse, err error) {
	code1, err := h.uc.ExplainCode(req)

	res = &code.CodeResponse{
		CodeExplain: code1.Explanation,
	}
	return
}
