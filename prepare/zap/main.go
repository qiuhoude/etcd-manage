package main

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
	"time"
)

/*
zap 没有提供归档功能

官方推荐的是 natefinch/lumberjack 进归档
*/
func main() {
	//zap1()
	//zapCritical()
	//zapCfg()
	//zapArchive()
	s, _ := filepath.Abs(".")
	fmt.Println(s)
}

func zapArchive() {
	logpath := "d://info.log"

	l := io.MultiWriter(&lumberjack.Logger{
		Filename:   logpath, // 日志文件路径
		MaxSize:    1024,    // megabytes
		MaxBackups: 3,       // 最多保留3个备份
		MaxAge:     7,       // days
		Compress:   true,    // 是否压缩 disabled by default
	}, os.Stdout)
	w := zapcore.AddSync(l)

	// https://github.com/uber-go/zap/blob/master/FAQ.md
	encoderConfig := zap.NewProductionEncoderConfig()
	//encoderConfig.EncodeTime =  zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.00"))
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		w,
		zap.InfoLevel,
	)

	logger := zap.New(core)
	//sugar := logger.Sugar()
	//logger.Info("DefaultLogger init succesaas")

	// 打印
	logger.Info("test log", zap.Int("line", 47))
}

func zapCfg() {
	// zap的配置
	rawJSON := []byte(`{
	  "level": "debug",
	  "encoding": "json",
	  "outputPaths": ["stdout", "./demo.log"],
	  "errorOutputPaths": ["stderr"],
	  "initialFields": {"foo": "bar"},
	  "encoderConfig": {
	    "messageKey": "message",
	    "levelKey": "level",
	    "levelEncoder": "lowercase"
	  }
	}`)

	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	logger, _ := cfg.Build()
	defer logger.Sync()

	logger.Info("logger construction succeeded")
}

// 不是很严格的情况下使用 SugaredLogger ,效率高效
func zap1() {
	logger, _ := zap.NewProduction()
	url := `http://www.google.com`
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	sugar.Infow("failed to fetch URL",
		// Structured context as loosely typed key-value pairs.
		"url", url,
		"attempt", 3,
		"backoff", time.Second,
	)
	sugar.Infof("Failed to fetch URL: %s", url)
}

// 严格模式 使用这个
func zapCritical() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	url := `http://www.google.com`
	logger.Info("failed to fetch URL",
		// Structured context as strongly typed Field values.
		zap.String("url", url),
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second),
	)
}
