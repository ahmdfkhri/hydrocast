package main

import (
	"fmt"
	"log"
	"net"
	"path/filepath"

	"github.com/ahmdfkhri/hydrocast/backend/config"
	"github.com/ahmdfkhri/hydrocast/backend/internal/database"
	"github.com/ahmdfkhri/hydrocast/backend/internal/interceptor"
	"github.com/ahmdfkhri/hydrocast/backend/internal/repository"
	"github.com/ahmdfkhri/hydrocast/backend/internal/service"
	"github.com/ahmdfkhri/hydrocast/backend/pkg/auth"
	"github.com/ahmdfkhri/hydrocast/backend/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Get environtment variables
	cfg := config.New()

	// Try to listen on GRPC_PORT
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", cfg.GRPCConfig.Port))
	if err != nil {
		log.Fatalf("Listening on local port %q: %v", cfg.GRPCConfig.Port, err)
	}

	// Create TLS-based credentials
	creds, err := credentials.NewServerTLSFromFile(filepath.Join("config", "x509", "server.crt"), filepath.Join("config", "x509", "server.key"))
	if err != nil {
		log.Fatalf("Loading credentials: %v", err)
	}

	// Initiate authorizer
	authorizer := auth.New(&cfg.JWTConfig)

	// Initiate database connection
	db := database.NewConnection(&cfg.DatabaseConfig)
	defer db.Close()

	// Initiate repositories
	userRepo := repository.NewUserRepository(db)

	// Initiate services
	authServer := service.NewAuthServer(userRepo, authorizer)

	// Initialize interceptors
	unaryAuth, streamAuth := interceptor.AuthInterceptors(authorizer)
	unaryAuthz, streamAuthz := interceptor.AuthzInterceptors()

	// Chain them
	unaryInt := grpc.ChainUnaryInterceptor(unaryAuth, unaryAuthz)
	streamInt := grpc.ChainStreamInterceptor(streamAuth, streamAuthz)

	// Create server
	s := grpc.NewServer(grpc.Creds(creds), unaryInt, streamInt)

	// Register services
	pb.RegisterAuthServer(s, authServer)

	// register reflection for testing
	reflection.Register(s)

	log.Printf("Serving on :%v\n", cfg.GRPCConfig.Port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Serving Echo service on local port: %v", err)
	}
}
