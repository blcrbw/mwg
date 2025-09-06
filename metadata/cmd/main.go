package main

import (
	"context"
	"fmt"
	"log"
	"mmoviecom/gen"
	"mmoviecom/internal/grpcutil"
	"mmoviecom/metadata/configs"
	"mmoviecom/metadata/internal/controller/metadata"
	grpchandler "mmoviecom/metadata/internal/handler/grpc"
	"mmoviecom/metadata/internal/repository/memory"
	"mmoviecom/metadata/internal/repository/mysql"
	"mmoviecom/pkg/discovery"
	"mmoviecom/pkg/discovery/consul"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "metadata"

func main() {
	log.Printf("Starting the movie metadata service")

	f, err := os.Open("/home/blc/prj/mwg/metadata/configs/defaults.yaml")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	var cfg configs.ServiceConfig
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}
	registry, err := consul.NewRegistry(cfg.ServiceDiscovery.Consul.Address)
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("metadata:%d", cfg.API.Port)); err != nil {
		panic(err)
	}
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Printf("Failed to report healthy state: %v", err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)

	repo, err := mysql.New(cfg.DatabaseConfig.Mysql)
	if err != nil {
		panic(err)
	}
	cache := memory.New()
	svc := metadata.New(repo, cache)
	h := grpchandler.New(svc)

	creds := grpcutil.GetX509Credentials("cert.crt", "cert.key")
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.API.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer(grpc.Creds(creds))
	gen.RegisterMetadataServiceServer(srv, h)
	reflection.Register(srv)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
