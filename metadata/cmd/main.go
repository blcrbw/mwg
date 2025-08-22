package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"mmoviecom/metadata/internal/controller/metadata"
	httphandler "mmoviecom/metadata/internal/handler/http"
	"mmoviecom/metadata/internal/repository/memory"
	"mmoviecom/pkg/discovery"
	"mmoviecom/pkg/discovery/consul"
	"net/http"
	"time"
)

const serviceName = "metadata"

func main() {
	var port int
	flag.IntVar(&port, "port", 8081, "API handler port")
	flag.Parse()

	log.Printf("Starting the movie metadata service on port  %d", port)
	registry, err := consul.NewRegistry("localhost:8500")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
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

	repo := memory.New()
	svc := metadata.New(repo)
	h := httphandler.New(svc)
	http.Handle("/metadata", http.HandlerFunc(h.GetMetadata))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		panic(err)
	}
}
