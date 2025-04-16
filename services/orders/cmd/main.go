package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/api"
	"github.com/ExonegeS/go-ecom-services-grpc/services/orders/internal/config"
	"github.com/ExonegeS/prettyslog"

	_ "github.com/lib/pq"
)

func main() {
	cfg := config.NewConfig(".env")
	logger := prettyslog.SetupPrettySlog(os.Stdout)

	connStr := cfg.DB.MakeConnectionString()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("could not open database: %v", err)
	}

	server := api.NewAPIServer(cfg, db, logger)
	if err := server.Run(); err != nil {
		log.Fatal(err)
	}
}
