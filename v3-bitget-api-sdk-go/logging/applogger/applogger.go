package applogger

import "go.uber.org/zap"

func Fatal(template string, args ...any) {
	zap.S().Fatalf(template, args...)
}

func Error(template string, args ...any) {
	zap.S().Errorf(template, args...)
}

func Panic(template string, args ...any) {
	zap.S().Panicf(template, args...)
}

func Warn(template string, args ...any) {
	zap.S().Warnf(template, args...)
}

func Info(template string, args ...any) {
	zap.S().Debugf(template, args...)
}

func Debug(template string, args ...any) {
	zap.S().Debugf(template, args...)
}
