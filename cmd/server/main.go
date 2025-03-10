package main

import (
	"log"

	"grpc-ddd-demo/internal/infrastructure/config"
	"grpc-ddd-demo/internal/infrastructure/grpc"
	mysqlInfra "grpc-ddd-demo/internal/infrastructure/persistence"
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
	defer db.Close()

	// 启动 gRPC 服务，使用配置文件中指定的端口
	port := cfg.Server.Port
	log.Printf("gRPC server starting on port %s...", port)
	if err := grpc.RunGRPCServer(port, db); err != nil {
		log.Fatalf("启动 gRPC 服务器失败: %v", err)
	}
}
