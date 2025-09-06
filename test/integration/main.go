package main

import (
	"context"
	"log"
	"mmoviecom/gen"
	metadatatest "mmoviecom/metadata/pkg/testutil"
	movietest "mmoviecom/movie/pkg/testutil"
	"mmoviecom/pkg/discovery"
	"mmoviecom/pkg/discovery/memory"
	ratingtest "mmoviecom/rating/pkg/testutil"
	"net"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	metadataServiceName = "metadata"
	ratingServiceName   = "rating"
	movieServiceName    = "movie"

	metadataServiceAddress = "localhost:8081"
	ratingServiceAddress   = "localhost:8082"
	movieServiceAddress    = "localhost:8083"
)

func main() {
	log.Println("Starting the integration test")

	ctx := context.Background()
	registry := memory.NewRegistry()

	log.Println("Setting up service handlers and clients")

	metadataSrv := startMetadataService(ctx, registry)
	defer metadataSrv.GracefulStop()
	ratingSrv := startRatingService(ctx, registry)
	defer ratingSrv.GracefulStop()
	movieSrv := startMovieService(ctx, registry)
	defer movieSrv.GracefulStop()

	opts := grpc.WithTransportCredentials(insecure.NewCredentials())
	metadataConn, err := grpc.NewClient(metadataServiceAddress, opts)
	if err != nil {
		panic(err)
	}
	defer metadataConn.Close()
	metadataClient := gen.NewMetadataServiceClient(metadataConn)

	ratingConn, err := grpc.NewClient(ratingServiceAddress, opts)
	if err != nil {
		panic(err)
	}
	defer ratingConn.Close()
	ratingClient := gen.NewRatingServiceClient(ratingConn)

	movieConn, err := grpc.NewClient(movieServiceAddress, opts)
	if err != nil {
		panic(err)
	}
	defer movieConn.Close()
	movieClient := gen.NewMovieServiceClient(movieConn)

	log.Println("Saving test metadata via metadata service")
	m := &gen.Metadata{
		Id:          "the-movie",
		Title:       "The Movie",
		Description: "The Movie, the one and only",
		Director:    "Mr. D",
	}

	if _, err := metadataClient.PutMetadata(ctx, &gen.PutMetadataRequest{Metadata: m}); err != nil {
		log.Fatalf("put metadata: %v", err)
	}

	log.Println("Retrieving test metadata via metadata service")

	getMetadataResp, err := metadataClient.GetMetadata(ctx, &gen.GetMetadataRequest{MovieId: m.Id})
	if err != nil {
		log.Fatalf("get metadata: %v", err)
	}
	if diff := cmp.Diff(getMetadataResp.Metadata, m, cmpopts.IgnoreUnexported(gen.Metadata{})); diff != "" {
		log.Fatalf("get metadata after put mismatch: %v", diff)
	}

	log.Println("Getting movie details via movie service")
	wantMovieDetails := &gen.MovieDetails{
		Metadata: m,
	}

	getMovieDetailsResp, err := movieClient.GetMovieDetails(ctx, &gen.GetMovieDetailsRequest{MovieId: m.Id})
	if err != nil {
		log.Fatalf("get movie details: %v", err)
	}
	if diff := cmp.Diff(getMovieDetailsResp.MovieDetails, wantMovieDetails, cmpopts.IgnoreUnexported(gen.MovieDetails{}, gen.Metadata{})); diff != "" {
		log.Fatalf("get movie details after put mismatch: %v", diff)
	}

	// @TODO: get user token.
	token := "1"

	log.Println("Saving first rating via rating service")
	const userID = "user0"
	const recordTypeMovie = "movie"
	firstRating := int32(5)
	if _, err = ratingClient.PutRating(ctx, &gen.PutRatingRequest{
		UserId:      userID,
		RecordId:    m.Id,
		RecordType:  recordTypeMovie,
		RatingValue: firstRating,
		Token:       token,
	}); err != nil {
		log.Fatalf("put rating: %v", err)
	}

	log.Println("Retrieving initial aggregated rating via rating service")
	getAggregatedRatingResp, err := ratingClient.GetAggregatedRating(ctx, &gen.GetAggregatedRatingRequest{
		RecordType: recordTypeMovie,
		RecordId:   m.Id,
	})
	if err != nil {
		log.Fatalf("get aggregated rating: %v", err)
	}

	if got, want := getAggregatedRatingResp.RatingValue, float64(5); got != want {
		log.Fatalf("rating mismatch: got %v, want %v", got, want)
	}

	log.Println("Saving second rating via rating service")
	secondRating := int32(1)
	if _, err = ratingClient.PutRating(ctx, &gen.PutRatingRequest{
		UserId:      userID,
		RecordId:    m.Id,
		RecordType:  recordTypeMovie,
		RatingValue: secondRating,
		Token:       token,
	}); err != nil {
		log.Fatalf("put rating: %v", err)
	}

	log.Println("Saving new aggregated rating via rating service")

	getAggregatedRatingResp, err = ratingClient.GetAggregatedRating(ctx, &gen.GetAggregatedRatingRequest{
		RecordType: recordTypeMovie,
		RecordId:   m.Id,
	})
	if err != nil {
		log.Fatalf("get aggregated rating: %v", err)
	}

	wantRating := float64((firstRating + secondRating) / 2)
	if got, want := getAggregatedRatingResp.RatingValue, wantRating; got != want {
		log.Fatalf("rating mismatch: got %v, want %v", got, want)
	}

	log.Println("Getting updated movie details via movie service")

	getMovieDetailsResp, err = movieClient.GetMovieDetails(ctx, &gen.GetMovieDetailsRequest{
		MovieId: m.Id,
	})
	if err != nil {
		log.Fatalf("get movie details: %v", err)
	}
	wantMovieDetails.Rating = wantRating
	if diff := cmp.Diff(getMovieDetailsResp.MovieDetails, wantMovieDetails, cmpopts.IgnoreUnexported(gen.MovieDetails{}, gen.Metadata{})); diff != "" {
		log.Fatalf("get movie details after update mismatch: %v", diff)
	}

	log.Println("Integration test execution successful")
}

