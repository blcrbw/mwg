package main

import (
	"context"
	"fmt"
	"log"
	"mmoviecom/auth/configs"
	grpchandler "mmoviecom/auth/internal/handler/grpc"
	"mmoviecom/gen"
	"mmoviecom/internal/grpcutil"
	"mmoviecom/pkg/discovery"
	"mmoviecom/pkg/discovery/consul"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "auth"

func main() {
	f, err := os.Open("defaults.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var cfg configs.ServiceConfig
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}

	log.Printf("Starting the auth service on port %d", cfg.API.Port)
	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Address)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("auth:%d", cfg.API.Port)); err != nil {
		panic(err)
	}
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state: " + err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.API.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	h := grpchandler.New(func() []byte {
		return []byte("test-secret")
	})

	creds := grpcutil.GetX509Credentials("cert.crt", "cert.key")
	srv := grpc.NewServer(grpc.Creds(creds))
	reflection.Register(srv)
	gen.RegisterAuthServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
