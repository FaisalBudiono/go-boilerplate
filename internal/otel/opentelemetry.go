package otel

import (
	"context"
	"errors"
	"time"

	"github.com/spf13/viper"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
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
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
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

	traceLogger := &lumberjack.Logger{
		Filename:   "./logs/trace.log",
		MaxSize:    100, // megabytes
		MaxBackups: 30,
		MaxAge:     90, // days
	}

	conf, err := newConfig(ctx, traceLogger)
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
	traceLogger *lumberjack.Logger
}

func newConfig(
	ctx context.Context,
	traceLogger *lumberjack.Logger,
) (*config, error) {
	res, err := resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName(viper.GetString("APP_NAME")),
			semconv.ServiceVersion(viper.GetString("APP_VERSION")),
		))
	if err != nil {
		return nil, err
	}

	return &config{
		res: res,
		ctx: ctx,

		traceLogger: traceLogger,
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
		trace.WithBatcher(traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
	)

	return traceProvider, nil
}

func (c *config) newTraceExporter() (trace.SpanExporter, error) {
	endpoint := viper.GetString("OTLP_ENDPOINT")

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
	endpoint := viper.GetString("OTLP_ENDPOINT")

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
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
	)

	return loggerProvider, nil
}

func (c *config) newLogExporter() (log.Exporter, error) {
	endpoint := viper.GetString("OTLP_ENDPOINT")

	if endpoint == "" {
		return stdoutlog.New(
			stdoutlog.WithWriter(c.traceLogger),
		)
	}

	return otlploghttp.New(
		c.ctx,
		otlploghttp.WithEndpointURL(endpoint),
	)
}
