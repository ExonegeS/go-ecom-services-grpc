package main

import (
	"os"

	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/config"
	"github.com/ExonegeS/go-ecom-services-grpc/services/statistics/internal/app"
	"github.com/ExonegeS/prettyslog"
)

func main() {
	cfg := config.NewConfig()
	logger := prettyslog.SetupPrettySlog(os.Stdout)

	server := app.NewAPIServer(cfg, logger)
	server.Run()
}
