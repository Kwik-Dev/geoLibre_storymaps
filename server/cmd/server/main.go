package main

import (
	"fmt"
	"log"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting server on port %d", cfg.AppPort)
	log.Printf("Data directory: %s", cfg.DataDir)
	log.Printf("Database path: %s", cfg.DBPath)
	log.Printf("Media directory: %s", cfg.MediaDir)
	log.Printf("Max upload bytes: %d", cfg.MaxUploadBytes)
	if cfg.AdminEmail != "" {
		log.Printf("Admin email: %s", cfg.AdminEmail)
	}
	if len(cfg.AllowedMediaHosts) > 0 {
		log.Printf("Allowed media hosts: %v", cfg.AllowedMediaHosts)
	}

	fmt.Println("Config loaded successfully — exiting.")
}
