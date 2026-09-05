package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Kwik-Dev/geoLibre_storymaps/server/internal/config"
)

// SeedDemo inserts a demo story with chapters if SEED_DEMO=1.
//
// It is idempotent: if a story with slug "demo-scrollytelling" already
// exists, SeedDemo returns early. If the environment variable SEED_DEMO
// is not "1", it is a no-op.
//
// The story author is the seeded admin user (from config) if one exists,
// otherwise a system placeholder user is created.
func SeedDemo(cfg *config.Config, db *sql.DB) error {
	if os.Getenv("SEED_DEMO") != "1" {
		return nil
	}

	const slug = "demo-scrollytelling"

	// Idempotency: skip if already seeded.
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM stories WHERE slug = ?`, slug).Scan(&count); err != nil {
		return fmt.Errorf("check existing demo story: %w", err)
	}
	if count > 0 {
		log.Printf("SeedDemo: story %q already exists, skipping", slug)
		return nil
	}

	// Resolve the author ID.
	authorID, err := resolveAuthor(cfg, db)
	if err != nil {
		return fmt.Errorf("resolve author for demo: %w", err)
	}

	// Insert the demo story.
	loc := locationJSON(13.4050, 52.5200, 12, 0, 0) // Berlin
	_, err = db.Exec(`
		INSERT INTO stories (slug, author_id, title, subtitle, byline, visibility, status, global_view, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'public', 'approved', ?, datetime('now'), datetime('now'))
	`, slug, authorID,
		"A Scrollytelling Tour of Berlin",
		"Explore the German capital through story-driven maps",
		"A demo story powered by GeoLibre Storymaps",
		loc,
	)
	if err != nil {
		return fmt.Errorf("insert demo story: %w", err)
	}

	// Get the new story ID.
	var storyID int64
	if err := db.QueryRow(`SELECT id FROM stories WHERE slug = ?`, slug).Scan(&storyID); err != nil {
		return fmt.Errorf("get demo story id: %w", err)
	}

	// Insert chapters.
	chapters := []struct {
		pos     int
		title   string
		desc    string
		loc     string
		align   string
		mediaT  string
		mediaR  string
	}{
		{
			pos:   0,
			title: "Welcome to Berlin",
			desc:  "## Brandenburger Tor\n\nThe **Brandenburg Gate** is an 18th-century neoclassical monument in Berlin, built on the site of a former city gate. It is one of the most recognizable landmarks of Germany.\n\nThis chapter demonstrates how Markdown content renders inside a scrollytelling card. You can use **bold**, *italic*, `code`, lists, and more.",
			loc:   locationJSON(13.3777, 52.5163, 15, 0, 0),
			align: "left",
			mediaT: "image",
			mediaR: "none",
		},
		{
			pos:   1,
			title: "Reichstag Building",
			desc:  "## Reichstag\n\nThe Reichstag building houses the German parliament, the Bundestag. Its iconic glass dome, designed by architect Norman Foster, offers a 360° view of the city.\n\n> \"This building is a symbol of Germany's democratic tradition.\"\n\n### Quick Facts\n- Completed: 1894\n- Dome opened: 1999\n- Height: 47 m (154 ft)",
			loc:   locationJSON(13.3733, 52.5186, 15, 0, 0),
			align: "right",
			mediaT: "image",
			mediaR: "none",
		},
	}

	for _, ch := range chapters {
		_, err = db.Exec(`
			INSERT INTO chapters (story_id, position, title, description_md, alignment, hidden, location, map_animation, rotate_animation, media_type, media_ref_type, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, 0, ?, 'flyTo', 0, ?, 'none', datetime('now'), datetime('now'))
		`, storyID, ch.pos, ch.title, ch.desc, ch.align, ch.loc, ch.mediaT)
		if err != nil {
			return fmt.Errorf("insert chapter %d: %w", ch.pos, err)
		}
	}

	log.Printf("SeedDemo: seeded demo story %q with %d chapters", slug, len(chapters))
	return nil
}

// resolveAuthor returns the ID of the seeded admin user if one exists,
// otherwise creates a system placeholder user with role "user".
func resolveAuthor(cfg *config.Config, db *sql.DB) (int64, error) {
	if cfg.AdminEmail != "" {
		var id int64
		err := db.QueryRow(`SELECT id FROM users WHERE admin_email = ?`, cfg.AdminEmail).Scan(&id)
		if err == nil {
			return id, nil
		}
		if err != sql.ErrNoRows {
			return 0, fmt.Errorf("lookup admin user: %w", err)
		}
	}

	// No admin — use or create a system user.
	var id int64
	err := db.QueryRow(`SELECT id FROM users WHERE github_login = 'system'`).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("lookup system user: %w", err)
	}

	res, err := db.Exec(`
		INSERT INTO users (github_login, role, created_at)
		VALUES ('system', 'user', datetime('now'))
	`)
	if err != nil {
		return 0, fmt.Errorf("create system user: %w", err)
	}
	return res.LastInsertId()
}

// locationJSON returns a JSON string suitable for the stories.global_view
// or chapters.location column.
func locationJSON(lng, lat float64, zoom, pitch, bearing float64) string {
	loc := map[string]interface{}{
		"center":  []float64{lng, lat},
		"zoom":    zoom,
		"pitch":   pitch,
		"bearing": bearing,
	}
	b, _ := json.Marshal(loc)
	return string(b)
}
