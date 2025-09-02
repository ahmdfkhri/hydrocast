package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ahmdfkhri/hydrocast/backend/config"
	"github.com/ahmdfkhri/hydrocast/backend/internal/client"
	"github.com/ahmdfkhri/hydrocast/backend/internal/types"
	"github.com/ahmdfkhri/hydrocast/backend/pkg/pb"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	// Get environment variables
	cfg := config.New()

	// Create new gRPC client
	grpcClient := client.NewGRPCClient(&cfg.ServerConfig)
	defer grpcClient.Close()

	// New context
	ctx := context.Background()

	// Try logging in
	loginResp, err := grpcClient.AuthClient.Login(ctx, &pb.LoginRequest{
		UsernameOrEmail: cfg.AdminConfig.Username,
		Password:        cfg.AdminConfig.Password,
	})
	if err != nil {
		log.Fatalf("failed to login into admin: %v", err)
	}
	fmt.Printf("Login Response:\n%+v\n", loginResp)

	// Create new context with access token
	md := metadata.New(map[string]string{
		string(types.MD_Authorization): "Bearer " + loginResp.AccessToken,
	})
	newCtx := metadata.NewOutgoingContext(ctx, md)

	// Try getting the user profile
	getProfileResp, err := grpcClient.AuthClient.GetProfile(newCtx, &emptypb.Empty{})
	if err != nil {
		log.Fatalf("failed to get profile: %v", err)
	}
	fmt.Printf("GetProfile Response:\n%+v\n", getProfileResp)

	// Create new context with refresh
	md = metadata.New(map[string]string{
		string(types.MD_Authorization): "Bearer " + loginResp.RefreshToken,
	})
	newCtx = metadata.NewOutgoingContext(ctx, md)

	// Try refreshing token with refresh_token
	refreshTokenResp, err := grpcClient.AuthClient.RefreshToken(newCtx, &pb.RefreshTokenRequest{
		RefreshToken: loginResp.RefreshToken,
	})
	if err != nil {
		log.Fatalf("failed to refresh token: %v", err)
	}
	fmt.Printf("RefreshToken Response:\n%+v\n", refreshTokenResp)

}
