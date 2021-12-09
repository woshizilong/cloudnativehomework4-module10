package logger

import (
	"runtime"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type LoggerProvider struct {
	level       string
	servicename string
}

// InitLogger 初始化Log部品
func NewLogger(level string, serviceName string) *LoggerProvider {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	l := &LoggerProvider{}
	l.level = level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		panic("log level NG")
	}
	zerolog.SetGlobalLevel(logLevel)
	l.servicename = serviceName

	log.Info().Msg("Logger init success on " + l.servicename)
	return l
}

// Info 输出普通日志
func (l *LoggerProvider) Info(message string) {
	funcName, line := l.getFuncInfo()
	l.addHeader(log.Info(), funcName, line).Msg(message)
}

// InfoI 输出普通日志
func (l *LoggerProvider) InfoI(message string, key string, i interface{}) {
	funcName, line := l.getFuncInfo()
	l.addHeader(log.Info(), funcName, line).Interface(key, i).Msg(message)
}

// Warn 输出警告日志
func (l *LoggerProvider) Warn(message string) {
	funcName, line := l.getFuncInfo()
	l.addHeader(log.Warn(), funcName, line).Timestamp().Msg(message)
}

// WarnI 输出警告日志
func (l *LoggerProvider) WarnI(message string, key string, i interface{}) {
	funcName, line := l.getFuncInfo()
	l.addHeader(log.Warn(), funcName, line).Interface(key, i).Msg(message)
}

// Debug 输出调试日志
func (l *LoggerProvider) Debug(message string) {
	funcName, line := l.getFuncInfo()
	l.addHeader(log.Debug(), funcName, line).Msg(message)
}

// DebugI 输出调试日志，含任意对象
func (l *LoggerProvider) DebugI(message string, key string, i interface{}) {
	funcName, line := l.getFuncInfo()
	l.addHeader(log.Debug(), funcName, line).Interface(key, i).Msg(message)
}

// Error 输入错误日志
func (l *LoggerProvider) Error(message string, err error) {
	funcName, line := l.getFuncInfo()
	l.addHeader(log.Err(err), funcName, line).Stack().Msg(message)
}

// ErrorI 输入错误日志
func (l *LoggerProvider) ErrorI(message string, err error, key string, i interface{}) {
	funcName, line := l.getFuncInfo()
	l.addHeader(log.Err(err), funcName, line).Stack().Interface(key, i).Msg(message)
}

// Fatal 严重问题（os.Exit(1)）
func (l *LoggerProvider) Fatal(message string) {
	funcName, line := l.getFuncInfo()
	l.addHeader(log.Fatal(), funcName, line).Msg(message)
}

// FatalI 严重问题（os.Exit(1)）
func (l *LoggerProvider) FatalI(message string, key string, i interface{}) {
	funcName, line := l.getFuncInfo()
	l.addHeader(log.Fatal(), funcName, line).Interface(key, i).Msg(message)
}

// Panic 恐慌问题（os.Exit(1)）
func (l *LoggerProvider) Panic(message string) {
	funcName, line := l.getFuncInfo()
	l.addHeader(log.Panic(), funcName, line).Msg(message)
}

// PanicI 恐慌问题（os.Exit(1)）
func (l *LoggerProvider) PanicI(message string, key string, i interface{}) {
	funcName, line := l.getFuncInfo()
	l.addHeader(log.Panic(), funcName, line).Interface(key, i).Msg(message)
}

func (l *LoggerProvider) addHeader(log *zerolog.Event, funcName string, line int) *zerolog.Event {
	// 暂时先不输出文件名和行号
	// if line > 0 {
	// 	return log.Timestamp().Str("func", funcName).Int("line", line)
	// }
	return log.Timestamp()
}

func (l *LoggerProvider) getFuncInfo() (string, int) {
	if pc, _, line, ok := runtime.Caller(2); ok {
		funcName := runtime.FuncForPC(pc).Name()
		return funcName, line
	}
	return "", 0
}
