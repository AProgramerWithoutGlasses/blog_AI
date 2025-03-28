package config

import (
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"log"
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
	Redis struct {
		Addr     string `mapstructure:"addr"`     // Redis 地址
		Password string `mapstructure:"password"` // Redis 密码
		DB       int    `mapstructure:"db"`       // Redis 数据库编号
		Timeout  int    `mapstructure:"timeout"`  // 操作超时时间（秒）
	} `mapstructure:"redis"`
	Logger struct {
		LogPath string `mapstructure:"logPath"` // 日志输出文件
		AppName string `mapstructure:"appName"` // 项目名称
		Level   int8   `mapstructure:"level"`
	} `mapstructure:"log"`
	Token struct {
		SecretKey        string `mapstructure:"secretKey"`        // token验证密钥
		GenerateTokenKey string `mapstructure:"generateTokenKey"` // token生成密钥
	} `mapstructure:"token"`
	Llm struct {
		ApiKey             string  `mapstructure:"apiKey"`
		Model              string  `mapstructure:"model"`
		BaseURL            string  `mapstructure:"baseURL"`
		TemperatureCode    float64 `mapstructure:"temperatureCode"`
		TemperatureArticle float64 `mapstructure:"temperatureArticle"`
	} `mapstructure:"llm"`
}

// LoadConfig 加载并解析配置文件
func LoadConfig(path string, name string) (config Config, err error) {
	viper.SetConfigName(name)
	viper.AddConfigPath(path)
	viper.SetConfigType("yaml")

	if err = viper.ReadInConfig(); err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("无法解析配置文件: %v", err)
	}

	successMsg := fmt.Sprintf("%s.yaml 初始化成功", name)
	fmt.Println(successMsg)
	zap.L().Info(successMsg)
	zap.L().Info("1111")
	return
}
