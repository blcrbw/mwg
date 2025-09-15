package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mmoviecom/gen"
	"mmoviecom/internal/grpcutil"
	"os"
	"path"

	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.NewClient("localhost:8083", grpc.WithTransportCredentials(grpcutil.GetX509Credentials("cert.crt", "cert.key")))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	client := gen.NewMovieServiceClient(conn)
	filePath := "upload.txt"
	if err := uploadFile(client, filePath); err != nil {
		log.Fatalf("Failed to upload file: %v", err)
	}
}

func uploadFile(client gen.MovieServiceClient, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	base := path.Base(filePath)

	stream, err := client.UploadFile(context.Background())
	if err != nil {
		return fmt.Errorf("failed to upload stream: %w", err)
	}
	buf := make([]byte, 1024)
	for {
		n, err := file.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return fmt.Errorf("failed to read file: %w", err)
			}
		}

		if err := stream.Send(&gen.UploadRequest{
			Filename: base,
			Chunk:    buf[:n],
		}); err != nil {
			return fmt.Errorf("failed to send chunk: %w", err)
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("failed to receive response: %w", err)
	}

	fmt.Println("Server response: ", resp.GetMessage())
	return nil
}
