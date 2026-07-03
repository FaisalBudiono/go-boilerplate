package otel

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"FaisalBudiono/go-boilerplate/internal/app/core/util/app"

	"go.opentelemetry.io/contrib/processors/minsev"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"gopkg.in/natefinch/lumberjack.v2"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(
	ctx context.Context,
	opts ...option,
) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

	cfgs := &cfg{}
	for _, opt := range opts {
		opt(cfgs)
	}

	// shutdown calls cleanup functions registered via shutdownFuncs.
	// The errors from the calls are joined.
	// Each registered cleanup will be invoked once.
	shutdown = func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}
		shutdownFuncs = nil
		return err
	}

	// handleErr calls shutdown for cleanup and makes sure that all errors are returned.
	handleErr := func(inErr error) {
		err = errors.Join(inErr, shutdown(ctx))
	}

	err = os.MkdirAll(defaultLocalDir, 0o755)
	if err != nil {
		handleErr(err)
		return
	}

	traceLogger, err := logger(
		filepath.Join(defaultLocalDir, "trace.log"), false,
	)
	if err != nil {
		handleErr(err)
		return
	}

	logLogger, err := logger(
		filepath.Join(defaultLocalDir, "log.log"), cfgs.loggingInStdout,
	)
	if err != nil {
		handleErr(err)
		return
	}

	conf, err := newConfig(ctx, traceLogger, logLogger)
	if err != nil {
		handleErr(err)
		return
	}

	// Set up propagator.
	prop := conf.newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := conf.newTraceProvider()
	if err != nil {
		handleErr(err)
		return
	}

	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)

	// Set up logger provider.
	loggerProvider, err := conf.newLoggerProvider()
	if err != nil {
		handleErr(err)
		return
	}
	shutdownFuncs = append(shutdownFuncs, loggerProvider.Shutdown)
	global.SetLoggerProvider(loggerProvider)

	return
}

type config struct {
	res *resource.Resource

	ctx         context.Context
	traceLogger io.Writer
	logLogger   io.Writer
}

func newConfig(
	ctx context.Context,
	traceLogger io.Writer,
	logLogger io.Writer,
) (*config, error) {
	res, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(app.ENV().AppName),
			semconv.ServiceVersion(app.Version()),
		))
	if err != nil {
		return nil, err
	}

	return &config{
		res: res,
		ctx: ctx,

		traceLogger: traceLogger,
		logLogger:   logLogger,
	}, nil
}

func (c *config) newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

func (c *config) newTraceProvider() (*trace.TracerProvider, error) {
	traceExporter, err := c.newTraceExporter()
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithResource(c.res),
		trace.WithBatcher(traceExporter),
	)

	return traceProvider, nil
}

func (c *config) newTraceExporter() (trace.SpanExporter, error) {
	endpoint := app.ENV().Otel.TraceURL

	if endpoint == "" {
		return stdouttrace.New(
			stdouttrace.WithWriter(c.traceLogger),
		)
	}

	return otlptracehttp.New(
		c.ctx,
		otlptracehttp.WithEndpointURL(endpoint),
	)
}

func (c *config) newLoggerProvider() (*log.LoggerProvider, error) {
	logExporter, err := c.newLogExporter()
	if err != nil {
		return nil, err
	}

	minLog := minimumLogLevel()

	options := []log.LoggerProviderOption{log.WithResource(c.res)}

	for _, ex := range logExporter {
		options = append(
			options,
			log.WithProcessor(minsev.NewLogProcessor(
				log.NewBatchProcessor(ex),
				minLog,
			)),
		)
	}

	loggerProvider := log.NewLoggerProvider(options...)

	return loggerProvider, nil
}

func (c *config) newLogExporter() ([]log.Exporter, error) {
	endpoint := app.ENV().Otel.LogURL

	exStdout, err := stdoutlog.New(
		stdoutlog.WithWriter(c.logLogger),
	)
	if err != nil {
		return nil, err
	}

	if endpoint == "" {
		return []log.Exporter{exStdout}, nil
	}

	exHTTP, err := otlploghttp.New(
		c.ctx,
		otlploghttp.WithEndpointURL(endpoint),
	)
	if err != nil {
		return nil, err
	}

	return []log.Exporter{exStdout, exHTTP}, nil
}

func logger(filename string, withStdout bool) (io.Writer, error) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}

	fileLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    100, // megabytes
		MaxBackups: 30,
		MaxAge:     30, // days
	}

	writters := []io.Writer{fileLogger}

	if withStdout {
		writters = append(writters, os.Stdout)
	}

	return io.MultiWriter(writters...), nil
}

const defaultLocalDir = "./logs"

func minimumLogLevel() minsev.Severity {
	switch app.ENV().Log.Level {
	case app.LogLevelDebug:
		return minsev.SeverityDebug
	case app.LogLevelInfo:
		return minsev.SeverityInfo
	case app.LogLevelWarn:
		return minsev.SeverityWarn
	case app.LogLevelError:
		return minsev.SeverityError
	default:
		return minsev.SeverityInfo
	}
}

type cfg struct {
	loggingInStdout bool
}

type option func(*cfg)

func WithStdoutLogging() option {
	return func(cfg *cfg) {
		cfg.loggingInStdout = true
	}
}
