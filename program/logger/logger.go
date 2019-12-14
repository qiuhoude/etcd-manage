package logger

import (
	"github.com/qiuhoude/etcd-manage/program/common"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"
)

// 日志对象
var (
	Log *zap.SugaredLogger
)

func InitLogger(logPath string, isDebug bool) (*zap.SugaredLogger, error) {

	infoLogPath := ""
	// errorLogPath := ""
	if logPath == "" {
		logRoot := common.GetRootDir() + "logs" + string(os.PathSeparator)
		if isExt, _ := common.PathExists(logRoot); isExt == false {
			os.MkdirAll(logRoot, os.ModePerm)
		}
		infoLogPath = logRoot + time.Now().Format("20060102") + ".log"
		// errorLogPath = logRoot + time.Now().Format("20060102_error") + ".log"
	} else {
		logPath = strings.TrimRight(logPath, string(os.PathSeparator))
		infoLogPath = logPath + string(os.PathSeparator) + time.Now().Format("20060102") + ".log"
		// errorLogPath = logPath + string(os.PathSeparator) + time.Now().Format("20060102_error") + ".log"
	}

	// 兼容win根完整路径问题
	_ = zap.RegisterSink("winfile", func(u *url.URL) (zap.Sink, error) {
		return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	})

	// zap的配置
	cfg := &zap.Config{
		Encoding: "json",
	}
	cfg.EncoderConfig = zap.NewProductionEncoderConfig()
	atom := zap.NewAtomicLevel()
	if isDebug == true {
		atom.SetLevel(zapcore.DebugLevel)
		cfg.OutputPaths = []string{"stdout"}
		// cfg.ErrorOutputPaths = []string{"stdout"}
	} else {
		atom.SetLevel(zapcore.InfoLevel)
		if runtime.GOOS == "windows" {
			cfg.OutputPaths = []string{"winfile:///" + infoLogPath}
		} else {
			cfg.OutputPaths = []string{infoLogPath}
		}
		// cfg.ErrorOutputPaths = []string{errorLogPath}
	}
	cfg.Level = atom

	// 构建logger
	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}
	// 使用非严格模式,提高效率
	Log = logger.Sugar()
	return Log, nil
}
