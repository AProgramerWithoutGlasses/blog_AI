package grpc

import (
	"gorm.io/gorm"
	pbcode "grpc-ddd-demo/proto/code"
	pbllm "grpc-ddd-demo/proto/llm"
	pbuser "grpc-ddd-demo/proto/user"
	"net"

	"google.golang.org/grpc"
	server "grpc-ddd-demo/internal/server/grpc"
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

	// 注册 LLMService
	llmSvc, err := server.NewLLMGRPCHandler()
	if err != nil {
		return err
	}
	pbllm.RegisterLLMServiceServer(grpcServer, llmSvc)

	// 注册 CodeService
	pbcode.RegisterCodeServiceServer(grpcServer, server.NewCodeGRPCHandler(db))

	return grpcServer.Serve(lis)
}
