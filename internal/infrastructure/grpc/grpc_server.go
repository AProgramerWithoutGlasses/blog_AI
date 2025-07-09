package grpc

import (
	"fmt"
	"github.com/bits-and-blooms/bloom/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"net"
	serviceimpl "siwuai/internal/domain/service/impl" // 导入服务实现
	"siwuai/internal/infrastructure/cache"
	"siwuai/internal/infrastructure/config"
	"siwuai/internal/infrastructure/constant"
	"siwuai/internal/infrastructure/redis_utils"
	server "siwuai/internal/server/grpc"
	pb "siwuai/proto/article"
	pbcode "siwuai/proto/code"
	pbquestion "siwuai/proto/question"
	pbtoken "siwuai/proto/token"
	pbvector "siwuai/proto/vector"
)

// RunGRPCServer 启动 gRPC 服务器，并启用 token 验证
func RunGRPCServer(port string, db *gorm.DB, rdb *redis_utils.RedisClient, bf *bloom.BloomFilter, cfg config.Config, cacheManager *cache.CacheManager, jc constant.JudgingCacheType) error {
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
	pb.RegisterArticleServiceServer(grpcServer, server.NewArticleGRPCHandler(db, cfg, cacheManager, jc))

	// 注册 TokenService
	pbtoken.RegisterTokenServiceServer(grpcServer, server.NewTokenGRPCHandler(cfg))

	// 注册 QuestionService
	pbquestion.RegisterQuestionServiceServer(grpcServer, server.NewQuestionGRPCHandler(db, cfg, cacheManager, jc))

	// 注册 VectorService
	pbvector.RegisterVectorServiceServer(grpcServer, server.NewVectorGrpcHandler(cfg))

	msg := fmt.Sprintf("gRPC 服务器成功启动在端口 %s...", port)
	fmt.Println(msg)
	zap.L().Info(msg)

	return grpcServer.Serve(lis)
}
