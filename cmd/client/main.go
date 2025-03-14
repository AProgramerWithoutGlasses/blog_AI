package main

import (
	"context"
	"fmt"
	"log"
	pbcode "siwuai/proto/code"
	"time"

	"google.golang.org/grpc"
)

// 用于模拟客户端通过gRPC调用LLM服务
func main() {
	// 连接到 gRPC 服务器（假设运行在 localhost:50051）
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("无法连接到服务器: %v", err)
	}
	defer conn.Close()

	// 设置请求上下文和超时
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	client1 := pbcode.NewCodeServiceClient(conn)

	req1 := &pbcode.CodeRequest{
		CodeQuestion: "你是谁创造的",
		UserId:       2,
	}

	resp1, err := client1.ExplainCode(ctx, req1)
	if err != nil {
		fmt.Println("client1.ExplainCode()", err)
	}

	fmt.Println("resp1:", resp1)

}
