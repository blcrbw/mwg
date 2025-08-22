package main

import (
	"log"
	"mmoviecom/movie/internal/controller/movie"
	metadatagateway "mmoviecom/movie/internal/gateway/metadata/http"
	ratinggateway "mmoviecom/movie/internal/gateway/rating/http"
	moviehandler "mmoviecom/movie/internal/handler/http"
	"net/http"
)

func main() {
	log.Println("Starting the movie service")
	metadataGateway := metadatagateway.New("localhost:8087")
	ratingGateway := ratinggateway.New("localhost:8089")
	ctrl := movie.New(ratingGateway, metadataGateway)
	h := moviehandler.New(ctrl)
	http.Handle("/movie", http.HandlerFunc(h.GetMovieDetails))
	if err := http.ListenAndServe(":8086", nil); err != nil {
		panic(err)
	}
}
