package otel

import (
	"FaisalBudiono/go-boilerplate/internal/app/core/util/app"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"go.opentelemetry.io/contrib/processors/minsev"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.34.0"
	"gopkg.in/natefinch/lumberjack.v2"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDK(ctx context.Context) (shutdown func(context.Context) error, err error) {
	var shutdownFuncs []func(context.Context) error

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

	dir := "./logs"
	err = os.MkdirAll(dir, 0755)
	if err != nil {
		handleErr(err)
		return
	}

	traceLogger, err := logger(filepath.Join(dir, "trace.log"))
	if err != nil {
		handleErr(err)
		return
	}

	logLogger, err := logger(filepath.Join(dir, "log.log"))
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
		resource.NewWithAttributes(semconv.SchemaURL,
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

// @todo add configurable env for trace and log exporter so it can sent to turn on the opentelemetry one by one
func (c *config) newTraceExporter() (trace.SpanExporter, error) {
	endpoint := app.ENV().OtelEndpoint

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

func (c *config) newMeterProvider() (*metric.MeterProvider, error) {
	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(c.res),
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(time.Minute))),
	)

	return meterProvider, nil
}

func (c *config) newMetricExporter() (metric.Exporter, error) {
	endpoint := app.ENV().OtelEndpoint

	if endpoint == "" {
		return stdoutmetric.New()
	}

	return otlpmetrichttp.New(
		c.ctx,
		otlpmetrichttp.WithEndpointURL(endpoint),
	)
}

func (c *config) newLoggerProvider() (*log.LoggerProvider, error) {
	logExporter, err := c.newLogExporter()
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithResource(c.res),
		log.WithProcessor(minsev.NewLogProcessor(
			log.NewBatchProcessor(logExporter),
			minimumLogLevel(),
		)),
	)

	return loggerProvider, nil
}

func (c *config) newLogExporter() (log.Exporter, error) {
	return stdoutlog.New(
		stdoutlog.WithWriter(c.logLogger),
	)
}

func logger(filename string) (io.Writer, error) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	file.Close()

	fileLogger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    100, // megabytes
		MaxBackups: 30,
		MaxAge:     30, // days
	}

	return io.MultiWriter(
		// os.Stdout, // uncomment to log to stdout
		fileLogger,
	), nil
}

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
