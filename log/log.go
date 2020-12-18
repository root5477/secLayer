package log

import (
	"fmt"
	"strings"

	"github.com/cihub/seelog"
)

const (
	LevelOff = iota
	LevelCritical
	LevelError
	LevelWarn
	LevelInfo
	LevelDebug
	LevelTrace
)

var logLevel int = LevelTrace

// Tracef formats message according to format specifier
// and writes to default logger with log level = Trace.
func Tracef(format string, params ...interface{}) {
	if logLevel < LevelTrace {
		return
	}
	seelog.Trace(fmt.Sprintf(format, params...))
}

// Debugf formats message according to format specifier
// and writes to default logger with log level = Debug.
func Debugf(format string, params ...interface{}) {
	if logLevel < LevelDebug {
		return
	}
	seelog.Debug(fmt.Sprintf(format, params...))
}

// Infof formats message according to format specifier
// and writes to default logger with log level = Info.
func Infof(format string, params ...interface{}) {
	if logLevel < LevelInfo {
		return
	}
	seelog.Info(fmt.Sprintf(format, params...))
}

// Warnf formats message according to format specifier and writes to default logger with log level = Warn
func Warnf(format string, params ...interface{}) error {
	if logLevel < LevelWarn {
		return nil
	}
	return seelog.Warn(fmt.Sprintf(format, params...))
}

// Errorf formats message according to format specifier and writes to default logger with log level = Error
func Errorf(format string, params ...interface{}) error {
	if logLevel < LevelError {
		return nil
	}
	return seelog.Error(fmt.Sprintf(format, params...))
}

// Criticalf formats message according to format specifier and writes to default logger with log level = Critical
func Criticalf(format string, params ...interface{}) error {
	if logLevel < LevelCritical {
		return nil
	}
	return seelog.Critical(fmt.Sprintf(format, params...))
}

// Trace formats message using the default formats for its operands and writes to default logger with log level = Trace
func Trace(v ...interface{}) {
	if logLevel < LevelTrace {
		return
	}
	seelog.Trace(fmt.Sprint(v...))
}

// Debug formats message using the default formats for its operands and writes to default logger with log level = Debug
func Debug(v ...interface{}) {
	if logLevel < LevelDebug {
		return
	}
	seelog.Debug(fmt.Sprint(v...))
}

// Info formats message using the default formats for its operands and writes to default logger with log level = Info
func Info(v ...interface{}) {
	if logLevel < LevelInfo {
		return
	}
	seelog.Info(fmt.Sprint(v...))
}

// Warn formats message using the default formats for its operands and writes to default logger with log level = Warn
func Warn(v ...interface{}) error {
	if logLevel < LevelWarn {
		return nil
	}
	return seelog.Warn(fmt.Sprint(v...))
}

// Error formats message using the default formats for its operands and writes to default logger with log level = Error
func Error(v ...interface{}) error {
	if logLevel < LevelError {
		return nil
	}
	return seelog.Error(fmt.Sprint(v...))
}

// Critical formats message using the default formats for its operands and writes to default logger with log level = Critical
func Critical(v ...interface{}) error {
	if logLevel < LevelCritical {
		return nil
	}
	return seelog.Critical(fmt.Sprint(v...))
}

func Print(level int, v ...interface{}) {
	if logLevel < level {
		return
	}

	switch level {
	case LevelOff:
		return
	case LevelTrace:
		seelog.Trace(fmt.Sprint(v...))
	case LevelDebug:
		seelog.Debug(fmt.Sprint(v...))
	case LevelInfo:
		seelog.Info(fmt.Sprint(v...))
	case LevelWarn:
		seelog.Warn(fmt.Sprint(v...))
	case LevelError:
		seelog.Error(fmt.Sprint(v...))
	case LevelCritical:
		seelog.Critical(fmt.Sprint(v...))
	default:
		seelog.Warn(fmt.Sprintf("unknown log level: %v, print as warning", level))
		seelog.Warn(fmt.Sprint(v...))
	}
}

func Printf(level int, format string, params ...interface{}) {
	if logLevel < level {
		return
	}

	switch level {
	case LevelOff:
		return
	case LevelTrace:
		seelog.Trace(fmt.Sprintf(format, params...))
	case LevelDebug:
		seelog.Debug(fmt.Sprintf(format, params...))
	case LevelInfo:
		seelog.Info(fmt.Sprintf(format, params...))
	case LevelWarn:
		seelog.Warn(fmt.Sprintf(format, params...))
	case LevelError:
		seelog.Error(fmt.Sprintf(format, params...))
	case LevelCritical:
		seelog.Critical(fmt.Sprintf(format, params...))
	default:
		seelog.Warn(fmt.Sprintf("unknown log level: %v, print as warning", level))
		seelog.Warn(fmt.Sprintf(format, params...))
	}
}

func Flush() {
	seelog.Flush()
}

func ReplaceLogger(logger seelog.LoggerInterface) error {
	err := logger.SetAdditionalStackDepth(1)
	if err != nil {
		return err
	}
	return seelog.ReplaceLogger(logger)
}

func Init(configFile, minLevel string) error {
	// 自定义格式 不能与默认格式的前缀相同 会优先匹配默认格式
	err := seelog.RegisterCustomFormatter("ExShortFunc", ExShortFunctionNameFormatterFuncCreator)
	if err != nil {
		return err
	}

	logger, err := seelog.LoggerFromConfigAsFile(configFile)
	if err != nil {
		return err
	}

	err = ReplaceLogger(logger)
	if err != nil {
		return err
	}

	logLevel = LogLevel(minLevel)

	Trace("Trace log print begin")
	Debug("Debug log print begin")
	Info("Info log print begin")
	Warn("Warn log print begin")

	return nil
}

func SetLogLevel(level string) {
	logLevel = LogLevel(level)
	return
}

func GetLogLevel() string {
	switch logLevel {
	case LevelTrace:
		return "trace"
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelCritical:
		return "critical"
	default:
		return "unknown"
	}
}

func LogLevel(lv string) int {
	switch strings.ToLower(lv) {
	case "cri", "critical":
		return LevelCritical
	case "err", "error":
		return LevelError
	case "warn", "warning":
		return LevelWarn
	case "inf", "info":
		return LevelInfo
	case "debug":
		return LevelDebug
	case "trace":
		return LevelTrace
	default:
		fmt.Printf("unknown log level: %v, use level trace\n", lv)
		return LevelTrace
	}
}

func ExShortFunctionNameFormatterFuncCreator(unused string) seelog.FormatterFunc {
	return func(message string, level seelog.LogLevel, context seelog.LogContextInterface) interface{} {
		f := context.Func()
		spl := strings.Split(f, "/")
		return spl[len(spl)-1]
	}
}
