package logger

import "fmt"

type (
	Logger = func(format string, args ...interface{})
)

var defaultLogger Logger = func(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func SetDefaultLogger(logger Logger) {
	defaultLogger = logger
}

func GetDefaultLogger() Logger {
	return defaultLogger
}
