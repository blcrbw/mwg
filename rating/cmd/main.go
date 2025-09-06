package main

import (
	"context"
	"fmt"
	"log"
	"mmoviecom/gen"
	"mmoviecom/internal/grpcutil"
	"mmoviecom/pkg/discovery"
	"mmoviecom/pkg/discovery/consul"
	"mmoviecom/rating/configs"
	"mmoviecom/rating/internal/controller/rating"
	authgateway "mmoviecom/rating/internal/gateway/auth/grpc"
	grpchandler "mmoviecom/rating/internal/handler/grpc"
	"mmoviecom/rating/internal/ingester/kafka"
	"mmoviecom/rating/internal/repository/mysql"
	"net"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
)

const serviceName = "rating"

func main() {
	log.Printf("Starting the rating service")

	f, err := os.Open("defaults.yaml")
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
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("rating:%d", cfg.API.Port)); err != nil {
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

	repo, err := mysql.New(cfg.DatabaseConfig.Mysql)
	if err != nil {
		panic(err)
	}
	ingester, err := kafka.NewIngester(cfg.MessengerConfig.Kafka.Address, "rating", "ratings")
	if err != nil {
		log.Fatalf("Failed to initialize ingester: %v", err)
	}

	creds := grpcutil.GetX509Credentials("cert.crt", "cert.key")
	auth := authgateway.New(registry, creds)
	svc := rating.New(repo, ingester, auth)
	go func() {
		if err := svc.StartIngestion(ctx); err != nil {
			log.Fatalf("Failed to start ingestion: %v", err)
		}
	}()
	h := grpchandler.New(svc)

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", cfg.API.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer(grpc.Creds(creds))
	gen.RegisterRatingServiceServer(srv, h)
	reflection.Register(srv)
	log.Printf("Register reflectrion")
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
