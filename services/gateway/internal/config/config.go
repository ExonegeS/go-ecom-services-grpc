package config

import (
	"os"
)

type Config struct {
	Environment string
	Version     string
	Port        string
	Services    []Service
	CorsURLs    string
}

type Service struct {
	Name       string
	URLBase    string
	ApiVersion string
	Status     string
}

func NewConfig() Config {
	return Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Port:        getEnv("PORT", "8080"),
		Version:     getEnv("VERSION", "v1"),
		Services: []Service{
			{
				Name:       "inventory service",
				URLBase:    getEnv("INVENTORY_SERVICE_URL", "http://localhost:8081"),
				ApiVersion: getEnv("INVENTORY_SERVICE_API_VERSION", "v1"),
				Status:     "down",
			},
			{
				Name:       "orders service",
				URLBase:    getEnv("ORDERS_SERVICE_URL", "http://localhost:8082"),
				ApiVersion: getEnv("ORDERS_SERVICE_API_VERSION", "v1"),
				Status:     "down",
			},
		},
		CorsURLs: getEnv("CORS_URLS", "*"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
