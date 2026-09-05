package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Unset all related env vars to test defaults.
	for _, key := range []string{
		"DATA_DIR", "DB_PATH", "JWT_SECRET",
		"ADMIN_EMAIL", "ADMIN_PASSWORD",
		"MEDIA_DIR", "ALLOWED_MEDIA_HOSTS",
		"MEDIA_MAX_BYTES", "APP_PORT",
		"MODERATION_REQUIRED", "BASE_PATH",
	} {
		os.Unsetenv(key)
	}

	// When JWT_SECRET is empty, Load should return an error.
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is empty, got nil")
	}
	if err.Error() != "JWT_SECRET is required but not set" {
		t.Fatalf("unexpected error message: %v", err)
	}

	// Set JWT_SECRET and verify defaults.
	os.Setenv("JWT_SECRET", "test-secret")
	defer os.Unsetenv("JWT_SECRET")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DataDir != "./data" {
		t.Errorf("expected DATA_DIR default ./data, got %q", cfg.DataDir)
	}

	expectedDBPath := "./data/sqlite.db"
	if cfg.DBPath != expectedDBPath {
		t.Errorf("expected DB_PATH derived from DATA_DIR: %q, got %q", expectedDBPath, cfg.DBPath)
	}

	expectedMediaDir := "./data/media"
	if cfg.MediaDir != expectedMediaDir {
		t.Errorf("expected MEDIA_DIR derived from DATA_DIR: %q, got %q", expectedMediaDir, cfg.MediaDir)
	}

	if cfg.MaxUploadBytes != 25*1024*1024 {
		t.Errorf("expected MEDIA_MAX_BYTES default %d, got %d", 25*1024*1024, cfg.MaxUploadBytes)
	}

	if cfg.AppPort != 8080 {
		t.Errorf("expected APP_PORT default %d, got %d", 8080, cfg.AppPort)
	}

	if cfg.ModerationRequired != false {
		t.Errorf("expected MODERATION_REQUIRED default false, got %v", cfg.ModerationRequired)
	}

	if cfg.BasePath != "" {
		t.Errorf("expected BASE_PATH default empty, got %q", cfg.BasePath)
	}

	if cfg.AllowedMediaHosts != nil {
		t.Errorf("expected ALLOWED_MEDIA_HOSTS default nil, got %v", cfg.AllowedMediaHosts)
	}

	if cfg.JWTSecret != "test-secret" {
		t.Errorf("expected JWT_SECRET test-secret, got %q", cfg.JWTSecret)
	}

	if cfg.AdminEmail != "" {
		t.Errorf("expected empty ADMIN_EMAIL, got %q", cfg.AdminEmail)
	}

	if cfg.AdminPassword != "" {
		t.Errorf("expected empty ADMIN_PASSWORD, got %q", cfg.AdminPassword)
	}
}

func TestLoadCustomEnv(t *testing.T) {
	os.Setenv("DATA_DIR", "/custom/data")
	os.Setenv("DB_PATH", "/custom/db.sqlite")
	os.Setenv("JWT_SECRET", "custom-secret")
	os.Setenv("ADMIN_EMAIL", "admin@example.com")
	os.Setenv("ADMIN_PASSWORD", "hunter2")
	os.Setenv("MEDIA_DIR", "/custom/media")
	os.Setenv("ALLOWED_MEDIA_HOSTS", "example.com, media.example.org")
	os.Setenv("MEDIA_MAX_BYTES", "52428800")
	os.Setenv("APP_PORT", "9090")
	os.Setenv("MODERATION_REQUIRED", "1")

	defer func() {
		for _, key := range []string{
			"DATA_DIR", "DB_PATH", "JWT_SECRET",
			"ADMIN_EMAIL", "ADMIN_PASSWORD",
			"MEDIA_DIR", "ALLOWED_MEDIA_HOSTS",
			"MEDIA_MAX_BYTES", "APP_PORT",
			"MODERATION_REQUIRED",
		} {
			os.Unsetenv(key)
		}
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.DataDir != "/custom/data" {
		t.Errorf("expected DATA_DIR /custom/data, got %q", cfg.DataDir)
	}
	if cfg.DBPath != "/custom/db.sqlite" {
		t.Errorf("expected DB_PATH /custom/db.sqlite, got %q", cfg.DBPath)
	}
	if cfg.JWTSecret != "custom-secret" {
		t.Errorf("expected JWT_SECRET custom-secret, got %q", cfg.JWTSecret)
	}
	if cfg.AdminEmail != "admin@example.com" {
		t.Errorf("expected ADMIN_EMAIL admin@example.com, got %q", cfg.AdminEmail)
	}
	if cfg.AdminPassword != "hunter2" {
		t.Errorf("expected ADMIN_PASSWORD hunter2, got %q", cfg.AdminPassword)
	}
	if cfg.MediaDir != "/custom/media" {
		t.Errorf("expected MEDIA_DIR /custom/media, got %q", cfg.MediaDir)
	}
	if len(cfg.AllowedMediaHosts) != 2 || cfg.AllowedMediaHosts[0] != "example.com" || cfg.AllowedMediaHosts[1] != "media.example.org" {
		t.Errorf("expected ALLOWED_MEDIA_HOSTS [example.com media.example.org], got %v", cfg.AllowedMediaHosts)
	}
	if cfg.MaxUploadBytes != 52428800 {
		t.Errorf("expected MEDIA_MAX_BYTES 52428800, got %d", cfg.MaxUploadBytes)
	}
	if cfg.AppPort != 9090 {
		t.Errorf("expected APP_PORT 9090, got %d", cfg.AppPort)
	}

	if cfg.ModerationRequired != true {
		t.Errorf("expected MODERATION_REQUIRED true, got %v", cfg.ModerationRequired)
	}
}
