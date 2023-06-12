package logger

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func CoreLogInit() {
	InitLogger()
}

var Logger *zap.SugaredLogger

func Errorf(template string, args ...interface{}) {
	Logger.Errorf(template, args)
}

func Infof(template string, args ...interface{}) {
	Logger.Infof(template, args)
}

func Info(template interface{}) {
	Logger.Info(template)
}

func Error(template interface{}) {
	Logger.Error(template)
}

func InitLogger() {
	writeSyncer := getLogWriter()
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)

	logger := zap.New(core, zap.AddCaller())
	Logger = logger.Sugar()
	fmt.Println("Logger init done")
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	fileName := "./log/log.log"
	lumberJackLogger := &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    2028,
		MaxBackups: 2,
		MaxAge:     30,
		Compress:   false,
	}
	return zapcore.AddSync(lumberJackLogger)
}
