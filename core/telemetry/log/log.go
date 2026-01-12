package log

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kweaver-ai/idrm-go-frame/core/telemetry"

	"github.com/kweaver-ai/idrm-go-frame/core/logx/zapx"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_log"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/public"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/resource"
	"github.com/kweaver-ai/TelemetrySDK-Go/span/v2/encoder"
	"github.com/kweaver-ai/TelemetrySDK-Go/span/v2/field"
	spanLog "github.com/kweaver-ai/TelemetrySDK-Go/span/v2/log"
	"github.com/kweaver-ai/TelemetrySDK-Go/span/v2/open_standard"
	"github.com/kweaver-ai/TelemetrySDK-Go/span/v2/runtime"
)

type logObject struct {
	logger    zapx.Logger
	arOnce    sync.Once
	zapxOnce  sync.Once
	ar_logger spanLog.Logger
}

var logObj *logObject = nil

type log interface {
	Debug(msg string, fields ...zapx.Field)
	Info(msg string, fields ...zapx.Field)
	Warn(msg string, fields ...zapx.Field)
	Error(msg string, fields ...zapx.Field)
	Fatal(msg string, fields ...zapx.Field)
	Trace(msg string, fields ...zapx.Field)

	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	Fatalf(format string, v ...interface{})

	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
}

type spanLogger struct {
	ctx context.Context
}

func WithContext(ctx context.Context) log {
	return &spanLogger{ctx: ctx}
}

// InitLogger  .
func InitLogger(cs []zapx.Options, tc *telemetry.Config) {
	if logObj != nil {
		return
	}
	logObj = new(logObject)
	if logObj.ar_logger == nil {
		logObj.arOnce.Do(func() {
			logObj.ar_logger = initARLog(tc)
		})
	}

	logObj.zapxOnce.Do(func() {
		skip := 1
		// 增加callerskip，因为telemetry/log对zapx进行了一次封装
		zapx.CustomLoads(skip, zapx.LogConfigs{
			Logs: cs,
		})
		//zapx = zapx.GetLogger(cs.Logs[0].Name)
	})

}

func initARLog(config *telemetry.Config) spanLog.Logger {
	// ARLogger 程序日志记录器，使用异步发送模式，无返回值。
	//spanLog.AllLevel
	var ARLogger = spanLog.NewSamplerLogger(spanLog.WithSample(1.0), spanLog.WithLevel(getLogLevel(config.LogLevel)))
	// 初始化ar_log
	public.SetServiceInfo(config.ServerName, config.ServerVersion, "")
	// 1.初始化系统日志器，系统日志在控制台输出，同时上报到AnyRobot。
	//systemLogClient := public.NewFileClient("")
	systemLogClient := public.NewHTTPClient(public.WithAnyRobotURL(config.LogUrl),
		public.WithCompression(1), public.WithTimeout(1*time.Second),
		public.WithRetry(true, 5*time.Second, 30*time.Second, 1*time.Minute))
	systemLogExporter := ar_log.NewExporter(systemLogClient)
	systemLogWriter := open_standard.OpenTelemetryWriter(
		encoder.NewJsonEncoderWithExporters(systemLogExporter),
		resource.LogResource())
	systemLogRunner := runtime.NewRuntime(systemLogWriter, field.NewSpanFromPool)
	systemLogRunner.SetUploadInternalAndMaxLog(3*time.Second, 10)
	// 运行SystemLogger日志器。
	go systemLogRunner.Run()

	ARLogger.SetLevel(getLogLevel(config.LogLevel))
	ARLogger.SetRuntime(systemLogRunner)
	return ARLogger
}

// getLogLevel Log配置转换为spanlog配置，默认不填的日志级别为error
func getLogLevel(level string) int {
	switch level {
	case telemetry.ALL:
		return spanLog.AllLevel
	case telemetry.TRACE:
		return spanLog.TraceLevel
	case telemetry.DEBUG:
		return spanLog.DebugLevel
	case telemetry.ERROR:
		return spanLog.ErrorLevel
	case telemetry.FATAL:
		return spanLog.FatalLevel
	case telemetry.WARN:
		return spanLog.WarnLevel
	case telemetry.INFO:
		return spanLog.InfoLevel
	case telemetry.OFF:
		return spanLog.OffLevel
	default:
		return spanLog.ErrorLevel
	}
}

