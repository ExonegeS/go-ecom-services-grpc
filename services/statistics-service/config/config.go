package config

import (
	"fmt"
	"os"
	"strconv"
)

type (
	Config struct {
		Server   Server
		NATS     NATSConfig
		Database DatabaseConfig
		GRPC     GRPCConfig
	}

	Server struct {
		Address string
		Port    string
	}

	NATSConfig struct {
		URL       string
		ClusterID string
		ClientID  string
	}

	DatabaseConfig struct {
		DBUser     string
		DBPassword string
		DBHost     string
		DBPort     string
		DBName     string
	}

	GRPCConfig struct {
		Port string
	}
)

func NewConfig() *Config {
	return &Config{
		Server{
			Address: getEnvStr("ADDRESS", "localhost"),
			Port:    getEnvStr("GRPC_PORT", "50053"),
		},
		NATSConfig{
			URL:       getEnvStr("NATS_URL", "nats://localhost:4222"),
			ClusterID: getEnvStr("NATS_CLUSTER_ID", "test-cluster"),
			ClientID:  getEnvStr("NATS_CLIENT_ID", "statistics-service"),
		},
		DatabaseConfig{
			DBUser:     getEnvStr("POSTGRES_USER", "admin"),
			DBPassword: getEnvStr("POSTGRES_PASSWORD", "adminadmin"),
			DBHost:     getEnvStr("DB_HOST", "localhost"),
			DBPort:     getEnvStr("DB_PORT", "5434"),
			DBName:     getEnvStr("POSTGRES_DB", "stats_db"),
		},
		GRPCConfig{
			Port: getEnvStr("ORDERS_GRPC_PORT", "50052"),
		},
	}
}

func (d *DatabaseConfig) MakeConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		d.DBHost, d.DBPort, d.DBUser, d.DBPassword, d.DBName,
	)
}

func getEnvStr(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvInt64(key string, fallback int64) int64 {
	if value, ok := os.LookupEnv(key); ok {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fallback
		}

		return i
	}

	return fallback
}
