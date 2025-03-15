package main

import (
	"context"
	"fmt"
	"log"
	"siwuai/internal/infrastructure/config"
	"siwuai/internal/infrastructure/etcd"
	pbcode "siwuai/proto/code"
	"time"

	"google.golang.org/grpc"

	pbllm "siwuai/proto/llm"
)

// 用于模拟客户端通过gRPC调用LLM服务
func main() {
	// 加载配置文件
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// etcd 注册初始化，使用配置文件中的 etcd 配置
	etcdCfg := cfg.Etcd
	registry, err := etcd.NewEtcdRegistry(etcdCfg.Endpoints, etcdCfg.ServiceName, etcdCfg.ServiceAddr, etcdCfg.TTL)
	if err != nil {
		log.Fatalf("创建 etcd 注册器失败: %v", err)
	}

	// 服务发现
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addrs, err := registry.Discover(ctx, cfg.Etcd.ServiceName)
	if err != nil {
		log.Fatalf("服务发现失败: %v", err)
	}
	if len(addrs) == 0 {
		log.Fatalf("未找到可用的服务实例")
	}

	// 连接到 gRPC 服务器（假设运行在 localhost:50051）
	conn, err := grpc.Dial(addrs[0], grpc.WithInsecure())
	if err != nil {
		log.Fatalf("无法连接到服务器: %v", err)
	}
	defer conn.Close()

	// 创建 LLMService 客户端
	client := pbllm.NewLLMServiceClient(conn)

	// 设置请求上下文和超时
	ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	// 构造请求
	req := &pbllm.GenerateRequest{
		Prompt: "下午好",
	}

	// 调用 Generate 接口
	resp, err := client.Generate(ctx, req)
	if err != nil {
		log.Fatalf("调用 Generate 接口失败: %v", err)
	}

	log.Printf("LLM生成的回复: %s", resp.Completion)

	// -----------------------------------
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