func (s *spanLogger) Info(msg string, fields ...zapx.Field) {
	doInfo(s.ctx, msg, fields...)
}

func (s *spanLogger) Infof(format string, v ...interface{}) {
	doInfof(s.ctx, format, v...)
}

func (s *spanLogger) Infow(msg string, keysAndValues ...interface{}) {
	doInfow(s.ctx, msg, keysAndValues...)
}

func (s *spanLogger) Debug(msg string, fields ...zapx.Field) {
	doDebug(s.ctx, msg, fields...)
}

func (s *spanLogger) Debugf(format string, v ...interface{}) {
	doDebugf(s.ctx, format, v...)
}

func (s *spanLogger) Debugw(msg string, keysAndValues ...interface{}) {
	doDebugw(s.ctx, msg, keysAndValues)
}

func (s *spanLogger) Warn(msg string, fields ...zapx.Field) {
	doWarn(s.ctx, msg, fields...)
}

func (s *spanLogger) Warnf(format string, v ...interface{}) {
	doWarnf(s.ctx, format, v...)
}

func (s *spanLogger) Warnw(msg string, keysAndValues ...interface{}) {
	doWarnw(s.ctx, msg, keysAndValues...)
}

func (s *spanLogger) Error(msg string, fields ...zapx.Field) {
	doError(s.ctx, msg, fields...)
}

func (s *spanLogger) Errorf(format string, v ...interface{}) {
	doErrorf(s.ctx, format, v...)
}

func (s *spanLogger) Errorw(msg string, keysAndValues ...interface{}) {
	doErrorw(s.ctx, msg, keysAndValues...)
}

func (s *spanLogger) Panic(msg string, fields ...zapx.Field) {
	doPanic(s.ctx, msg, fields...)
}

func (s *spanLogger) Panicf(format string, v ...interface{}) {
	doPanicf(s.ctx, format, v...)
}

func (s *spanLogger) Panicw(msg string, keysAndValues ...interface{}) {
	doPanicw(s.ctx, msg, keysAndValues...)
}

func (s *spanLogger) Fatal(msg string, fields ...zapx.Field) {
	doFatal(s.ctx, msg, fields...)
}

func (s *spanLogger) Fatalf(format string, v ...interface{}) {
	doFatalf(s.ctx, format, v...)
}

func (s *spanLogger) Fatalw(msg string, keysAndValues ...interface{}) {
	doFatalw(s.ctx, msg, keysAndValues...)
}

func (s *spanLogger) Trace(msg string, fields ...zapx.Field) {
	if s.ctx != nil {
		logObj.ar_logger.Trace(msg, field.WithContext(s.ctx))
	} else {
		logObj.ar_logger.Trace(msg)
	}
	zapx.Info(msg, fields...)
}

func (s *spanLogger) Flush() {
	if logObj.ar_logger != nil {
		logObj.ar_logger.Close()
	}
	zapx.Flush()

}

func Info(msg string, fields ...zapx.Field) {
	doInfo(nil, msg, fields...)
}

func Infof(format string, v ...interface{}) {
	doInfof(nil, format, v...)
}

func Infow(msg string, keysAndValues ...interface{}) {
	doInfow(nil, msg, keysAndValues...)
}

func Debug(msg string, fields ...zapx.Field) {
	doDebug(nil, msg, fields...)
}

func Debugf(format string, v ...interface{}) {
	doDebugf(nil, format, v...)
}

func Debugw(msg string, keysAndValues ...interface{}) {
	doDebugw(nil, msg, keysAndValues)
}
func Warn(msg string, fields ...zapx.Field) {
	doWarn(nil, msg, fields...)
}

func Warnf(format string, v ...interface{}) {
	doWarnf(nil, format, v...)
}

