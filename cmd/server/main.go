package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"siwuai/internal/infrastructure/loggers"
	"siwuai/internal/infrastructure/redis_utils"
	"siwuai/internal/infrastructure/utils"
	"syscall"

	"siwuai/internal/infrastructure/config"
	"siwuai/internal/infrastructure/etcd"
	"siwuai/internal/infrastructure/grpc"
	mysqlInfra "siwuai/internal/infrastructure/persistence"
)

func main() {
	// 加载配置文件
	cfg, err := config.LoadConfig("configs", "dev")
	if err != nil {
		fmt.Printf("加载配置文件失败: %v\n", err)
		return
	}

	// 初始化日志。
	loggers.LogInit(cfg)
	zap.L().Info(fmt.Sprintf("config初始化成功: %#v\n", cfg))

	// 封装 MySQL 初始化
	db, err := mysqlInfra.NewMySQLDB(cfg)
	if err != nil {
		zap.L().Error("初始化 MySQL 失败: %v", zap.Error(err))
		return
	}
	zap.L().Info("初始化 MySQL 成功")

	// 初始化 Redis
	redisClient, err := redis_utils.NewRedisClient(cfg)
	if err != nil {
		zap.L().Error("初始化 Redis 失败", zap.Error(err))
		return
	}
	defer redisClient.Close()
	zap.L().Info(" 初始化 Redis 成功")

	// 初始化布隆过滤器（假设预计存储 100 万条记录，误判率 0.01）
	bf, err := utils.LoadBloomFilter(db)
	if err != nil {
		zap.L().Error(fmt.Sprintf("加载布隆过滤器失败 utils.LoadBloomFilter() %v", err))
		return
	}

	// etcd 注册初始化，使用配置文件中的 etcd 配置
	etcdCfg := cfg.Etcd
	etcdRegistry, err := etcd.NewEtcdRegistry(etcdCfg.Endpoints, etcdCfg.ServiceName, etcdCfg.ServiceAddr, etcdCfg.TTL)
	if err != nil {
		zap.L().Error(fmt.Sprintf("创建 etcd 实例失败: %v", err))
		return
	}
	zap.L().Info("初始化 etcd 成功")

	// 创建上下文控制 etcd 注册生命周期
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 注册服务到 etcd
	if err = etcdRegistry.Register(ctx); err != nil {
		zap.L().Error(fmt.Sprintf("服务注册到 etcd 失败: %v", err))
		return
	}

	// 优雅退出：捕获退出信号时注销服务并关闭 etcd 客户端
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		zap.L().Info(fmt.Sprintln("接收到退出信号，开始注销etcd服务..."))
		if err = etcdRegistry.Deregister(ctx); err != nil {
			zap.L().Error(fmt.Sprintf("etcd 注销服务失败: %v", err))
			return
		}
		etcdRegistry.Close()
		cancel()
		os.Exit(0)
	}()

	// 启动 gRPC 服务，使用配置文件中指定的端口（例如：cfg.Server.Port）
	port := cfg.Server.Port
	if err = grpc.RunGRPCServer(port, db, redisClient, bf, cfg); err != nil {
		zap.L().Error(fmt.Sprintf("启动 gRPC 服务器失败: %v", err))
		return
	}

}
