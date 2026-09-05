package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/config"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/db"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Data directory: %s", cfg.DataDir)
	log.Printf("Database path: %s", cfg.DBPath)
	log.Printf("Media directory: %s", cfg.MediaDir)
	log.Printf("Max upload bytes: %d", cfg.MaxUploadBytes)
	log.Printf("App port: %d", cfg.AppPort)
	if cfg.AdminEmail != "" {
		log.Printf("Admin email: %s", cfg.AdminEmail)
	}
	if len(cfg.AllowedMediaHosts) > 0 {
		log.Printf("Allowed media hosts: %v", cfg.AllowedMediaHosts)
	}

	// Ensure the data + media directories exist before opening the DB, so a
	// fresh checkout starts cleanly without a manual mkdir.
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}
	if err := os.MkdirAll(cfg.MediaDir, 0o755); err != nil {
		log.Fatalf("Failed to create media directory: %v", err)
	}

	database, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed the admin-only local login user from the environment. If
	// ADMIN_EMAIL / ADMIN_PASSWORD are set, the row is upserted idempotently
	// with a fresh bcrypt hash; otherwise this is a no-op.
	adminHandler := auth.NewAdminHandler(auth.GitHubConfig{JWTSecret: cfg.JWTSecret}, database)
	if err := adminHandler.EnsureAdmin(cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Fatalf("Failed to seed admin user: %v", err)
	}

	whoamiHandler := auth.NewWhoamiHandler(database)

	srv := server.New(cfg, database, adminHandler, whoamiHandler)

	addr := ":" + strconv.Itoa(cfg.AppPort)
	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