func Warnw(msg string, keysAndValues ...interface{}) {
	doWarnw(nil, msg, keysAndValues...)
}

func Error(msg string, fields ...zapx.Field) {
	doError(nil, msg, fields...)
}

func Errorf(format string, v ...interface{}) {
	doErrorf(nil, format, v...)
}

func Errorw(msg string, keysAndValues ...interface{}) {
	doErrorw(nil, msg, keysAndValues...)
}

func Panic(msg string, fields ...zapx.Field) {
	doPanic(nil, msg, fields...)
}
func Panicf(format string, v ...interface{}) {
	doPanicf(nil, format, v...)
}
func Panicw(msg string, keysAndValues ...interface{}) {
	doPanicw(nil, msg, keysAndValues...)
}
func Fatal(msg string, fields ...zapx.Field) {
	doFatal(nil, msg, fields...)
}
func Fatalf(format string, v ...interface{}) {
	doFatalf(nil, format, v...)
}
func Fatalw(msg string, keysAndValues ...interface{}) {
	doFatalw(nil, msg, keysAndValues...)
}

func NewZapWriter(name string) *zapx.ZapWriter {
	return zapx.NewZapWriter(name)
}

func Flush() {
	zapx.Flush()
}

func doInfo(ctx context.Context, msg string, fields ...zapx.Field) {
	if ctx != nil {
		logObj.ar_logger.Info(msg, field.WithContext(ctx))
	} else {
		logObj.ar_logger.Info(msg)
	}
	zapx.Info(msg, fields...)
}

func doInfof(ctx context.Context, format string, v ...interface{}) {
	if ctx != nil {
		logObj.ar_logger.Info(fmt.Sprintf(format, v...), field.WithContext(ctx))
	} else {
		logObj.ar_logger.Info(fmt.Sprintf(format, v...))
	}
	zapx.Infof(format, v...)
}

func doInfow(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if ctx != nil {
		attr := field.NewAttribute("keysAndValues", field.MallocJsonField(keysAndValues))
		logObj.ar_logger.Info(msg, field.WithAttribute(attr), field.WithContext(ctx))
	} else {
		attr := field.NewAttribute("keysAndValues", field.MallocJsonField(keysAndValues))
		logObj.ar_logger.Info(msg, field.WithAttribute(attr))
	}
	zapx.Infow(msg, keysAndValues...)
}

func doDebugw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if ctx != nil {
		attr := field.NewAttribute("keysAndValues", field.MallocJsonField(keysAndValues))
		logObj.ar_logger.Debug(msg, field.WithAttribute(attr), field.WithContext(ctx))
	} else {
		attr := field.NewAttribute("keysAndValues", field.MallocJsonField(keysAndValues))
		logObj.ar_logger.Debug(msg, field.WithAttribute(attr))
	}
	zapx.Debugw(msg, keysAndValues)
}

func doDebug(ctx context.Context, msg string, fields ...zapx.Field) {
	if ctx != nil {
		logObj.ar_logger.Debug(msg, field.WithContext(ctx))
	} else {
		logObj.ar_logger.Debug(msg)
	}
	zapx.Debug(msg, fields...)
}

func doDebugf(ctx context.Context, format string, v ...interface{}) {
	if ctx != nil {
		logObj.ar_logger.Debug(fmt.Sprintf(format, v...), field.WithContext(ctx))
	} else {
		logObj.ar_logger.Debug(fmt.Sprintf(format, v...))
	}
	zapx.Debugf(format, v...)
}

func doWarn(ctx context.Context, msg string, fields ...zapx.Field) {
	if ctx != nil {
		logObj.ar_logger.Warn(msg, field.WithContext(ctx))
	} else {
		logObj.ar_logger.Warn(msg)
	}
	zapx.Warn(msg, fields...)
}

func doWarnf(ctx context.Context, format string, v ...interface{}) {
	if ctx != nil {
		logObj.ar_logger.Warn(fmt.Sprintf(format, v...), field.WithContext(ctx))
	} else {
		logObj.ar_logger.Warn(fmt.Sprintf(format, v...))
	}
	zapx.Warnf(format, v...)
}

