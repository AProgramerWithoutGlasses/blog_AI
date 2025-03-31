package etcd

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdRegistry 负责将服务注册到 etcd 中
type EtcdRegistry struct {
	client      *clientv3.Client
	leaseID     clientv3.LeaseID
	serviceName string
	serviceAddr string
	ttl         int64
}

// NewEtcdRegistry 创建一个新的 etcd 注册实例
func NewEtcdRegistry(endpoints []string, serviceName, serviceAddr string, ttl int64) (*EtcdRegistry, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		err = fmt.Errorf("clientv3.New() err: %v", err)
		return nil, err
	}
	return &EtcdRegistry{
		client:      cli,
		serviceName: serviceName,
		serviceAddr: serviceAddr,
		ttl:         ttl,
	}, nil
}

// Register 将服务注册到 etcd，并启动心跳续租
func (r *EtcdRegistry) Register(ctx context.Context) error {
	// 申请租约
	leaseResp, err := r.client.Grant(ctx, r.ttl)
	if err != nil {
		err = fmt.Errorf("client.Grant() err: %v", err)
		return err
	}
	r.leaseID = leaseResp.ID

	key := fmt.Sprintf("/%s", r.serviceName)
	_, err = r.client.Put(ctx, key, r.serviceAddr, clientv3.WithLease(r.leaseID))
	if err != nil {
		err = fmt.Errorf("client.Put() err: %v", err)
		return err
	}

	msg := fmt.Sprintf("etcd 服务注册成功: %s - %s", key, r.serviceAddr)
	fmt.Println(msg)
	zap.L().Info(msg)

	// 启动续租
	ch, err := r.client.KeepAlive(ctx, r.leaseID)
	if err != nil {
		err = fmt.Errorf("client.KeepAlive() err: %v", err)
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case ka, ok := <-ch:
				if !ok {
					// 租约可能过期
					return
				}
				// 这里可以记录续租日志：fmt.Printf("续租成功：%+v\n", ka)
				_ = ka
			}
		}
	}()

	return nil
}

// Deregister 注销服务
func (r *EtcdRegistry) Deregister(ctx context.Context) error {
	key := fmt.Sprintf("/%s/%s", r.serviceName, r.serviceAddr)
	_, err := r.client.Delete(ctx, key)
	if err != nil {
		err = fmt.Errorf("client.Delete() err: %v", err)
		return err
	}
	return nil
}

// Close 关闭 etcd 客户端连接
func (r *EtcdRegistry) Close() error {
	return r.client.Close()
}

// Discover 从 etcd 中发现指定服务的所有实例地址
func (r *EtcdRegistry) Discover(ctx context.Context, serviceName string) ([]string, error) {
	prefix := fmt.Sprintf("/%s/", serviceName)
	resp, err := r.client.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		err = fmt.Errorf("client.Get() err: %v", err)
		return nil, err
	}

	var addrs []string
	for _, kv := range resp.Kvs {
		addrs = append(addrs, string(kv.Value))
	}

	fmt.Println("etcd 发现服务:", addrs)
	return addrs, nil
}
