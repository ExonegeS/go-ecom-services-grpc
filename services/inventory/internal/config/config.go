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
	DB          DataBase
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

func NewConfig(filename string) Config {
	if filename != "" {
		if err := loadEnv(filename); err != nil {
			panic(fmt.Sprintf("Error loading .env file: %v", err))
		}
	}

	return Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Version:     getEnv("VERSION", "v1"),
		Server: Server{
			Address:  getEnv("ADDRESS", ""),
			Port:     getEnv("PORT", "8080"),
			GRPCPort: getEnv("GRPC_PORT", "5050"),
		},
		DB: DataBase{
			DBUser:     getEnv("DB_USER", "admin"),
			DBPassword: getEnv("DB_PASSWORD", "admin"),
			DBHost:     getEnv("DB_HOST", "db"),
			DBPort:     getEnv("DB_PORT", "5432"),
			DBName:     getEnv("DB_NAME", "inventory_db"),
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
