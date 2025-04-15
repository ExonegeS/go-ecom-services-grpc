package main

import (
	"log"
	"os"

	"github.com/ExonegeS/go-ecom-services-grpc/services/gateway/internal/api"
	"github.com/ExonegeS/go-ecom-services-grpc/services/gateway/internal/config"
	"github.com/ExonegeS/prettyslog"
	"github.com/gorilla/mux"
)

func main() {
	// Start services
	cfg := config.NewConfig()
	logger := prettyslog.SetupPrettySlog(os.Stdout)

	router := mux.NewRouter()
	// Start the server
	server := api.NewAPIServer(router, cfg, logger)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
