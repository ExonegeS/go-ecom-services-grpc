package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Environment string
	Version     string
	Server      Server
	Clients     map[string]Server
	DB          DataBase
	NATS        NATSConfig
}

type Server struct {
	Address  string
	Port     string
	GRPCPort string
}

type DataBase struct {
	DBUser     string
	DBPassword string
	DBHost     string
	DBPort     string
	DBName     string
}
type NATSConfig struct {
	URL string
}

func NewConfig(filename string) Config {
	if filename != "" {
		if err := loadEnv(filename); err != nil {
			panic(fmt.Sprintf("Error loading .env file: %v", err))
		}
	}

	clients := make(map[string]Server, 1)
	clients[getEnv("INVENTORY_CLIENT_NAME", "inventory client")] = Server{
		Address:  getEnv("INVENTORY_CLIENT_ADDRESS", "localhost"),
		GRPCPort: getEnv("INVENTORY_CLIENT_GRPC_PORT", "50051"),
	}
	return Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Version:     getEnv("VERSION", "v1"),
		Server: Server{
			Address:  getEnv("ADDRESS", "localhost"),
			GRPCPort: getEnv("GRPC_PORT", "50052"),
		},
		Clients: clients,
		DB: DataBase{
			DBUser:     getEnv("POSTGRES_USER", "admin"),
			DBPassword: getEnv("POSTGRES_PASSWORD", "admin"),
			DBHost:     getEnv("DB_HOST", "localhost"),
			DBPort:     getEnv("DB_PORT", "5432"),
			DBName:     getEnv("POSTGRES_DB", "orders_db"),
		},
		NATS: NATSConfig{
			URL: getEnv("NATS_URL", "nats://localhost:4222"),
		},
	}
}

func loadEnv(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key, value := parts[0], parts[1]
			if err := os.Setenv(key, value); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

func (d *DataBase) MakeConnectionString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		d.DBHost, d.DBPort, d.DBUser, d.DBPassword, d.DBName,
	)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
