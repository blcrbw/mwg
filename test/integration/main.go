package main

import (
	"context"
	authtest "mmoviecom/auth/pkg/testutil"
	"mmoviecom/gen"
	metadatatest "mmoviecom/metadata/pkg/testutil"
	movietest "mmoviecom/movie/pkg/testutil"
	"mmoviecom/pkg/discovery"
	"mmoviecom/pkg/discovery/memory"
	"mmoviecom/pkg/metrics"
	ratingtest "mmoviecom/rating/pkg/testutil"
	"net"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/uber-go/tally/v6"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	metadataServiceName = "metadata"
	ratingServiceName   = "rating"
	movieServiceName    = "movie"
	authServiceName     = "auth"

	metadataServiceAddress = "localhost:8081"
	ratingServiceAddress   = "localhost:8082"
	movieServiceAddress    = "localhost:8083"
	authServiceAddress     = "localhost:8084"
)

func main() {
	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	log = log.With(zap.String("env", "integration-test"))
	log.Info("Starting the integration test")

	ctx := context.Background()
	registry := memory.NewRegistry(log)
	scope, closer := metrics.NewMetricsReporter(log, "integration_test", 9099)
	defer closer.Close()

	log.Info("Setting up service handlers and clients")

	authSrv := startAuthService(ctx, registry, log, scope)
	defer authSrv.GracefulStop()
	metadataSrv := startMetadataService(ctx, registry, log, scope)
	defer metadataSrv.GracefulStop()
	ratingSrv := startRatingService(ctx, registry, log, scope)
	defer ratingSrv.GracefulStop()
	movieSrv := startMovieService(ctx, registry, log, scope)
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

	authConn, err := grpc.NewClient(authServiceAddress, opts)
	if err != nil {
		panic(err)
	}
	defer authConn.Close()
	authClient := gen.NewAuthServiceClient(authConn)

	log.Info("Saving test metadata via metadata service")
	m := &gen.Metadata{
		Id:          "the-movie",
		Title:       "The Movie",
		Description: "The Movie, the one and only",
		Director:    "Mr. D",
	}

	if _, err := metadataClient.PutMetadata(ctx, &gen.PutMetadataRequest{Metadata: m}); err != nil {
		log.Fatal("put metadata", zap.Error(err))
	}

	log.Info("Retrieving test metadata via metadata service")

	getMetadataResp, err := metadataClient.GetMetadata(ctx, &gen.GetMetadataRequest{MovieId: m.Id})
	if err != nil {
		log.Fatal("get metadata", zap.Error(err))
	}
	if diff := cmp.Diff(getMetadataResp.Metadata, m, cmpopts.IgnoreUnexported(gen.Metadata{})); diff != "" {
		log.Fatal("get metadata after put mismatch", zap.String("diff", diff))
	}

	log.Info("Getting movie details via movie service")
	wantMovieDetails := &gen.MovieDetails{
		Metadata: m,
	}

	getMovieDetailsResp, err := movieClient.GetMovieDetails(ctx, &gen.GetMovieDetailsRequest{MovieId: m.Id})
	if err != nil {
		log.Fatal("get movie details", zap.Error(err))
	}
	if diff := cmp.Diff(getMovieDetailsResp.MovieDetails, wantMovieDetails, cmpopts.IgnoreUnexported(gen.MovieDetails{}, gen.Metadata{})); diff != "" {
		log.Fatal("get movie details after put mismatch", zap.String("diff", diff))
	}

	const userID = "user0"
	log.Info("Getting token via auth service")
	getTokenResp, err := authClient.GetToken(ctx, &gen.GetTokenRequest{
		Username: userID,
		Password: "password",
	})
	if err != nil {
		log.Fatal("get token", zap.Error(err))
	}
	token := getTokenResp.GetToken()
	if token == "" {
		log.Fatal("get token: empty token")
	}

	log.Info("Verifying token via auth service")
	validateTokenResp, err := authClient.ValidateToken(ctx, &gen.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		log.Fatal("validate token", zap.Error(err))
	}
	if validateTokenResp.GetUsername() != userID {
		log.Fatal("validate token: wrong username")
	}

	log.Info("Saving first rating via rating service")
	const recordTypeMovie = "movie"
	firstRating := int32(5)
	if _, err = ratingClient.PutRating(ctx, &gen.PutRatingRequest{
		UserId:      userID,
		RecordId:    m.Id,
		RecordType:  recordTypeMovie,
		RatingValue: firstRating,
		Token:       token,
	}); err != nil {
		log.Fatal("put rating", zap.Error(err))
	}

	log.Info("Retrieving initial aggregated rating via rating service")
	getAggregatedRatingResp, err := ratingClient.GetAggregatedRating(ctx, &gen.GetAggregatedRatingRequest{
		RecordType: recordTypeMovie,
		RecordId:   m.Id,
	})
	if err != nil {
		log.Fatal("get aggregated rating", zap.Error(err))
	}

	if got, want := getAggregatedRatingResp.RatingValue, float64(5); got != want {
		log.Fatal("rating mismatch", zap.Float64("got", got), zap.Float64("want", want))
	}

	log.Info("Saving second rating via rating service")
	secondRating := int32(1)
	if _, err = ratingClient.PutRating(ctx, &gen.PutRatingRequest{
		UserId:      userID,
		RecordId:    m.Id,
		RecordType:  recordTypeMovie,
		RatingValue: secondRating,
		Token:       token,
	}); err != nil {
		log.Fatal("put rating", zap.Error(err))
	}

	log.Info("Saving new aggregated rating via rating service")

	getAggregatedRatingResp, err = ratingClient.GetAggregatedRating(ctx, &gen.GetAggregatedRatingRequest{
		RecordType: recordTypeMovie,
		RecordId:   m.Id,
	})
	if err != nil {
		log.Fatal("get aggregated rating", zap.Error(err))
	}

	wantRating := float64((firstRating + secondRating) / 2)
	if got, want := getAggregatedRatingResp.RatingValue, wantRating; got != want {
		log.Fatal("rating mismatch: got %v, want %v", zap.Float64("got", got), zap.Float64("want", want))
	}

	log.Info("Getting updated movie details via movie service")

	getMovieDetailsResp, err = movieClient.GetMovieDetails(ctx, &gen.GetMovieDetailsRequest{
		MovieId: m.Id,
	})
	if err != nil {
		log.Fatal("get movie details", zap.Error(err))
	}
	wantMovieDetails.Rating = wantRating
	if diff := cmp.Diff(getMovieDetailsResp.MovieDetails, wantMovieDetails, cmpopts.IgnoreUnexported(gen.MovieDetails{}, gen.Metadata{})); diff != "" {
		log.Fatal("get movie details after update mismatch", zap.String("diff", diff))
	}

	log.Info("Integration test execution successful")
}

