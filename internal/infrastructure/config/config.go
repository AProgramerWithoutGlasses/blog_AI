package config

import (
	"log"

	"github.com/spf13/viper"
)

// Config 定义了应用的配置结构体
type Config struct {
	Server struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"server"`
	MySQL struct {
		Host      string `mapstructure:"host"`
		Port      int    `mapstructure:"port"`
		Username  string `mapstructure:"username"`
		Password  string `mapstructure:"password"`
		DBName    string `mapstructure:"dbname"`
		ParseTime bool   `mapstructure:"parseTime"`
	} `mapstructure:"mysql"`
	Etcd struct {
		Endpoints   []string `mapstructure:"endpoints"`   // etcd 集群地址列表
		ServiceName string   `mapstructure:"serviceName"` // 服务名称，用于服务注册
		ServiceAddr string   `mapstructure:"serviceAddr"` // 服务地址，例如 "127.0.0.1:50051"
		TTL         int64    `mapstructure:"ttl"`         // 租约时间（秒）
	} `mapstructure:"etcd"`
}

// LoadConfig 加载并解析配置文件
func LoadConfig(path string) (config Config, err error) {
	viper.SetConfigName("local")
	viper.AddConfigPath(path)
	viper.SetConfigType("yaml")

	if err = viper.ReadInConfig(); err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("无法解析配置文件: %v", err)
	}
	return
}
