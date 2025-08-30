package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ahmdfkhri/hydrocast/backend/config"
	"github.com/ahmdfkhri/hydrocast/backend/internal/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	ErrInvalidTokenType  = errors.New("invalid token type")
	ErrTokenInvalid      = errors.New("token is invalid")
	ErrTokenTypeMismatch = errors.New("token type mismatch")
)

type Authorizer struct {
	secret            []byte
	tokenExpiry       map[types.TokenType]time.Duration
	ReRefreshDuration time.Duration
}

func New(conf *config.JWTConfig) *Authorizer {
	return &Authorizer{
		secret:            conf.Secret,
		tokenExpiry:       conf.TokenExpiry,
		ReRefreshDuration: conf.ReRefreshDuration,
	}
}

type Claims struct {
	jwt.RegisteredClaims
	UserID    uuid.UUID       `json:"user_id"`
	UserRole  types.UserRole  `json:"user_role"`
	TokenType types.TokenType `json:"token_type"`
}

func (a *Authorizer) GenerateToken(userID uuid.UUID, userRole types.UserRole, tokenType types.TokenType) (string, error) {
	// Get expiry duration for the token type
	expiry, exists := a.tokenExpiry[tokenType]
	if !exists {
		return "", ErrInvalidTokenType
	}

	claims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Subject:   userID.String(),
			ID:        uuid.NewString(),
		},
		UserID:    userID,
		UserRole:  userRole,
		TokenType: tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenStr, err := token.SignedString(a.secret)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (a *Authorizer) VerifyToken(tokenString string, expectedTokenType types.TokenType) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.secret, nil
	})

	if err != nil {
		return nil, err
	}

	// Extract claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	// Verify token type matches expected type
	if claims.TokenType != expectedTokenType {
		return nil, ErrTokenTypeMismatch
	}

	return claims, nil
}

func (a *Authorizer) GenerateAccessToken(userID uuid.UUID, userRole types.UserRole) (string, error) {
	return a.GenerateToken(userID, userRole, types.TT_Access)
}

func (a *Authorizer) GenerateRefreshToken(userID uuid.UUID, userRole types.UserRole) (string, error) {
	return a.GenerateToken(userID, userRole, types.TT_Refresh)
}

func (a *Authorizer) VerifyAccessToken(tokenString string) (*Claims, error) {
	return a.VerifyToken(tokenString, types.TT_Access)
}

func (a *Authorizer) VerifyRefreshToken(tokenString string) (*Claims, error) {
	return a.VerifyToken(tokenString, types.TT_Refresh)
}

// GetUserID extracts the user ID (UUID) from gRPC metadata
func GetUserID(ctx context.Context) (*uuid.UUID, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "missing metadata in context")
	}

	values := md[string(types.MD_UserID)]
	if len(values) == 0 {
		return nil, status.Error(codes.Unauthenticated, "user id not found in metadata")
	}

	id, err := uuid.Parse(values[0])
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user id format: %v", err)
	}

	return &id, nil
}
