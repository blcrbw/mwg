package grpc

import (
	"context"
	"fmt"
	"mmoviecom/gen"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SecretProvider defines a provider of secrets for auth handler.
type SecretProvider func() []byte

// Handler defines an auth handler.
type Handler struct {
	secretProvider SecretProvider
	gen.UnimplementedAuthServiceServer
}

// New creates a new auth gRPC handler.
func New(secretProvider SecretProvider) *Handler {
	return &Handler{secretProvider: secretProvider}
}

// GetToken performs verification of user credentials and returns a JWT token in case of success.
func (h *Handler) GetToken(ctx context.Context, req *gen.GetTokenRequest) (*gen.GetTokenResponse, error) {
	username, password := req.GetUsername(), req.GetPassword()
	if !validCredentials(username, password) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials")
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"iat":      time.Now().Unix(),
	})
	tokenString, err := token.SignedString(h.secretProvider())
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
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
	token, err := jwt.Parse(req.GetToken(), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return h.secretProvider(), nil
	})
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token")
	}
	var username string
	if v, ok := claims["username"]; ok {
		if v, ok := v.(string); ok {
			username = v
		}
	}
	return &gen.ValidateTokenResponse{Username: username}, nil
}
