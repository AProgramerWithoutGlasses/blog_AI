package loggers

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"siwuai/internal/infrastructure/config"
	"time"
)

func LogInit(cfg config.Config) {
	fmt.Printf("LogPath: %s, AppName: %s, Level: %d\n", cfg.Logger.LogPath, cfg.Logger.AppName, cfg.Logger.Level)
	writeSyncer := GetLogWriter(cfg.Logger.LogPath, cfg.Logger.AppName)
	if writeSyncer == nil {
		zap.L().Error("writeSyncer is nil, check GetLogWriter")
		//panic("writeSyncer is nil, check GetLogWriter")
	}

	encoder := GetEncoder()

	// 将日志输出到控制台
	consoleCore := zapcore.NewCore(encoder, zapcore.AddSync(zapcore.Lock(os.Stdout)), zapcore.Level(cfg.Logger.Level))

	// 将日志输出到文件
	fileCore := zapcore.NewCore(encoder, writeSyncer, zapcore.Level(cfg.Logger.Level))

	// 合并控制台输出和文件输出
	core := zapcore.NewTee(consoleCore, fileCore)

	// 只输出到文件
	// core := zapcore.NewTee(fileCore)

	//构建logger
	//zap.AddCaller()：
	//可选参数，启用调用者信息。
	//日志会包含"caller": "file.go:line"字段，显示日志调用的文件名和行号。
	logger := zap.New(core, zap.AddCaller())

	defer logger.Sync() // 添加 Sync

	// 替换全局zap
	zap.ReplaceGlobals(logger)

	// 替换全局log
	//log.SetOutput(zap.NewStdLog(logger).Writer())
	// 配置 log 不输出时间和文件行号
	//stdLogger := zap.NewStdLog(logger)
	//log.SetFlags(0) // 去掉时间、文件等前缀
	//log.SetOutput(stdLogger.Writer())
}

// GetEncoder 获取编码器
func GetEncoder() zapcore.Encoder {
	// 日志编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	// 使用自定义时间编码器
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	// 大小写编码器
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	// 根据encoderConfig创建一个JSON格式的编码器
	return zapcore.NewJSONEncoder(encoderConfig)
}

//func GetEncoder() zapcore.Encoder {
//	// 使用 Console 编码器（普通文本格式）
//	encoderConfig := zapcore.EncoderConfig{
//		TimeKey:        "time",                         // 时间字段名
//		LevelKey:       "level",                        // 级别字段名
//		CallerKey:      "caller",                       // 调用者字段名
//		MessageKey:     "msg",                          // 消息字段名
//		StacktraceKey:  "stacktrace",                   // 堆栈字段名
//		LineEnding:     zapcore.DefaultLineEnding,      // 换行符
//		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 级别小写
//		EncodeTime:     customTimeEncoder,              // ISO 8601 时间格式
//		EncodeDuration: zapcore.SecondsDurationEncoder, // 时长格式
//		EncodeCaller:   zapcore.ShortCallerEncoder,     // 短调用者格式
//	}
//	return zapcore.NewConsoleEncoder(encoderConfig)
//}

// GetLogWriter 获取日志写入器
func GetLogWriter(logPath, appName string) zapcore.WriteSyncer {
	if logPath == "" {
		zap.L().Error("logPath is empty")
		//panic("logPath is empty")
	}

	// 确保日志目录存在
	if err := os.MkdirAll(logPath, os.ModePerm); err != nil {
		zap.L().Error("failed to create log directory", zap.Error(err))
		//fmt.Printf("failed to create log directory: %v\n", err)
		return nil
	}

	currentDate := time.Now().Format("2006-01-02")
	fileName := fmt.Sprintf("./%s/%s-%s.log", logPath, appName, currentDate)
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		zap.L().Error("创建日志文件失败")
		//log.Fatal("创建日志文件失败")
	}
	return zapcore.AddSync(file)
}

// 自定义时间编码器，带时区
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	// 获取时区名称和偏移量（秒）
	zoneName, offset := t.Zone()
	// 将偏移量（秒）转为小时
	offsetHours := offset / 3600

	// 转换为中文时区描述
	var zoneStr string
	switch offsetHours {
	case 8:
		zoneStr = "东八区"
	case 7:
		zoneStr = "东七区"
	case -8:
		zoneStr = "西八区"
	default:
		zoneStr = fmt.Sprintf("%s%+d", zoneName, offsetHours) // 通用格式
	}

	// 格式化为“2025年3月19日 07时00分00秒123毫秒 东八区”
	timeStr := t.Format("2006年1月2日 15时04分05秒") + t.Format(".000")[1:] + "毫秒 " + zoneStr
	enc.AppendString(timeStr)
}
