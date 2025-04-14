package main

import (
	"log"
	"net/http"
	"os"

	"github.com/ExonegeS/go-ecom-services-grpc/services/gateway/internal/api"
	"github.com/ExonegeS/go-ecom-services-grpc/services/gateway/internal/config"
	"github.com/ExonegeS/prettyslog"
)

func main() {
	// Start services
	cfg := config.NewConfig()
	logger := prettyslog.SetupPrettySlog(os.Stdout)

	mux := http.NewServeMux()
	// Start the server
	server := api.NewAPIServer(mux, cfg, logger)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
