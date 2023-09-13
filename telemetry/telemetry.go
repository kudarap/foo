package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	neturl "net/url"
	"os"
	"strings"
	"time"

	otellogs "github.com/dbdoyc/opentelemetry-logs-go"
	"github.com/dbdoyc/opentelemetry-logs-go/exporters/otlp/otlplogs"
	"github.com/dbdoyc/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogsgrpc"
	"github.com/dbdoyc/opentelemetry-logs-go/exporters/otlp/otlplogs/otlplogshttp"
	"github.com/dbdoyc/opentelemetry-logs-go/exporters/stdout/stdoutlogs"
	sdklogs "github.com/dbdoyc/opentelemetry-logs-go/sdk/logs"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.20.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// InitProvider Initializes an OTLP exporter, and configures the corresponding trace and metric providers.
func InitProvider(conf Config, suffix, versionTag string) (func(context.Context) error, error) {
	if !conf.Enabled {
		// return no-op shutdown to prevent nil pointer.
		return func(context.Context) error { return nil }, nil
	}

	ctx := context.Background()
	host, _ := os.Hostname()
	version := strings.TrimLeft(versionTag, "v")
	res, err := sdkresource.New(ctx,
		sdkresource.WithSchemaURL(semconv.SchemaURL),
		sdkresource.WithAttributes(
			semconv.ServiceName(fmt.Sprintf("%s-%s", conf.ServiceName, suffix)),
			semconv.ServiceVersion(version),
			semconv.HostName(host),
			semconv.TelemetrySDKLanguageGo,
			semconv.DeploymentEnvironment(conf.Env),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// When collector URL has value it will automatically send data to collector.
	var traceExporter sdktrace.SpanExporter
	var logExporter sdklogs.LogRecordExporter
	if strings.TrimSpace(conf.CollectorURL) != "" {
		url, err := neturl.Parse(conf.CollectorURL)
		if err != nil {
			return nil, err
		}

		switch url.Scheme {
		case "grpc":
			traceExporter, err = grpcTraceExporter(url.Host)
			if err != nil {
				return nil, err
			}
			logExporter, err = grpcLogExporter(url.Host)
			if err != nil {
				return nil, err
			}
		case "http", "https":
			traceExporter, err = httpTraceExporter(url.Host)
			if err != nil {
				return nil, err
			}
			logExporter, err = httpLogExporter(url.Host)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("un-supported protocol scheme: %s", url.Scheme)
		}
	} else {
		traceExporter, err = stdoutTraceExporter()
		if err != nil {
			return nil, err
		}
		logExporter, err = stdoutLogExporter()
		if err != nil {
			return nil, err
		}
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	lrp := sdklogs.NewBatchLogRecordProcessor(logExporter)
	loggerProvider := sdklogs.NewLoggerProvider(
		sdklogs.WithLogRecordProcessor(lrp),
		sdklogs.WithResource(res),
	)
	otellogs.SetLoggerProvider(loggerProvider)

	// Set global propagator to TraceContext and Baggage.
	// This allows us to trace external http or grpc calls that has open-telemetry integration.
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Shutdown will flush any remaining spans and shut down the exporter.
	return shutdownProviders(tracerProvider.Shutdown, loggerProvider.Shutdown), nil
}

type Config struct {
	Enabled      bool
	CollectorURL string
	ServiceName  string
	Env          string
}

type providerCloser func(ctx context.Context) error

// shutdownProviders closes all providers.
func shutdownProviders(pc ...providerCloser) providerCloser {
	return func(ctx context.Context) error {
		for _, closer := range pc {
			if err := closer(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

func stdoutTraceExporter() (sdktrace.SpanExporter, error) {
	return stdouttrace.New(
		stdouttrace.WithWriter(os.Stdout),
		stdouttrace.WithPrettyPrint(),
		stdouttrace.WithoutTimestamps(),
	)
}

func httpTraceExporter(url string) (sdktrace.SpanExporter, error) {
	ctx := context.Background()
	exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint(url), otlptracehttp.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}
	return exporter, nil
}

func grpcTraceExporter(url string) (sdktrace.SpanExporter, error) {
	ctx := context.Background()

	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
	// endpoint of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, url,
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}
	return traceExporter, nil
}

func stdoutLogExporter() (sdklogs.LogRecordExporter, error) {
	return stdoutlogs.NewExporter(stdoutlogs.WithPrettyPrint())
}

func httpLogExporter(url string) (sdklogs.LogRecordExporter, error) {
	opts := []otlplogshttp.Option{otlplogshttp.WithEndpoint(url)}
	if !strings.HasPrefix(url, "https://") {
		opts = append(opts, otlplogshttp.WithInsecure())
	}

	ctx := context.Background()
	c := otlplogshttp.NewClient(opts...)
	logExporter, err := otlplogs.NewExporter(ctx, otlplogs.WithClient(c))
	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}
	return logExporter, nil
}

func grpcLogExporter(url string) (sdklogs.LogRecordExporter, error) {
	ctx := context.Background()

	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` endpoint. Otherwise, replace `localhost` with the
	// endpoint of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns.
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, url,
		// Note the use of insecure transport here. TLS is recommended in production.
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	c := otlplogsgrpc.NewClient(otlplogsgrpc.WithGRPCConn(conn))
	logExporter, err := otlplogs.NewExporter(ctx, otlplogs.WithClient(c))
	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}
	return logExporter, nil
}

func jsonAttribute(k string, v interface{}) attribute.KeyValue {
	b, _ := json.Marshal(v)
	return attribute.String(k, string(b))
}