func startMetadataService(ctx context.Context, registry discovery.Registry, log *zap.Logger, scope tally.Scope) *grpc.Server {
	log.Info("Starting metadata service on " + metadataServiceAddress)
	h := metadatatest.NewTestMetadataGRPCServer(log, scope)
	l, err := net.Listen("tcp", metadataServiceAddress)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
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
				log.Warn("Failed to deregister", zap.String("instanceServiceName", metadataServiceName), zap.Error(err))
			}
		}()
		if err := srv.Serve(l); err != nil {
			panic(err)
		}
	}()
	go func() {
		for {
			if err := registry.ReportHealthyState(id, metadataServiceName); err != nil {
				log.Warn("Failed to report healthy state", zap.Error(err))
			}
			time.Sleep(1 * time.Second)
		}
	}()
	return srv
}

func startRatingService(ctx context.Context, registry discovery.Registry, log *zap.Logger, scope tally.Scope) *grpc.Server {
	log.Info("Starting rating service on " + ratingServiceAddress)
	h := ratingtest.NewTestRatingGRPCServer(registry, log, scope)
	l, err := net.Listen("tcp", ratingServiceAddress)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
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
				log.Warn("Failed to deregister", zap.String("instanceServiceName", ratingServiceName), zap.Error(err))
			}
		}()
		if err := srv.Serve(l); err != nil {
			panic(err)
		}
	}()
	go func() {
		for {
			if err := registry.ReportHealthyState(id, ratingServiceName); err != nil {
				log.Warn("Failed to report healthy state", zap.Error(err))
			}
			time.Sleep(1 * time.Second)
		}
	}()
	return srv
}

func startMovieService(ctx context.Context, registry discovery.Registry, log *zap.Logger, scope tally.Scope) *grpc.Server {
	log.Info("Starting movie service on " + movieServiceAddress)
	h := movietest.NewTestMovieGRPCServer(registry, log, scope)
	l, err := net.Listen("tcp", movieServiceAddress)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
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
				log.Warn("Failed to deregister", zap.String("instanceServiceName", movieServiceName), zap.Error(err))
			}
		}()
		if err := srv.Serve(l); err != nil {
			panic(err)
		}
	}()
	go func() {
		for {
			if err := registry.ReportHealthyState(id, movieServiceName); err != nil {
				log.Warn("Failed to report healthy state", zap.Error(err))
			}
			time.Sleep(1 * time.Second)
		}
	}()
	return srv
}

func startAuthService(ctx context.Context, registry discovery.Registry, log *zap.Logger, scope tally.Scope) *grpc.Server {
	log.Info("Starting auth service on " + authServiceAddress)
	h := authtest.NewTestAuthGRPCServer(scope, log)
	l, err := net.Listen("tcp", authServiceAddress)
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}
	srv := grpc.NewServer()
	gen.RegisterAuthServiceServer(srv, h)
	id := discovery.GenerateInstanceID(authServiceName)
	if err := registry.Register(ctx, id, authServiceName, authServiceAddress); err != nil {
		panic(err)
	}
	go func() {
		defer func() {
			if err := registry.Deregister(ctx, id, authServiceName); err != nil {
				log.Warn("Failed to deregister", zap.String("instanceServiceName", authServiceName), zap.Error(err))
			}
		}()
		if err := srv.Serve(l); err != nil {
			panic(err)
		}
	}()
	go func() {
		for {
			if err := registry.ReportHealthyState(id, authServiceName); err != nil {
				log.Warn("Failed to report healthy state", zap.Error(err))
			}
			time.Sleep(1 * time.Second)
		}
	}()
	return srv
}
