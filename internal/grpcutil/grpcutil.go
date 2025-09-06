package grpcutil

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"math/rand"
	"mmoviecom/pkg/discovery"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ServiceConnection attempts to select a random service instance
// and returns a gRPC connection to it.
func ServiceConnection(ctx context.Context, serviceName string, registry discovery.Registry, creds credentials.TransportCredentials) (*grpc.ClientConn, error) {
	addrs, err := registry.ServiceAddresses(ctx, serviceName)
	if err != nil {
		return nil, err
	}
	return grpc.NewClient(addrs[rand.Intn(len(addrs))], grpc.WithTransportCredentials(creds))
}

// GetX509Credentials reads cert and key files and prepares TLS credentials.
func GetX509Credentials(c string, k string) credentials.TransportCredentials {
	certBytes, err := os.ReadFile(c)
	if err != nil {
		log.Fatalf("Failed to read certificate: %v", err)
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(certBytes) {
		log.Fatalf("Failed to append certificate")
	}
	cert, err := tls.LoadX509KeyPair(c, k)
	if err != nil {
		log.Fatalf("Failed to load key pair: %v", err)
	}
	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	})
	return creds
}
