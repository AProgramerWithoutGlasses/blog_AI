package grpc

import (
	"github.com/bits-and-blooms/bloom/v3"
	"gorm.io/gorm"
	"net"
	"siwuai/internal/infrastructure/config"
	"siwuai/internal/infrastructure/redis_utils"
	pb "siwuai/proto/article"
	pbcode "siwuai/proto/code"
	pbtoken "siwuai/proto/token"

	"google.golang.org/grpc"
	server "siwuai/internal/server/grpc"
)

// RunGRPCServer 启动 gRPC 服务器，同时注册 UserService 和 LLMService
func RunGRPCServer(port string, db *gorm.DB, rdb *redis_utils.RedisClient, bf *bloom.BloomFilter, cfg config.Config) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()

	// 注册 CodeService
	pbcode.RegisterCodeServiceServer(grpcServer, server.NewCodeGRPCHandler(db, rdb, bf))

	// 注册 ArticleService
	pb.RegisterArticleServiceServer(grpcServer, server.NewArticleGRPCHandler(db))

	pbtoken.RegisterTokenServiceServer(grpcServer, server.NewTokenGRPCHandler(cfg))

	return grpcServer.Serve(lis)
}
