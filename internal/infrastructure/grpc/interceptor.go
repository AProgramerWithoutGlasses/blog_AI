package grpc

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"siwuai/internal/domain/service"
	"strings"
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

		// 去掉 Bearer 前缀（支持大小写无关）
		tokenString = strings.TrimSpace(tokenString) // 去掉首尾空格
		originalToken := tokenString
		if strings.HasPrefix(strings.ToLower(tokenString), "bearer ") {
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
			tokenString = strings.TrimPrefix(tokenString, "bearer ")
		}
		if tokenString == originalToken {
			return nil, status.Error(codes.Unauthenticated, "无效的 Authorization 头格式，缺少 Bearer 前缀")
		}

		// 验证 token
		err = tokenSvc.ValidateToken(tokenString, secretKey)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "token 验证失败: %v", err)
		}

		// token 有效，继续处理
		return handler(ctx, req)
	}
}

// StreamTokenValidationInterceptor 创建一个 gRPC 流式拦截器，用于验证 token
func StreamTokenValidationInterceptor(tokenSvc service.TokenDomainService, secretKey string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 跳过 TokenService.GenerateToken 方法的 token 验证
		if info.FullMethod == "/token.TokenService/GenerateToken" {
			fmt.Println("Skipping token validation for /token.TokenService/GenerateToken")
			return handler(srv, ss)
		}

		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "缺少元数据")
		}

		tokens, ok := md["authorization"]
		if !ok || len(tokens) == 0 {
			return status.Error(codes.Unauthenticated, "缺少 token")
		}
		tokenString := tokens[0]

		tokenString = strings.TrimSpace(tokenString)
		originalToken := tokenString
		if strings.HasPrefix(strings.ToLower(tokenString), "bearer ") {
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
			tokenString = strings.TrimPrefix(tokenString, "bearer ")
		}
		if tokenString == originalToken {
			return status.Error(codes.Unauthenticated, "无效的 Authorization 头格式，缺少 Bearer 前缀")
		}

		err := tokenSvc.ValidateToken(tokenString, secretKey)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "token 验证失败: %v", err)
		}

		return handler(srv, ss)
	}
}
