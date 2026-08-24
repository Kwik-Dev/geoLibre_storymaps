package server

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
)

// purgeBatchSize bounds each DELETE so a single purge pass never runs one huge
// transaction (avoids holding the DB write lock for a long time). The purge
// loops until a batch returns fewer rows than this, so it is fully idempotent
// and self-terminating.
const purgeBatchSize = 500

// PurgeResult reports how many rows were hard-deleted in one purge pass.
type PurgeResult struct {
	Stories     int64
	Chapters    int64
	MediaAssets int64
}

// PurgeOnce hard-deletes rows whose deleted_at is set AND older than ttl.
//
// It is idempotent and processes in bounded batches. Only rows that already
// have deleted_at set are ever removed — purge is irreversible, so live rows
// are never touched. Foreign keys are respected: chapters referencing a purged
// media_asset are detached first, and chapters belonging to a purged story are
// removed so the story row can be deleted.
func PurgeOnce(db *sql.DB, ttl time.Duration) (PurgeResult, error) {
	var res PurgeResult
	if ttl <= 0 {
		return res, fmt.Errorf("PURGE_TTL must be > 0")
	}
	cutoff := time.Now().Add(-ttl).UTC().Format("2006-01-02 15:04:05")

	// 1. Detach media_asset references from chapters that are about to be
	//    purged so the FK on chapters.media_asset_id doesn't block deletion.
	if _, err := db.Exec(`
		UPDATE chapters SET media_asset_id = NULL
		WHERE media_asset_id IN (
			SELECT id FROM media_assets
			WHERE deleted_at IS NOT NULL AND deleted_at < ?
		)`, cutoff); err != nil {
		return res, fmt.Errorf("detach purged media refs: %w", err)
	}

	// 2. Delete soft-deleted chapters older than cutoff, in bounded batches.
	for {
		r, err := db.Exec(`
			DELETE FROM chapters
			WHERE id IN (
				SELECT id FROM chapters
				WHERE deleted_at IS NOT NULL AND deleted_at < ?
				LIMIT ?
			)`, cutoff, purgeBatchSize)
		if err != nil {
			return res, fmt.Errorf("purge chapters: %w", err)
		}
		n, _ := r.RowsAffected()
		res.Chapters += n
		if n < purgeBatchSize {
			break
		}
	}

	// 3. Delete chapters belonging to stories being purged (so the FK on
	//    chapters.story_id doesn't block story deletion), then the stories.
	for {
		r, err := db.Exec(`
			DELETE FROM chapters
			WHERE story_id IN (
				SELECT id FROM stories
				WHERE deleted_at IS NOT NULL AND deleted_at < ?
				LIMIT ?
			)`, cutoff, purgeBatchSize)
		if err != nil {
			return res, fmt.Errorf("purge orphan chapters: %w", err)
		}
		n, _ := r.RowsAffected()
		res.Chapters += n
		if n < purgeBatchSize {
			break
		}
	}
	for {
		r, err := db.Exec(`
			DELETE FROM stories
			WHERE id IN (
				SELECT id FROM stories
				WHERE deleted_at IS NOT NULL AND deleted_at < ?
				LIMIT ?
			)`, cutoff, purgeBatchSize)
		if err != nil {
			return res, fmt.Errorf("purge stories: %w", err)
		}
		n, _ := r.RowsAffected()
		res.Stories += n
		if n < purgeBatchSize {
			break
		}
	}

	// 4. Delete soft-deleted media_assets older than cutoff.
	for {
		r, err := db.Exec(`
			DELETE FROM media_assets
			WHERE id IN (
				SELECT id FROM media_assets
				WHERE deleted_at IS NOT NULL AND deleted_at < ?
				LIMIT ?
			)`, cutoff, purgeBatchSize)
		if err != nil {
			return res, fmt.Errorf("purge media_assets: %w", err)
		}
		n, _ := r.RowsAffected()
		res.MediaAssets += n
		if n < purgeBatchSize {
			break
		}
	}

	return res, nil
}

// PurgeJob runs PurgeOnce on a ticker and once on startup. Stop() halts it.
type PurgeJob struct {
	stop chan struct{}
	done chan struct{}
}

// StartPurge launches a background goroutine that runs a purge pass once on
// startup and then every interval. ttl must be > 0 (PURGE_TTL). Returns a
// *PurgeJob that can be stopped with Stop(). This is the P7.3 HANDOFF.
func StartPurge(db *sql.DB, ttl, interval time.Duration) *PurgeJob {
	j := &PurgeJob{stop: make(chan struct{}), done: make(chan struct{})}
	go func() {
		defer close(j.done)
		// Run once on startup.
		if _, err := PurgeOnce(db, ttl); err != nil {
			log.Printf("purge (startup): %v", err)
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if _, err := PurgeOnce(db, ttl); err != nil {
					log.Printf("purge (tick): %v", err)
				}
			case <-j.stop:
				return
			}
		}
	}()
	return j
}

// Stop halts the purge job and waits for it to exit.
func (j *PurgeJob) Stop() {
	close(j.stop)
	<-j.done
}

// PurgeTTLFromEnv returns the PURGE_TTL duration (default 30 days). It is used
// to wire StartPurge from the environment. A non-positive or unparseable value
// falls back to the default.
func PurgeTTLFromEnv() time.Duration {
	if v := os.Getenv("PURGE_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return 30 * 24 * time.Hour
}
