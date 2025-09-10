package main

import (
	"context"
	"fmt"
	"mmoviecom/auth/configs"
	grpchandler "mmoviecom/auth/internal/handler/grpc"
	"mmoviecom/gen"
	"mmoviecom/internal/grpcutil"
	"mmoviecom/pkg/discovery"
	"mmoviecom/pkg/discovery/consul"
	"mmoviecom/pkg/logging"
	"mmoviecom/pkg/tracing"

	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "auth"

func main() {
	log, err := zap.NewProduction()
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
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("auth:%d", cfg.API.Port)); err != nil {
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
		err := registry.Deregister(ctx, instanceID, serviceName)
		if err != nil {
			log.Warn("Failed to deregister service", zap.Error(err))
		}
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.API.Port))
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}
	h := grpchandler.New(func() []byte {
		return []byte("test-secret")
	})

	creds := grpcutil.GetX509Credentials("cert.crt", "cert.key")
	srv := grpc.NewServer(
		grpc.Creds(creds),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)
	reflection.Register(srv)
	gen.RegisterAuthServiceServer(srv, h)

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
