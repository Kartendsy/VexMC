package logger

import (
	"fmt"
	"time"
)

const (
	ColorReset = "\033[0m"
	ColorDebug = "\033[36m"
	ColorInfo  = "\033[32m"
	ColorWarn  = "\033[33m"
	ColorError = "\033[31m"
	ColorTime  = "\033[90m"
)

func Debug(format string, v ...any) {
	logMessage("DEBUG", ColorDebug, format, v...)
}

func Info(format string, v ...any) {
	logMessage("INFO", ColorInfo, format, v...)
}

func Warn(format string, v ...any) {
	logMessage("WARN", ColorWarn, format, v...)
}

func Error(format string, v ...any) {
	logMessage("ERROR", ColorError, format, v...)
}

func logMessage(level, color, format string, v ...any) {
	timestamp := time.Now().Format("15:04:05")

	msg := fmt.Sprintf(format, v...)

	fmt.Printf("%s[%s]%s %s%-5s%s %s\n", ColorTime, timestamp, ColorReset, color, level, ColorReset, msg)
}
