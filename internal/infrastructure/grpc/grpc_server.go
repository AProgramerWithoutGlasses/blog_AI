package grpc

import (
	"gorm.io/gorm"
	"net"
	pbcode "siwuai/proto/code"
	pbuser "siwuai/proto/user"

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

	// 注册 UserService
	pbuser.RegisterUserServiceServer(grpcServer, server.NewUserGRPCHandler(db))

	// 注册 CodeService
	pbcode.RegisterCodeServiceServer(grpcServer, server.NewCodeGRPCHandler(db))

	return grpcServer.Serve(lis)
}
