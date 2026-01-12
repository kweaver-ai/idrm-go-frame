package trace

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/kweaver-ai/idrm-go-frame/core/telemetry"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"

	"context"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/public"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	log "github.com/kweaver-ai/idrm-go-frame/core/logx/zapx"
)

const (
	HTTP_METHOD    = "http.method"
	HTTP_ROUTE     = "http.route"
	HTTP_CLIENT_IP = "http.client_ip"
	FUNC_PATH      = "func.path"
)

func OpenSpan(ctx *context.Context) trace.Span {
	pc, file, linkNo, ok := runtime.Caller(1)
	if !ok {
		log.Error("start span error")
	}
	funcPaths := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	spanName := funcPaths[len(funcPaths)-1]
	newCtx, span := ar_trace.Tracer.Start(*ctx, spanName)
	span.SetAttributes(attribute.String("func path", fmt.Sprintf("%s:%v", file, linkNo)))
	ctx = &newCtx
	return span
}

func CloseSpan(span trace.Span, errs ...error) {
	var err error = nil
	if len(errs) > 0 {
		err = errs[0]
	}
	CloseSpan(span, err)
}

// StartInternalSpan 内部方法调用
func StartInternalSpan(ctx context.Context) (context.Context, trace.Span) {
	//if c, ok := ctx.(*gin.Context); ok {
	//	ctx = c.Request.Context()
	//}
	pc, file, linkNo, ok := runtime.Caller(1)
	if !ok {
		log.Error("start span error")
		newCtx, span := ar_trace.Tracer.Start(ctx, "unknow", trace.WithSpanKind(trace.SpanKindInternal))
		return newCtx, span
	}
	funcPaths := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	spanName := funcPaths[len(funcPaths)-1]
	newCtx, span := ar_trace.Tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindInternal))
	span.SetAttributes(attribute.String(FUNC_PATH, fmt.Sprintf("%s:%v", file, linkNo)))
	return newCtx, span
}

func StartSpan(ctx context.Context) context.Context {
	pc, file, linkNo, ok := runtime.Caller(1)
	if !ok {
		log.Error("start span error")
	}
	funcPaths := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	spanName := funcPaths[len(funcPaths)-1]
	newCtx, span := ar_trace.Tracer.Start(ctx, spanName)
	span.SetAttributes(attribute.String("func path", fmt.Sprintf("%s:%v", file, linkNo)))
	return newCtx
}

// StartServerSpan 接口层调用时使用
func StartServerSpan(c *gin.Context) (ctx context.Context, span trace.Span) {
	newCtx := context.Background()
	for key, val := range c.Keys {
		newCtx = context.WithValue(newCtx, key, val)
	}
	ctx = otel.GetTextMapPropagator().Extract(newCtx, propagation.HeaderCarrier(c.Request.Header))
	ctx, span = ar_trace.Tracer.Start(ctx, c.FullPath(), trace.WithSpanKind(trace.SpanKindServer))
	span.SetAttributes(attribute.String(HTTP_METHOD, c.Request.Method))
	span.SetAttributes(attribute.String(HTTP_ROUTE, c.FullPath()))
	span.SetAttributes(attribute.String(HTTP_CLIENT_IP, c.ClientIP()))
	return ctx, span
}

// StartConsumerSpan 消费者消费消息时记录使用
func StartConsumerSpan(ctx context.Context) (context.Context, trace.Span) {
	pc, file, linkNo, ok := runtime.Caller(1)
	if !ok {
		log.Error("start span error")
		newCtx, span := ar_trace.Tracer.Start(ctx, "unknow", trace.WithSpanKind(trace.SpanKindConsumer))
		return newCtx, span
	}

	funcPaths := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	spanName := funcPaths[len(funcPaths)-1]
	newCtx, span := ar_trace.Tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindConsumer))
	span.SetAttributes(attribute.String(FUNC_PATH, fmt.Sprintf("%s:%v", file, linkNo)))

	return newCtx, span
}

