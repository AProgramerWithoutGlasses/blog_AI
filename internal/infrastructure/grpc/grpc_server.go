package grpc

import (
	"github.com/bits-and-blooms/bloom/v3"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"net"
	serviceimpl "siwuai/internal/domain/service/impl" // 导入服务实现
	"siwuai/internal/infrastructure/config"
	"siwuai/internal/infrastructure/redis_utils"
	server "siwuai/internal/server/grpc"
	pb "siwuai/proto/article"
	pbcode "siwuai/proto/code"
	pbtoken "siwuai/proto/token"
)

// RunGRPCServer 启动 gRPC 服务器，并启用 token 验证
func RunGRPCServer(port string, db *gorm.DB, rdb *redis_utils.RedisClient, bf *bloom.BloomFilter, cfg config.Config) error {
	lis, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		return err
	}

	// 初始化 token 服务
	tokenSvc := serviceimpl.NewTokenDomainService()

	// 创建 gRPC 服务器并注册拦截器
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(TokenValidationInterceptor(tokenSvc, cfg.Token.SecretKey)),
		grpc.StreamInterceptor(StreamTokenValidationInterceptor(tokenSvc, cfg.Token.SecretKey)),
	)

	// 注册 CodeService
	pbcode.RegisterCodeServiceServer(grpcServer, server.NewCodeGRPCHandler(db, rdb, bf, cfg))

	// 注册 ArticleService
	pb.RegisterArticleServiceServer(grpcServer, server.NewArticleGRPCHandler(db, cfg))

	// 注册 TokenService
	pbtoken.RegisterTokenServiceServer(grpcServer, server.NewTokenGRPCHandler(cfg))

	return grpcServer.Serve(lis)
}
