package log

import (
	"context"
	"fmt"
	"github.com/RealJonathanYip/framework"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
)

type loger struct {
	logger *zap.Logger
	Level  zap.AtomicLevel
}

const (
	LEVEL_DEBUG = "debug"
	LEVEL_INFO  = "info"
	LEVEL_WARN  = "warn"
	LEVEL_ERROR = "error"
	LEVEL_PANIC = "panic"
	LEVEL_FATAL = "fatal"
)

type LogWatcher interface {
	OnMessage(logLevel, log string)
}

var (
	defaultLog  *loger = nil
	stdlog      *log.Logger
	once        = &sync.Once{}
	_aryWatcher []LogWatcher
	traceSkip   = 2
)

// 启动后自动初始化日志对象
func init() {
	//初始化日志为stdout
	procname := path.Base(os.Args[0])

	InitLog(processName(procname), SetTarget("asyncfile"), LogFilePath("./logs"), LogFileRotate("date"))
}

// 根据options的设置,再次初始化日志系统。
func InitLog(options ...logOption) {
	var (
		err    error
		logger *zap.Logger
		level  zap.AtomicLevel
	)
	config := defaultLogOptions
	for _, option := range options {
		option.apply(&config)
	}

	if level, logger, err = zapLogInit(&config); err != nil {
		Panicf(context.TODO(), "ZapLogInit err:%v", err)
	}

	logger = logger.WithOptions(zap.AddCallerSkip(traceSkip))
	if defaultLog == nil {
		defaultLog = &loger{logger, level}
	} else {
		defaultLog.logger = logger
		defaultLog.Level = level
	}
	// redirect go log to defaultLog.logger
	zap.RedirectStdLog(defaultLog.logger)
	stdlog = zap.NewStdLog(defaultLog.logger)
}

func GetLogger() *loger {
	return defaultLog
}

func GetStdLogger() *log.Logger {
	return stdlog
}

func (l *loger) GetZLog(opts ...zap.Option) *zap.Logger {
	return l.logger.WithOptions(opts...)
}

func AddWatcher(ptrWatcher LogWatcher) {
	_aryWatcher = append(_aryWatcher, ptrWatcher)
}

// Example : GetLogger().Clone(zap.AddCallerSkip(2))
func (l *loger) Clone(opts ...zap.Option) *loger {
	nl := &loger{
		logger: l.logger,
		Level:  l.Level,
	}

	nl.logger = l.logger.WithOptions(opts...)
	return nl
}

// implement Write for io.Writer
func (l *loger) Write(ctx context.Context, p []byte) (n int, err error) {
	l.writelog(ctx, "info", string(p))
	return len(p), nil
}

// implement Log for github.com/go-log/log.logger
func (l *loger) Log(ctx context.Context, v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.writelog(ctx, "info", msg)
}

// implement Logf for github.com/go-log/log.logger
func (l *loger) Logf(ctx context.Context, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.writelog(ctx, "info", msg)
}

func (l *loger) writelog(ctx context.Context, level, msg string, fields ...zapcore.Field) {
	msg += framework.GetLogText(ctx)
	switch level {
	case "info":
		l.logger.Info(msg, fields...)
	case "debug":
		l.logger.Debug(msg, fields...)
	case "warn":
		l.logger.Warn(msg, fields...)
	case "error":
		l.logger.Error(msg, fields...)
	case "panic":
		l.logger.Panic(msg, fields...)
	case "dpanic":
		l.logger.DPanic(msg, fields...)
	case "fatal":
		l.logger.Fatal(msg, fields...)
	default:
		l.logger.Info(msg, fields...)
	}

	for _, ptrWatcher := range _aryWatcher {
		ptrWatcher.OnMessage(level, msg)
	}
}

// level : debug info warn error panic fatal
func Log(ctx context.Context, level string, v ...interface{}) {
	msg := fmt.Sprint(v...)
	defaultLog.writelog(ctx, level, msg)
}

