package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"siwuai/internal/domain/service"
)

// TokenValidationInterceptor 创建一个 gRPC 一元拦截器，用于验证 token
func TokenValidationInterceptor(tokenSvc service.TokenDomainService, secretKey string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		// 跳过 TokenService.GenerateToken 方法的 token 验证
		if info.FullMethod == "/token.TokenService/GenerateToken" {
			return handler(ctx, req)
		}

		// 从上下文中获取元数据
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "缺少元数据")
		}

		// 从元数据中提取 token
		tokens, ok := md["authorization"]
		if !ok || len(tokens) == 0 {
			return nil, status.Error(codes.Unauthenticated, "缺少 token")
		}
		tokenString := tokens[0]

		// 验证 token
		err = tokenSvc.ValidateToken(tokenString, secretKey)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "token 验证失败: %v", err)
		}

		fmt.Println("token 验证成功，放行")

		// token 有效，继续处理
		return handler(ctx, req)
	}
}
