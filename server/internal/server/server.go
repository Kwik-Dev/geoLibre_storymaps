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

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/api"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/auth"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/config"
	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/media"
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
//	GET  /api/auth/whoami             — current user profile (requires Bearer JWT)
//	GET  /api/stories                 — public listing (anon) / filtered by role (auth)
//	POST /api/stories                 — create a story (requires Bearer JWT)
//	GET/PUT/DELETE /api/stories/:id   — read / update / soft-delete (authz by owner/admin)
//	/*                               — static files from ../dist (SPA fallback to index.html)
//	                                 — /api and /media paths are never served statically
//
// The admin, whoami, and stories handlers are optional: if nil, those routes
// are not mounted (a pure GitHub-auth server can pass nil handlers). All /api
// routes run behind auth.RequireAuth, which enforces the public-route
// allowlist (GET /api/stories public listing, health, auth) and requires a
// valid Bearer JWT (or refresh cookie) for everything else. No public
// registration route is ever mounted.
func New(cfg *config.Config, db *sql.DB, admin *auth.AdminHandler, whoami *auth.WhoamiHandler) *Server {
	s := &Server{
		cfg: cfg,
		db:  db,
		mux: chi.NewRouter(),
	}

	// Global middleware
	s.mux.Use(middleware.Logger)
	s.mux.Use(middleware.Recoverer)
	s.mux.Use(middleware.RequestID)

	// API routes with CORS; every /api route runs behind RequireAuth (which
	// itself allowlists the public auth/health/stories-listing/export paths).
		auther := auth.NewAuthenticator(cfg.JWTSecret, false)
	// Serve + soft-delete handler (P4.4). auther is used for optional auth on
	// the public GET /media/:aid route.
	mediaHandler := media.NewMediaHandler(db, cfg.MediaDir, auther)
	s.mux.Route("/api", func(r chi.Router) {
		r.Use(s.corsMiddleware)
		r.Use(auther.RequireAuth)
		r.Get("/health", s.handleHealth)
		if admin != nil {
			r.Route("/auth", func(ar chi.Router) {
				ar.Post("/admin/login", admin.Login)
				ar.Post("/admin/refresh", admin.Refresh)
			})
		}
		if whoami != nil {
			r.Get("/auth/whoami", whoami.ServeHTTP)
		}
		// Stories CRUD (P3.1). The handler enforces visibility/authz per-op;
		// the middleware already allowlists the public GET /api/stories list.
		api.NewStoriesHandler(db, auther).Routes(r)
		// Nested chapters CRUD + reorder (P3.2). Every op loads the parent
		// story and runs canAccess, so authz is enforced on every route. The
		// configured media external-URL allow-list (P4.3) gates media_ref_type
		// = external chapters.
		api.NewChaptersHandler(db).SetAllowedMediaHosts(cfg.AllowedMediaHosts).Routes(r)
		// Legacy story-JSON export (P3.4). The route is allowlisted by the
		// middleware; the handler enforces the public/owner/admin check.
		api.NewExportHandler(db, auther).Routes(r)
		// Media upload (P4.1). Requires a session (RequireAuth); the handler
		// validates by magic bytes, caps the size, and stores a random-named
		// file under MEDIA_DIR.
		r.Post("/media/upload", media.NewUploadHandler(db, cfg.MediaDir, cfg.MaxUploadBytes).ServeHTTP)
		// Media soft-delete (P4.4). Owner/admin only; the route is behind
		// RequireAuth so an authenticated user is always present.
		r.Delete("/media/{aid}", mediaHandler.Delete)
	})

	// Serve uploaded media (P4.4). This route lives OUTSIDE /api (it is not
	// behind RequireAuth) because public stories' assets must be served to
	// anonymous browsers. The handler performs optional auth itself so the
	// owner/admin of a private story can still reach its bytes.
	s.mux.Get("/media/{aid}", mediaHandler.Serve)

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