func startMetadataService(ctx context.Context, registry discovery.Registry) *grpc.Server {
	log.Println("Starting metadata service on " + metadataServiceAddress)
	h := metadatatest.NewTestMetadataGRPCServer()
	l, err := net.Listen("tcp", metadataServiceAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	gen.RegisterMetadataServiceServer(srv, h)
	id := discovery.GenerateInstanceID(metadataServiceName)
	if err := registry.Register(ctx, id, metadataServiceName, metadataServiceAddress); err != nil {
		panic(err)
	}
	go func() {
		defer func() {
			if err := registry.Deregister(ctx, id, metadataServiceName); err != nil {
				log.Printf("Failed to deregister %s: %v", metadataServiceName, err)
			}
		}()
		if err := srv.Serve(l); err != nil {
			panic(err)
		}
	}()
	go func() {
		for {
			if err := registry.ReportHealthyState(id, metadataServiceName); err != nil {
				log.Printf("Failed to report healthy state: %v", err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()
	return srv
}

func startRatingService(ctx context.Context, registry discovery.Registry) *grpc.Server {
	log.Println("Starting rating service on " + ratingServiceAddress)
	h := ratingtest.NewTestRatingGRPCServer()
	l, err := net.Listen("tcp", ratingServiceAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	gen.RegisterRatingServiceServer(srv, h)
	id := discovery.GenerateInstanceID(ratingServiceName)
	if err := registry.Register(ctx, id, ratingServiceName, ratingServiceAddress); err != nil {
		panic(err)
	}
	go func() {
		defer func() {
			if err := registry.Deregister(ctx, id, ratingServiceName); err != nil {
				log.Printf("Failed to deregister %s: %v", ratingServiceName, err)
			}
		}()
		if err := srv.Serve(l); err != nil {
			panic(err)
		}
	}()
	go func() {
		for {
			if err := registry.ReportHealthyState(id, ratingServiceName); err != nil {
				log.Printf("Failed to report healthy state: %v", err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()
	return srv
}

func startMovieService(ctx context.Context, registry discovery.Registry) *grpc.Server {
	log.Println("Starting movie service on " + movieServiceAddress)
	h := movietest.NewTestMovieGRPCServer(registry)
	l, err := net.Listen("tcp", movieServiceAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	gen.RegisterMovieServiceServer(srv, h)
	id := discovery.GenerateInstanceID(movieServiceName)
	if err := registry.Register(ctx, id, movieServiceName, movieServiceAddress); err != nil {
		panic(err)
	}
	go func() {
		defer func() {
			if err := registry.Deregister(ctx, id, movieServiceName); err != nil {
				log.Printf("Failed to deregister %s: %v", movieServiceName, err)
			}
		}()
		if err := srv.Serve(l); err != nil {
			panic(err)
		}
	}()
	go func() {
		for {
			if err := registry.ReportHealthyState(id, movieServiceName); err != nil {
				log.Printf("Failed to report healthy state: %v", err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()
	return srv
}
