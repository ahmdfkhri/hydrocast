package interceptor

import (
	"context"
	"log"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/ahmdfkhri/hydrocast/backend/internal/types"
	"github.com/ahmdfkhri/hydrocast/backend/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/authz"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var errMissingMetadata = status.Error(codes.InvalidArgument, "missing metadata") // Doesn't need to be wrapped with grpc status

// Predetermined slices of publically accessible services methods
var publicEndpoints = []string{
	"/hydrocast.Auth/Login",
	"/hydrocast.Auth/Signup",
	"/hydrocast.Auth/RefreshToken",
}

// Wrapper for stream interceptor
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

func newWrappedStream(ctx context.Context, s grpc.ServerStream) grpc.ServerStream {
	return &wrappedStream{s, ctx}
}

func AuthInterceptors(a *auth.Authorizer) (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor) {
	unary := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		// skip public endpoints
		if slices.Contains(publicEndpoints, info.FullMethod) {
			return handler(ctx, req)
		}

		tokenStr, err := parseTokenFromContext(ctx)
		if err != nil {
			return nil, err
		}

		claims, err := a.VerifyAccessToken(tokenStr)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "%v", err)
		}

		newCtx := newContextFromClaims(ctx, claims)
		return handler(newCtx, req)
	}

	stream := func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// Ignore public endpoints
		if slices.Contains(publicEndpoints, info.FullMethod) {
			return handler(srv, ss)
		}

		tokenStr, err := parseTokenFromContext(ss.Context())
		if err != nil {
			return err
		}

		claims, err := a.VerifyAccessToken(tokenStr)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "%v", err)
		}

		newCtx := newContextFromClaims(ss.Context(), claims)
		return handler(srv, newWrappedStream(newCtx, ss))
	}

	return unary, stream
}

func AuthzInterceptors() (grpc.UnaryServerInterceptor, grpc.StreamServerInterceptor) {
	fw, err := authz.NewFileWatcher(filepath.Join("config", "grpc", "policy.json"), 100*time.Millisecond)
	if err != nil {
		log.Fatalf("Creating a static authz interceptor: %v", err)
	}
	log.Printf("Using file watcher authz policy at %v\n", filepath.Join("config", "grpc", "policy.json"))
	return fw.UnaryInterceptor, fw.StreamInterceptor
}

// Check the existence and validity of authorization header
func parseTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errMissingMetadata
	}

	authorization := md[string(types.MD_Authorization)]
	if len(authorization) < 1 {
		return "", errMissingMetadata
	}

	authHeader := authorization[0]
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", status.Error(codes.InvalidArgument, "invalid authorization header format")
	}

	return strings.TrimPrefix(authHeader, "Bearer "), nil
}

// Appends original context with user_id and user_role data
func newContextFromClaims(ctx context.Context, claims *auth.Claims) context.Context {
	existingMD, _ := metadata.FromIncomingContext(ctx)
	newMD := existingMD.Copy()
	newMD.Set(string(types.MD_UserID), claims.UserID.String())
	newMD.Set(string(types.MD_UserRole), string(claims.UserRole))
	return metadata.NewIncomingContext(ctx, newMD)
}
