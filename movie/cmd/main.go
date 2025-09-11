package main

import (
	"context"
	"fmt"
	"mmoviecom/gen"
	"mmoviecom/internal/grpcutil"
	"mmoviecom/movie/configs"
	"mmoviecom/movie/internal/controller/movie"
	metadatagateway "mmoviecom/movie/internal/gateway/metadata/grpc"
	ratinggateway "mmoviecom/movie/internal/gateway/rating/grpc"
	moviegrpchandler "mmoviecom/movie/internal/handler/grpc"
	"mmoviecom/pkg/discovery"
	"mmoviecom/pkg/discovery/consul"
	"mmoviecom/pkg/limiter"
	"mmoviecom/pkg/logging"
	"mmoviecom/pkg/metrics"
	"mmoviecom/pkg/tracing"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/ratelimit"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "movie"

func main() {
	logConfig := zap.NewProductionConfig()
	log, err := logConfig.Build()
	if err != nil {
		panic(err)
	}
	log = log.With(zap.String(logging.FieldService, serviceName))

	f, err := os.Open("defaults.yaml")
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Warn("failed to close file", zap.Error(err))
		}
	}()
	var cfg configs.ServiceConfig
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}

	log.Info("Starting the service", zap.Int(logging.FieldPort, cfg.API.Port))

	ctx, cancel := context.WithCancel(context.Background())

	tp, err := tracing.NewJaegerProvider(cfg.Jaeger.URL, serviceName)
	if err != nil {
		log.Fatal("Failed to initialize jaeger provider", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			log.Fatal("Failed to shutdown jaeger provider", zap.Error(err))
		}
	}()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Address, log)
	if err != nil {
		panic(err)
	}
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("movie:%d", cfg.API.Port)); err != nil {
		panic(err)
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
				if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
					log.Warn("Failed to report healthy state", zap.Error(err))
				}
			}
		}
	}()
	defer func() {
		if err := registry.Deregister(ctx, instanceID, serviceName); err != nil {
			log.Warn("Failed to deregister service", zap.Error(err))
		}
	}()

	creds := grpcutil.GetX509Credentials("cert.crt", "cert.key")
	metadataGateway := metadatagateway.New(registry, creds, log)
	ratingGateway := ratinggateway.New(registry, creds, log)
	svc := movie.New(ratingGateway, metadataGateway, log)
	h := moviegrpchandler.New(svc, log)

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.API.Port))
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	l := limiter.New(log, 100, 50)
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(ratelimit.UnaryServerInterceptor(l)),
		grpc.Creds(creds),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	gen.RegisterMovieServiceServer(srv, h)
	log.Info("Register reflection")
	reflection.Register(srv)

	_, closer := metrics.NewMetricsReporter(log, serviceName, cfg.Prometheus.MetricsPort)
	defer func() {
		if err := closer.Close(); err != nil {
			log.Warn("Failed to close Prometheus reporter scope", zap.Error(err))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		s := <-sigChan
		cancel()
		log.Info("Got signal, attempting graceful shutdown", zap.Stringer(logging.FieldSignal, s))
		srv.GracefulStop()
		log.Info("Gracefully stopped the gRPC server")
	}()

	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
	wg.Wait()
}
