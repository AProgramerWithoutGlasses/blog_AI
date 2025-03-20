package grpc

import (
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"gorm.io/gorm"
	"siwuai/internal/app"
	appimpl "siwuai/internal/app/impl"
	"siwuai/internal/domain/model/dto"
	serviceimpl "siwuai/internal/domain/service/impl"
	persistenceimpl "siwuai/internal/infrastructure/persistence/impl"
	"siwuai/internal/infrastructure/redis_utils"
	pb "siwuai/proto/code"
)

type codeGRPCHandler struct {
	pb.UnimplementedCodeServiceServer
	uc app.CodeApp
}

// NewCodeGRPCHandler 初始化 gRPC 处理器及其依赖，传入 MySQL 连接
func NewCodeGRPCHandler(db *gorm.DB, redisClient *redis_utils.RedisClient, bf *bloom.BloomFilter) pb.CodeServiceServer {
	repo := persistenceimpl.NewMySQLCodeRepository(db)
	ds := serviceimpl.NewCodeDomainService(repo, redisClient, bf)
	uc := appimpl.NewCodeApp(repo, ds)
	return &codeGRPCHandler{uc: uc}
}

func (h *codeGRPCHandler) ExplainCode(req *pb.CodeRequest, stream pb.CodeService_ExplainCodeServer) error {
	// 接收
	req1 := dto.CodeReq{UserId: uint(req.UserId), Question: req.CodeQuestion, CodeType: req.CodeType}

	// 业务
	code1, err := h.uc.ExplainCode(&req1) // 此时code1中的chan还正在被写
	if err != nil {
		fmt.Println("ExplainCode()", err)
		return err
	}

	fmt.Printf("code1：%#v\n", code1)

	// 如果 code1.Stream 为 nil，则将 code1.Explanation 拆分为较小的片段并发送
	if code1.Stream == nil {
		fmt.Println("检测到缓存命中，开始将结果手动转换为流式")
		// 初始化 channel
		code1.Stream = make(chan string)
		go func() {
			splitIntoChunks(code1.Explanation, 3, code1.Stream)
		}()
	}

	// sse响应
	for chunk := range code1.Stream {
		if chunk != "" {
			if err = stream.Send(&pb.CodeResponse{CodeExplain: chunk}); err != nil {
				fmt.Println("stream.Send() err ", err)
				return err
			}
			fmt.Println(chunk)
		}
	}

	return nil
}

func splitIntoChunks(input string, chunkSize int, ch chan string) {
	defer close(ch)
	for start := 0; start < len(input); start += chunkSize {
		end := start + chunkSize
		if end > len(input) {
			end = len(input)
		}
		chunk := input[start:end]
		ch <- chunk
	}
}
