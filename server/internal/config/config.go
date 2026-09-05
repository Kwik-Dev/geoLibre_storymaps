package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds all server configuration values, populated from environment
// variables with sensible defaults.
type Config struct {
	DataDir           string   // root data directory (default ./data)
	DBPath            string   // SQLite database path (default $DATA_DIR/sqlite.db)
	JWTSecret         string   // REQUIRED: JWT signing secret
	AdminEmail        string   // optional: seeded admin email
	AdminPassword     string   // optional: seeded admin password
	MediaDir          string   // directory for uploaded media (default $DATA_DIR/media)
	StoreKind         string   // upload store kind: "local" (default) or "s3"
	AllowedMediaHosts  []string // optional comma-separated allow-list (empty = permissive)
	MaxUploadBytes     int64    // maximum upload size in bytes (default 25 MB)
	AppPort            int      // HTTP listen port (default 8080)
	ModerationRequired bool     // when true, stories set to public go to status=pending (P7.2)
}

const (
	defaultDataDir        = "./data"
	defaultMaxUploadBytes = 25 * 1024 * 1024 // 25 MB
	defaultAppPort        = 8080
)

// Load reads configuration from environment variables and returns a populated
// Config struct. It returns an error if JWT_SECRET is empty.
func Load() (*Config, error) {
	dataDir := getEnv("DATA_DIR", defaultDataDir)

	cfg := &Config{
		DataDir:       dataDir,
		DBPath:        getEnv("DB_PATH", dataDir+"/sqlite.db"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		AdminEmail:    os.Getenv("ADMIN_EMAIL"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		MediaDir:      getEnv("MEDIA_DIR", dataDir+"/media"),
		StoreKind:     getEnv("STORE_KIND", "local"),
		MaxUploadBytes: getEnvInt64("MEDIA_MAX_BYTES", defaultMaxUploadBytes),
		AppPort:       getEnvInt("APP_PORT", defaultAppPort),
		ModerationRequired: os.Getenv("MODERATION_REQUIRED") == "1",
	}

	if hosts := os.Getenv("ALLOWED_MEDIA_HOSTS"); hosts != "" {
		for _, h := range strings.Split(hosts, ",") {
			h = strings.TrimSpace(h)
			if h != "" {
				cfg.AllowedMediaHosts = append(cfg.AllowedMediaHosts, h)
			}
		}
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required but not set")
	}

	return cfg, nil
}

// getEnv returns the value of the environment variable named by key, or
// fallback if the variable is not set or empty.
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// getEnvInt returns the integer value of the environment variable named by
// key, or fallback if unset or unparseable.
func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

// getEnvInt64 returns the int64 value of the environment variable named by
// key, or fallback if unset or unparseable.
func getEnvInt64(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return fallback
}
