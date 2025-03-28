package grpc

import (
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"siwuai/internal/app"
	appimpl "siwuai/internal/app/impl"
	"siwuai/internal/domain/model/dto"
	serviceimpl "siwuai/internal/domain/service/impl"
	"siwuai/internal/infrastructure/constant"
	persistenceimpl "siwuai/internal/infrastructure/persistence/impl"
	"siwuai/internal/infrastructure/redis_utils"
	pb "siwuai/proto/code"
)

type codeGRPCHandler struct {
	pb.UnimplementedCodeServiceServer
	uc app.CodeApp
}

func NewCodeGRPCHandler(db *gorm.DB, redisClient *redis_utils.RedisClient, bf *bloom.BloomFilter) pb.CodeServiceServer {
	repo := persistenceimpl.NewMySQLCodeRepository(db)
	sign := constant.NewJudgingSign()
	ds := serviceimpl.NewCodeDomainService(repo, redisClient, bf, sign)
	uc := appimpl.NewCodeApp(repo, ds)
	return &codeGRPCHandler{uc: uc}
}

func (h *codeGRPCHandler) ExplainCode(req *pb.CodeRequest, stream pb.CodeService_ExplainCodeServer) error {
	// 接收
	req1 := dto.CodeReq{UserId: uint(req.UserId), Question: req.CodeQuestion, CodeType: req.CodeType}

	// 业务
	code1, err := h.uc.ExplainCode(&req1)
	if err != nil {
		zap.L().Error("ExplainCode() ", zap.Error(err))
		return err
	}
	fmt.Printf("最后收到的code1：%#v\n", code1)

	// 如果 code1.Stream 为 nil，说明缓存命中，那么则将缓存的结果手动转换为流式输出
	if code1.Stream == nil {
		code1.Stream = make(chan string) // 初始化通道
		go func() {
			defer close(code1.Stream) // 确保通道关闭
			code1.Stream <- code1.Explanation
		}()
	}

	// SSE 响应
	for chunk := range code1.Stream {
		if err = stream.Send(&pb.CodeResponse{CodeExplain: chunk}); err != nil {
			zap.L().Error("stream.Send(&pb.CodeResponse{CodeExplain: chunk}) err: ", zap.Error(err))
			return err
		}
		fmt.Println(chunk)

	}

	return nil
}