func doWarnw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if ctx != nil {
		attr := field.NewAttribute("keysAndValues", field.MallocJsonField(keysAndValues))
		logObj.ar_logger.Warn(msg, field.WithAttribute(attr), field.WithContext(ctx))
	} else {
		attr := field.NewAttribute("keysAndValues", field.MallocJsonField(keysAndValues))
		logObj.ar_logger.Warn(msg, field.WithAttribute(attr))
	}
	zapx.Warnw(msg, keysAndValues...)
}

func doError(ctx context.Context, msg string, fields ...zapx.Field) {
	if ctx != nil {
		logObj.ar_logger.Error(msg, field.WithContext(ctx))
	} else {
		logObj.ar_logger.Error(msg)
	}
	zapx.Error(msg, fields...)
}

func doErrorf(ctx context.Context, format string, v ...interface{}) {
	if ctx != nil {
		logObj.ar_logger.Error(fmt.Sprintf(format, v...), field.WithContext(ctx))
	} else {
		logObj.ar_logger.Error(fmt.Sprintf(format, v...))
	}
	zapx.Errorf(format, v...)
}

func doErrorw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if ctx != nil {
		attr := field.NewAttribute("keysAndValues", field.MallocJsonField(keysAndValues))
		logObj.ar_logger.Error(msg, field.WithAttribute(attr), field.WithContext(ctx))
	} else {
		attr := field.NewAttribute("keysAndValues", field.MallocJsonField(keysAndValues))
		logObj.ar_logger.Error(msg, field.WithAttribute(attr))
	}
	zapx.Errorw(msg, keysAndValues...)
}

func doPanic(ctx context.Context, msg string, fields ...zapx.Field) {
	if ctx != nil {
		logObj.ar_logger.Error(msg, field.WithContext(ctx))
	} else {
		logObj.ar_logger.Error(msg)
	}
	zapx.Panic(msg, fields...)
}

func doPanicf(ctx context.Context, format string, v ...interface{}) {
	if ctx != nil {
		logObj.ar_logger.Error(fmt.Sprintf(format, v...), field.WithContext(ctx))
	} else {
		logObj.ar_logger.Error(fmt.Sprintf(format, v...))
	}
	zapx.Panicf(format, v...)
}

func doPanicw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if ctx != nil {
		attr := field.NewAttribute("keysAndValues", field.MallocJsonField(keysAndValues))
		logObj.ar_logger.Error(msg, field.WithAttribute(attr), field.WithContext(ctx))
	} else {
		attr := field.NewAttribute("keysAndValues", field.MallocJsonField(keysAndValues))
		logObj.ar_logger.Error(msg, field.WithAttribute(attr))
	}
	zapx.Panicw(msg, keysAndValues...)
}

func doFatal(ctx context.Context, msg string, fields ...zapx.Field) {
	if ctx != nil {
		logObj.ar_logger.Fatal(msg, field.WithContext(ctx))
	} else {
		logObj.ar_logger.Fatal(msg)
	}
	zapx.Fatal(msg, fields...)
}

func doFatalf(ctx context.Context, format string, v ...interface{}) {
	if ctx != nil {
		logObj.ar_logger.Fatal(fmt.Sprintf(format, v...), field.WithContext(ctx))
	} else {
		logObj.ar_logger.Fatal(fmt.Sprintf(format, v...))
	}
	zapx.Fatalf(format, v...)
}

func doFatalw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	if ctx != nil {
		attr := field.NewAttribute("keysAndValues", field.MallocJsonField(keysAndValues))
		logObj.ar_logger.Fatal(msg, field.WithAttribute(attr), field.WithContext(ctx))
	} else {
		attr := field.NewAttribute("keysAndValues", field.MallocJsonField(keysAndValues))
		logObj.ar_logger.Fatal(msg, field.WithAttribute(attr))
	}
	zapx.Fatalw(msg, keysAndValues...)
}
