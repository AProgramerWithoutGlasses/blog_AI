package main

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"

	pb "grpc-ddd-demo/proto/llm"
)

// 用于模拟客户端通过gRPC调用LLM服务
func main() {
	// 连接到 gRPC 服务器（假设运行在 localhost:50051）
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("无法连接到服务器: %v", err)
	}
	defer conn.Close()

	// 创建 LLMService 客户端
	client := pb.NewLLMServiceClient(conn)

	// 设置请求上下文和超时
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// 构造请求
	req := &pb.GenerateRequest{
		Prompt: "下午好",
	}

	// 调用 Generate 接口
	resp, err := client.Generate(ctx, req)
	if err != nil {
		log.Fatalf("调用 Generate 接口失败: %v", err)
	}

	log.Printf("LLM生成的回复: %s", resp.Completion)
}
