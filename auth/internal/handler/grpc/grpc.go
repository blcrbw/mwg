package grpc

import (
	"context"
	"fmt"
	"mmoviecom/gen"
	"mmoviecom/pkg/logging"
	"mmoviecom/pkg/metrics"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/uber-go/tally/v6"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SecretProvider defines a provider of secrets for auth handler.
type SecretProvider func() []byte

// Handler defines an auth handler.
type Handler struct {
	secretProvider SecretProvider
	gen.UnimplementedAuthServiceServer
	getTokenMetrics      *metrics.EndpointMetrics
	validateTokenMetrics *metrics.EndpointMetrics
	logger               *zap.Logger
}

// New creates a new auth gRPC handler.
func New(secretProvider SecretProvider, scope tally.Scope, logger *zap.Logger) *Handler {
	logger = logger.With(
		zap.String(logging.FieldComponent, "handler"),
		zap.String(logging.FieldType, "grpc"),
	)
	return &Handler{
		secretProvider:       secretProvider,
		getTokenMetrics:      metrics.NewEndpointMetrics(scope, "GetToken"),
		validateTokenMetrics: metrics.NewEndpointMetrics(scope, "ValidateToken"),
		logger:               logger,
	}
}

// GetToken performs verification of user credentials and returns a JWT token in case of success.
func (h *Handler) GetToken(ctx context.Context, req *gen.GetTokenRequest) (*gen.GetTokenResponse, error) {
	h.getTokenMetrics.Calls.Inc(1)
	username, password := req.GetUsername(), req.GetPassword()
	if !validCredentials(username, password) {
		h.getTokenMetrics.InvalidArgumentErrors.Inc(1)
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"iat":      time.Now().Unix(),
	})
	tokenString, err := token.SignedString(h.secretProvider())
	if err != nil {
		h.getTokenMetrics.InternalErrors.Inc(1)
		h.logger.Error("Failed to sign token", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to sign token")
	}
	h.getTokenMetrics.Successes.Inc(1)
	return &gen.GetTokenResponse{Token: tokenString}, nil
}

// validCredential imitate validation.
func validCredentials(username, password string) bool {
	if len(username) == 0 || len(password) == 0 {
		return false
	}
	return true
}

// ValidateToken performs JWT token validation.
func (h *Handler) ValidateToken(ctx context.Context, req *gen.ValidateTokenRequest) (*gen.ValidateTokenResponse, error) {
	h.validateTokenMetrics.Calls.Inc(1)
	token, err := jwt.Parse(req.GetToken(), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			h.validateTokenMetrics.InvalidArgumentErrors.Inc(1)
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.secretProvider(), nil
	})
	if err != nil {
		h.validateTokenMetrics.InvalidArgumentErrors.Inc(1)
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		h.validateTokenMetrics.InvalidArgumentErrors.Inc(1)
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}
	var username string
	if v, ok := claims["username"]; ok {
		if v, ok := v.(string); ok {
			username = v
		}
	}
	h.validateTokenMetrics.Successes.Inc(1)
	return &gen.ValidateTokenResponse{Username: username}, nil
}
