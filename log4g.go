package log4g

import (
	"fmt"
	"go.uber.org/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"strconv"
	"time"
)

var sugar *zap.SugaredLogger

// getLumberjackLogger 获取lumberjack.Logger
func getLumberjackLogger(root *config.YAML, level string) *lumberjack.Logger {
	option := root.Get("LOG4G").Get(level)
	filename := option.Get("FILE_PATH_NAME").String()
	maxSize, err := strconv.Atoi(option.Get("MAXSIZE").String())
	if err != nil {
		log.Fatalln(fmt.Sprintf("%s MAXSIZE参数不为数字或者参数不合法", level))
	}
	maxBackups, err := strconv.Atoi(option.Get("MAXBACKUP_COUNT").String())
	if err != nil {
		log.Fatalln(fmt.Sprintf("%s MAXBACKUP_COUNT参数不为数字或者参数不合法", level))
	}
	maxAge, err := strconv.Atoi(option.Get("MAXAGE").String())
	if err != nil {
		log.Fatalln(fmt.Sprintf("%s MAXAGE参数不为数字或者参数不合法", level))
	}
	compress, err := strconv.ParseBool(option.Get("COMPRESS").String())
	if err != nil {
		log.Fatalln(fmt.Sprintf("%s COMPRESS参数不为数字或者参数不合法", level))
	}
	return &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize, // megabytes
		MaxBackups: maxBackups,
		MaxAge:     maxAge,   // days
		Compress:   compress, // disabled by default
	}
}

func init() {
	var options config.YAMLOption = config.File("log4g.yml")
	root, _ := config.NewYAML(options)
	log4gMode := root.Get("LOG4G").Get("MODE").String()
	infoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		switch log4gMode {
		case "contain":
			return lvl >= zapcore.InfoLevel
		case "independent":
			return lvl >= zapcore.InfoLevel && lvl < zapcore.ErrorLevel
		default:
			return lvl >= zapcore.InfoLevel
		}
	})
	// 读取Product配置
	logConfig := zap.NewProductionEncoderConfig()
	logConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006.01.02 15:04:05"))
	}
	logConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	errorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})

	hookInfo := getLumberjackLogger(root, "INFO")
	defer func(hookInfo *lumberjack.Logger) {
		if err := hookInfo.Close(); err != nil {
			log.Fatalln(err)
		}
	}(hookInfo)

	hookError := getLumberjackLogger(root, "ERROR")
	defer func(hookError *lumberjack.Logger) {
		if err := hookError.Close(); err != nil {
			log.Fatalln(err)
		}
	}(hookError)

	// 最后创建具体的Logger
	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewJSONEncoder(logConfig), zapcore.AddSync(hookInfo), infoLevel),
		zapcore.NewCore(zapcore.NewJSONEncoder(logConfig), zapcore.AddSync(hookError), errorLevel),
	)
	logger := zap.New(core, zap.AddCaller())
	sugar = logger.Sugar()
}

func Info(args ...interface{}) {
	sugar.Info(args...)
}

func Infof(temp string, args ...interface{}) {
	sugar.Infof(temp, args...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	sugar.Infow(msg, keysAndValues...)
}

func Warn(args ...interface{}) {
	sugar.Warn(args...)
}

func Warnf(temp string, args ...interface{}) {
	sugar.Warnf(temp, args...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	sugar.Warnw(msg, keysAndValues...)
}

func Error(args ...interface{}) {
	sugar.Error(args...)
}

func Errorf(temp string, args ...interface{}) {
	sugar.Errorf(temp, args...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	sugar.Errorw(msg, keysAndValues...)
}

func Fatal(args ...interface{}) {
	sugar.Fatal(args...)
}

func Fatalf(temp string, args ...interface{}) {
	sugar.Fatalf(temp, args...)
}

func Fatalw(msg string, keysAndValues ...interface{}) {
	sugar.Fatalw(msg, keysAndValues...)
}
