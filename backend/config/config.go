package config

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/ahmdfkhri/hydrocast/backend/internal/types"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseConfig
	JWTConfig
	GRPCConfig
	AdminConfig
}

type DatabaseConfig struct {
	Host string
	Port int
	User string
	Pass string
	Name string
}

type JWTConfig struct {
	Secret            []byte
	TokenExpiry       map[types.TokenType]time.Duration
	ReRefreshDuration time.Duration
}

type GRPCConfig struct {
	Port int
}

type AdminConfig struct {
	Username string
	Email    string
	Password string
}

func New() *Config {
	// load .env
	godotenv.Load(filepath.Join("..", ".env"))

	return &Config{
		DatabaseConfig{
			Host: getEnv("DB_HOST", "127.0.0.1"),
			Port: getEnvAsInt("DB_PORT", 5432),
			User: getEnv("DB_USER", "postgres"),
			Pass: getEnv("DB_PASS", ""),
			Name: getEnv("DB_NAME", "hydrocast"),
		},
		JWTConfig{
			Secret: []byte(getEnv("JWT_SECRET", "secret-key")),
			TokenExpiry: map[types.TokenType]time.Duration{
				types.TT_Access:  getEnvAsDuration("JWT_ACCESS_EXPIRY", 15*time.Minute),
				types.TT_Refresh: getEnvAsDuration("JWT_REFRESH_EXPIRY", 72*time.Hour),
			},
			ReRefreshDuration: getEnvAsDuration("JWT_REREFRESH_DURATION", 24*time.Hour),
		},
		GRPCConfig{
			Port: getEnvAsInt("GRPC_PORT", 50051),
		},
		AdminConfig{
			Username: getEnv("ADMIN_USERNAME", "admin"),
			Email:    getEnv("ADMIN_EMAIL", "example@email.com"),
			Password: getEnv("ADMIN_PASSWORD", "admin123"),
		},
	}
}

func getEnv(key string, defaultValue string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	return val
}

func getEnvAsInt(key string, defaultValue int) int {
	strVal, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	val, err := strconv.Atoi(strVal)
	if err != nil {
		return defaultValue
	}

	return val
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	strVal, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	val, err := time.ParseDuration(strVal)
	if err != nil {
		return defaultValue
	}

	return val
}
