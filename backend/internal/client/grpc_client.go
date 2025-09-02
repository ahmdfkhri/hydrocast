package client

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/ahmdfkhri/hydrocast/backend/config"
	"github.com/ahmdfkhri/hydrocast/backend/pkg/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Add other necessary clients later on
type GRPCClient struct {
	conn       *grpc.ClientConn
	AuthClient pb.AuthClient
}

func NewGRPCClient(conf *config.ServerConfig) *GRPCClient {
	// Create tls based credential
	caCert := filepath.Join("config", "x509", "ca.crt")
	creds, err := credentials.NewClientTLSFromFile(caCert, conf.Host)
	if err != nil {
		log.Fatalf("failed to load credentials: %v", err)
	}

	// Connect to gRPC service
	addr := fmt.Sprintf("%s:%d", conf.Host, conf.GRPCPort)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("grpc.NewClient(%q): %v", addr, err)
	}

	// Register service-specifc client
	authClient := pb.NewAuthClient(conn)

	return &GRPCClient{
		conn:       conn,
		AuthClient: authClient,
	}
}

func (g *GRPCClient) Close() {
	if err := g.conn.Close(); err != nil {
		log.Printf("error closing gRPC connection: %v", err)
	}
}