// level : debug info warn error panic fatal
func LogF(ctx context.Context, szLevel string, szFormat string, v ...interface{}) {
	funcName, _, _, ok := runtime.Caller(1)
	szFuncName := runtime.FuncForPC(funcName).Name()
	szFormatFinal := ""
	if ok {
		szFormatFinal = szFuncName + " " + szFormat
	}
	msg := fmt.Sprintf(szFormatFinal, v...)

	defaultLog.writelog(ctx, szLevel, msg)
}

func LogMyF(ctx context.Context, skip int, szLevel string, szFormat string, v ...interface{}) {
	funcName, _, _, ok := runtime.Caller(skip)
	szFuncName := runtime.FuncForPC(funcName).Name()
	szFormatFinal := ""
	if ok {
		strings.Split(szFuncName, "/")

		szFormatFinal = szFuncName + " " + szFormat
	}
	msg := fmt.Sprintf(szFormatFinal, v...)

	defaultLog.writelog(ctx, szLevel, msg)
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Debug(ctx context.Context, msg string, fields ...zapcore.Field) {
	defaultLog.writelog(ctx, "debug", msg, fields...)
}

func Debugf(ctx context.Context, szFormat string, v ...interface{}) {
	msg := fmt.Sprintf(szFormat, v...)
	defaultLog.writelog(ctx, "debug", msg)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Info(ctx context.Context, msg string, fields ...zapcore.Field) {
	defaultLog.writelog(ctx, "info", msg, fields...)
}

func Infof(ctx context.Context, szFormat string, v ...interface{}) {
	msg := fmt.Sprintf(szFormat, v...)
	defaultLog.writelog(ctx, "info", msg)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Warn(ctx context.Context, msg string, fields ...zapcore.Field) {
	defaultLog.writelog(ctx, "warn", msg, fields...)
}

func Warningf(ctx context.Context, szFormat string, v ...interface{}) {
	msg := fmt.Sprintf(szFormat, v...)
	defaultLog.writelog(ctx, "warn", msg)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Error(ctx context.Context, msg string, fields ...zapcore.Field) {
	defaultLog.writelog(ctx, "error", msg, fields...)
}

func Errorf(ctx context.Context, szFormat string, v ...interface{}) {
	msg := fmt.Sprintf(szFormat, v...)
	defaultLog.writelog(ctx, "error", msg)
}

// DPanic logs a message at DPanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func DPanic(ctx context.Context, msg string, fields ...zapcore.Field) {
	defaultLog.writelog(ctx, "dpanic", msg, fields...)
}

func DPanicf(ctx context.Context, szFormat string, v ...interface{}) {
	msg := fmt.Sprintf(szFormat, v...)
	defaultLog.writelog(ctx, "dpanic", msg)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func Panic(ctx context.Context, msg string, fields ...zapcore.Field) {
	defaultLog.writelog(ctx, "panic", msg, fields...)
}

func Panicf(ctx context.Context, szFormat string, v ...interface{}) {
	msg := fmt.Sprintf(szFormat, v...)
	defaultLog.writelog(ctx, "panic", msg)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is disabled.
func Fatal(ctx context.Context, msg string, fields ...zapcore.Field) {
	defaultLog.writelog(ctx, "fatal", msg, fields...)
}

func Fatalf(ctx context.Context, szFormat string, v ...interface{}) {
	msg := fmt.Sprintf(szFormat, v...)
	defaultLog.writelog(ctx, "fatal", msg)
}

func CloudLog(ctx context.Context, fields ...zapcore.Field) {
	defaultLog.writelog(ctx, "info", "CloudLog", fields...)
}

func Sync() error {
	return defaultLog.logger.Sync()
}

func SetLogLevel(level string) {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error", "fatal":
		level = strings.ToLower(level)
	case "all":
		level = "debug"
	case "off", "none":
		level = "fatal"
	default:
		panic("not support level")
	}

	defaultLog.Level.UnmarshalText([]byte(level))
}
