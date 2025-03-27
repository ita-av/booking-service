package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	// key used to verify JWTs (must match user service's key)
	// prod. -> environment variables
	JWTSecret = []byte("secret_key_123")

	// Errors
	ErrMissingMetadata = errors.New("missing metadata")
	ErrMissingToken    = errors.New("missing token")
	ErrInvalidToken    = errors.New("invalid token")
)

// Claims represents the JWT payload with is_barber field
type Claims struct {
	IsBarber bool `json:"is_barber"`
	jwt.RegisteredClaims
}

// ExtractToken gets the token from gRPC metadata
func ExtractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", ErrMissingMetadata
	}

	// Token is expected in "authorization" header as "Bearer <token>"
	values := md.Get("authorization")
	if len(values) == 0 {
		return "", ErrMissingToken
	}

	authHeader := values[0]
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", ErrInvalidToken
	}

	return parts[1], nil
}

// VerifyToken validates the JWT and returns the claims
func VerifyToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// AuthInterceptor is a gRPC interceptor that checks for valid JWT tokens
func AuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	// Skip auth for health check or other public methods
	if isPublicMethod(info.FullMethod) {
		return handler(ctx, req)
	}

	// Extract token from context
	token, err := ExtractToken(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "authentication error: %v", err)
	}

	// Verify the token
	claims, err := VerifyToken(token)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
	}

	// Add claims to the context for use in handlers
	newCtx := context.WithValue(ctx, "user_claims", claims)

	// Continue execution of the handler
	return handler(newCtx, req)
}

// isPublicMethod determines if a method doesn't require authentication
func isPublicMethod(method string) bool {
	publicMethods := map[string]bool{
		"/grpc.health.v1.Health/Check": true,
		// Add other public methods here
	}
	return publicMethods[method]
}

// GetUserIDFromContext extracts the user ID from the context
func GetUserIDFromContext(ctx context.Context) (string, error) {
	claims, ok := ctx.Value("user_claims").(*Claims)
	if !ok || claims == nil {
		return "", errors.New("no user claims found in context")
	}

	// Extract the user ID from the subject field
	userID := claims.Subject
	if userID == "" {
		return "", errors.New("no user ID in token")
	}

	return userID, nil
}

// IsBarber checks if the user in the context has the barber flag set to true
func IsBarber(ctx context.Context) bool {
	claims, ok := ctx.Value("user_claims").(*Claims)
	if !ok || claims == nil {
		return false
	}

	return claims.IsBarber
}
