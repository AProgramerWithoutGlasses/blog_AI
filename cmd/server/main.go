package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"siwuai/internal/infrastructure/config"
	"siwuai/internal/infrastructure/etcd"
	"siwuai/internal/infrastructure/grpc"
	mysqlInfra "siwuai/internal/infrastructure/persistence"
)

func main() {
	// 加载配置文件
	cfg, err := config.LoadConfig(".")
	if err != nil {
		log.Fatalf("加载配置文件失败: %v", err)
	}

	// 封装 MySQL 初始化
	db, err := mysqlInfra.NewMySQLDB(cfg)
	if err != nil {
		log.Fatalf("初始化 MySQL 失败: %v", err)
	}

	// etcd 注册初始化，使用配置文件中的 etcd 配置
	etcdCfg := cfg.Etcd
	etcdRegistry, err := etcd.NewEtcdRegistry(etcdCfg.Endpoints, etcdCfg.ServiceName, etcdCfg.ServiceAddr, etcdCfg.TTL)
	if err != nil {
		log.Fatalf("创建 etcd 注册器失败: %v", err)
	}

	// 创建上下文控制 etcd 注册生命周期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 注册服务到 etcd
	if err = etcdRegistry.Register(ctx); err != nil {
		log.Fatalf("服务注册到 etcd 失败: %v", err)
	}
	log.Println("服务成功注册到 etcd")

	// 优雅退出：捕获退出信号时注销服务并关闭 etcd 客户端
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("接收到退出信号，开始注销服务...")
		if err := etcdRegistry.Deregister(ctx); err != nil {
			log.Printf("注销服务失败: %v", err)
		}
		etcdRegistry.Close()
		cancel()
		os.Exit(0)
	}()

	// 启动 gRPC 服务，使用配置文件中指定的端口（例如：cfg.Server.Port）
	port := cfg.Server.Port
	log.Printf("gRPC 服务器启动在端口 %s...", port)
	if err = grpc.RunGRPCServer(port, db); err != nil {
		log.Fatalf("启动 gRPC 服务器失败: %v", err)
	}
}
