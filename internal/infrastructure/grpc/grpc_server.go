package grpc

import (
	"gorm.io/gorm"
	"net"
	pb "siwuai/proto/article"
	pbcode "siwuai/proto/code"

	"google.golang.org/grpc"
	server "siwuai/internal/server/grpc"
)

// RunGRPCServer 启动 gRPC 服务器，同时注册 UserService 和 LLMService
func RunGRPCServer(port string, db *gorm.DB) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	grpcServer := grpc.NewServer()

	// 注册 CodeService
	pbcode.RegisterCodeServiceServer(grpcServer, server.NewCodeGRPCHandler(db))

	// 注册 ArticleService
	pb.RegisterArticleServiceServer(grpcServer, server.NewArticleGRPCHandler(db))

	return grpcServer.Serve(lis)
}
