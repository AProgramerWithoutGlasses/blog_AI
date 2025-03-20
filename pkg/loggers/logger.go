package loggers

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"siwuai/internal/infrastructure/config"
	"time"
)

func LogInit(cfg config.Config) *zap.SugaredLogger {
	writeSyncer := GetLogWriter(cfg.Logger.LogPath, cfg.Logger.AppName)
	encoder := GetEncoder()

	// 新增部分：将日志输出到文件
	fileCore := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)

	// 文件输出
	core := zapcore.NewTee(fileCore)

	log := zap.New(core, zap.AddCaller())

	return log.Sugar()
}

// GetEncoder 获取编码器
func GetEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

// GetLogWriter 获取日志写入器
func GetLogWriter(logPath, appName string) zapcore.WriteSyncer {
	// 确保日志目录存在
	if err := os.MkdirAll(logPath, os.ModePerm); err != nil {
		fmt.Printf("failed to create log directory: %v\n", err)
		return nil
	}

	currentDate := time.Now().Format("2006-01-02")
	fileName := fmt.Sprintf("./%s/%s-%s.log", logPath, appName, currentDate)
	file, _ := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	return zapcore.AddSync(file)
}
