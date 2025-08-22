package main

import (
	"log"
	"mmoviecom/rating/internal/controller/rating"
	httphandler "mmoviecom/rating/internal/handler/http"
	"mmoviecom/rating/internal/repository/memory"
	"net/http"
)

func main() {
	log.Printf("Starting the rating service")
	repo := memory.New()
	ctrl := rating.New(repo)
	h := httphandler.New(ctrl)
	http.Handle("/rating", http.HandlerFunc(h.Handle))
	if err := http.ListenAndServe(":8089", nil); err != nil {
		panic(err)
	}
}
