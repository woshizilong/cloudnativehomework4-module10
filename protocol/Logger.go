package protocol

// Logger 日志组件需要实现的协议
type Logger interface {
	Info(message string)
	InfoI(message string, key string, i interface{})
	Warn(message string)
	WarnI(message string, key string, i interface{})
	Debug(message string)
	DebugI(message string, key string, i interface{})
	Error(message string, err error)
	ErrorI(message string, err error, key string, i interface{})
	// Fatal 严重问题（os.Exit(1)）
	Fatal(message string)
	// FatalI 严重问题（os.Exit(1)）
	FatalI(message string, key string, i interface{})
	// Panic 恐慌问题（os.Exit(1)）
	Panic(message string)
	// PanicI 恐慌问题（os.Exit(1)）
	PanicI(message string, key string, i interface{})
}
