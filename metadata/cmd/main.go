package main

import (
	"log"
	"mmoviecom/metadata/internal/controller/metadata"
	httphandler "mmoviecom/metadata/internal/handler/http"
	"mmoviecom/metadata/internal/repository/memory"
	"net/http"
)

func main() {
	log.Println("Starting the movie metadata service")
	repo := memory.New()
	ctrl := metadata.New(repo)
	h := httphandler.New(ctrl)
	http.Handle("/metadata", http.HandlerFunc(h.GetMetadata))
	if err := http.ListenAndServe(":8087", nil); err != nil {
		panic(err)
	}
}