// StartProducerSpan 生产者生产消息时记录使用
func StartProducerSpan(ctx context.Context) (context.Context, trace.Span) {
	pc, file, linkNo, ok := runtime.Caller(1)
	if !ok {
		log.Error("start span error")
		newCtx, span := ar_trace.Tracer.Start(ctx, "unknow", trace.WithSpanKind(trace.SpanKindProducer))
		return newCtx, span
	}

	funcPaths := strings.Split(runtime.FuncForPC(pc).Name(), "/")
	spanName := funcPaths[len(funcPaths)-1]
	newCtx, span := ar_trace.Tracer.Start(ctx, spanName, trace.WithSpanKind(trace.SpanKindProducer))
	span.SetAttributes(attribute.String(FUNC_PATH, fmt.Sprintf("%s:%v", file, linkNo)))

	return newCtx, span
}

// EndSpan 关闭span
func EndSpan(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	TelemetrySpanEnd(span, err)
}

func End(ctx context.Context) {
	trace.SpanFromContext(ctx).End()
}

// TelemetrySpanEnd 关闭span
func TelemetrySpanEnd(span trace.Span, err error) {
	if span == nil {
		return
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "OK")
	}
	span.End()
}

func SetAttributes(ctx context.Context, kv ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(kv...)
}

func Span(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

func NewOtelHttpClient() *http.Client {
	return &http.Client{
		Transport: otelhttp.NewTransport(
			&http.Transport{
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
				MaxIdleConnsPerHost:   100,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		),
		Timeout: 180 * time.Second,
	}
}
func NewOTELHttpClient20() *http.Client {
	return &http.Client{
		Transport: otelhttp.NewTransport(
			&http.Transport{
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
				MaxIdleConnsPerHost:   100,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		),
		Timeout: 20 * time.Second,
	}
}
func NewOTELHttpClientWithTimeout(timeout time.Duration) *http.Client {
	return NewOTELHttpClientParam(timeout, &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		MaxIdleConnsPerHost:   100,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	})
}
func NewOTELHttpClientParam(timeout time.Duration, transport *http.Transport) *http.Client {
	return &http.Client{
		Transport: otelhttp.NewTransport(
			transport,
		),
		Timeout: timeout,
	}
}

func ReleaseFunc(tracerProvider *sdktrace.TracerProvider) func() {
	if tracerProvider == nil {
		return func() {}
	}
	return func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			panic(err)
		}
	}
}

// InitTracer 初始化trace
func InitTracer(tc *telemetry.Config, instance string) *sdktrace.TracerProvider {
	if tc.TraceEnabled {
		public.SetServiceInfo(tc.ServerName, tc.ServerVersion, "")

		traceClient := public.NewHTTPClient(public.WithAnyRobotURL(tc.TraceUrl),
			public.WithCompression(1), public.WithTimeout(1*time.Second),
			public.WithRetry(true, 5*time.Second, 30*time.Second, 1*time.Minute))
		traceExporter := ar_trace.NewExporter(traceClient)
		tracerProvider := sdktrace.NewTracerProvider(sdktrace.WithBatcher(traceExporter,
			sdktrace.WithMaxQueueSize(50000), sdktrace.WithBlocking(),
			sdktrace.WithMaxExportBatchSize(500),
			sdktrace.WithExportTimeout(time.Hour)), sdktrace.WithResource(ar_trace.TraceResource()))
		otel.SetTracerProvider(tracerProvider)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
		return tracerProvider
	} else {
		return nil
	}
}

// MiddlewareTrace 接口层链路追踪，设置span
func MiddlewareTrace() gin.HandlerFunc {

	return func(c *gin.Context) {

		newCtx, span := StartServerSpan(c)
		defer span.End()
		req := c.Request.WithContext(newCtx)
		c.Request = req

		c.Next()

		status := c.Writer.Status()
		if status/100 >= 4 {
			span.SetStatus(codes.Error, "REQUEST FAILED")
		} else {
			span.SetStatus(codes.Ok, "OK")
		}
		if status > 0 {
			span.SetAttributes(semconv.HTTPStatusCode(status))
		}
	}
}
