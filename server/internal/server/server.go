package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/config"
)

// Server wraps the chi router, config, and DB connection.
type Server struct {
	cfg *config.Config
	db  *sql.DB
	mux *chi.Mux
}

// New builds a Server with a chi Mux, wires routes, and returns an http.Handler.
// Routes:
//
//	GET  /api/health                  — JSON health check (200 if DB reachable, 503 otherwise)
//	POST /api/auth/admin/login        — admin-only local login (bcrypt, env-seeded)
//	POST /api/auth/admin/refresh      — rotate the admin session
//	/*                               — static files from ../dist (SPA fallback to index.html)
//	                                 — /api and /media paths are never served statically
//
// The admin handler is optional: if nil, the admin routes are not mounted (a
// pure GitHub-auth server). No public registration route is ever mounted.
func New(cfg *config.Config, db *sql.DB, admin *auth.AdminHandler) *Server {
	s := &Server{
		cfg: cfg,
		db:  db,
		mux: chi.NewRouter(),
	}

	// Global middleware
	s.mux.Use(middleware.Logger)
	s.mux.Use(middleware.Recoverer)
	s.mux.Use(middleware.RequestID)

	// API routes with CORS
	s.mux.Route("/api", func(r chi.Router) {
		r.Use(s.corsMiddleware)
		r.Get("/health", s.handleHealth)
		if admin != nil {
			r.Route("/auth", func(ar chi.Router) {
				ar.Post("/admin/login", admin.Login)
				ar.Post("/admin/refresh", admin.Refresh)
			})
		}
	})

	// Static file serve for everything else (non-API, non-media)
	s.mux.NotFound(s.handleStatic)

	return s
}

// ServeHTTP implements http.Handler by delegating to the chi Mux.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// handleHealth responds with a JSON health check.
// It pings the database and returns:
//
//	200 {"status":"ok","db":"ok"}          if DB is reachable
//	503 {"status":"error","db":"<reason>"} if DB ping fails
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	dbStatus := "ok"
	httpStatus := http.StatusOK

	if err := s.db.Ping(); err != nil {
		dbStatus = "error: " + err.Error()
		httpStatus = http.StatusServiceUnavailable
	}

	resp := map[string]string{"status": "ok", "db": dbStatus}
	if httpStatus != http.StatusOK {
		resp["status"] = "error"
	}

	writeJSON(w, httpStatus, resp)
}

// handleStatic serves static files from the ../dist directory for paths that
// are not /api or /media. If the requested file exists it is served directly;
// otherwise index.html is served as an SPA fallback. If dist/ does not exist,
// a helpful HTML message is returned.
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	// Never serve /api or /media paths from static
	path := r.URL.Path
	if strings.HasPrefix(path, "/api") || strings.HasPrefix(path, "/media") {
		http.NotFound(w, r)
		return
	}

	distDir := filepath.Join("..", "dist")

	// Try to serve the exact file
	fp := filepath.Join(distDir, path)
	if fi, err := os.Stat(fp); err == nil && !fi.IsDir() {
		http.ServeFile(w, r, fp)
		return
	}

	// SPA fallback: serve index.html
	indexFile := filepath.Join(distDir, "index.html")
	if _, err := os.Stat(indexFile); err == nil {
		http.ServeFile(w, r, indexFile)
		return
	}

	// dist/ does not exist or is empty
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><title>GeoLibre Storymaps</title></head><body><h1>GeoLibre Storymaps</h1><p>Frontend build not found. Run <code>npm run build</code> in the project root, or start the dev server with <code>npm run dev</code> and use the Vite proxy.</p></body></html>`))
}

// corsMiddleware adds permissive CORS headers for API routes.
// Allowed origins come from the CORS_ORIGINS env var (defaults to dev localhost origins).
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigins := os.Getenv("CORS_ORIGINS")
		if allowedOrigins == "" {
			allowedOrigins = "http://localhost:5173,http://localhost:8080,http://127.0.0.1:5173,http://127.0.0.1:8080"
		}
		origin := r.Header.Get("Origin")
		for _, allowed := range strings.Split(allowedOrigins, ",") {
			if strings.TrimSpace(allowed) == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// writeJSON marshals v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON error: %v", err)
	}
}
