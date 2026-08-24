-- Version: 1
-- Initial schema for user-created storymaps
-- All CREATE statements use IF NOT EXISTS so the DDL is idempotent.

CREATE TABLE IF NOT EXISTS users (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    github_login TEXT,
    github_id    TEXT UNIQUE,
    admin_email   TEXT UNIQUE,
    password_hash TEXT,
    role         TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at   TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS stories (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    slug        TEXT NOT NULL UNIQUE,
    author_id   INTEGER NOT NULL REFERENCES users(id),
    title       TEXT NOT NULL DEFAULT '',
    subtitle    TEXT NOT NULL DEFAULT '',
    byline      TEXT NOT NULL DEFAULT '',
    visibility  TEXT NOT NULL DEFAULT 'private' CHECK (visibility IN ('private', 'public')),
    status      TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending', 'approved')),
    global_view TEXT,  -- JSON
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at  TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at  TEXT
);

CREATE TABLE IF NOT EXISTS chapters (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    story_id         INTEGER NOT NULL REFERENCES stories(id),
    position         INTEGER NOT NULL DEFAULT 0,
    title            TEXT NOT NULL DEFAULT '',
    description_md   TEXT NOT NULL DEFAULT '',
    alignment        TEXT NOT NULL DEFAULT 'left' CHECK (alignment IN ('left', 'center', 'right')),
    hidden           INTEGER NOT NULL DEFAULT 0,
    location         TEXT,  -- JSON
    map_animation    TEXT NOT NULL DEFAULT 'flyTo' CHECK (map_animation IN ('flyTo', 'easeTo')),
    rotate_animation INTEGER NOT NULL DEFAULT 0,
    on_chapter_enter TEXT,  -- JSON
    on_chapter_exit  TEXT,  -- JSON
    source           TEXT NOT NULL DEFAULT '',
    media_type       TEXT NOT NULL DEFAULT 'none' CHECK (media_type IN ('image', 'video', 'audio', 'none')),
    media_ref_type   TEXT NOT NULL DEFAULT 'none' CHECK (media_ref_type IN ('external', 'local', 'none')),
    media_external_url TEXT NOT NULL DEFAULT '',
    media_asset_id   INTEGER REFERENCES media_assets(id),
    created_at       TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at       TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at       TEXT
);

CREATE TABLE IF NOT EXISTS media_assets (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    kind        TEXT NOT NULL CHECK (kind IN ('image', 'video', 'audio')),
    stored_path TEXT NOT NULL,
    filename    TEXT NOT NULL DEFAULT '',
    bytes       INTEGER NOT NULL DEFAULT 0,
    mime        TEXT NOT NULL DEFAULT '',
    created_at  TEXT NOT NULL DEFAULT (datetime('now')),
    deleted_at  TEXT
);

-- Indexes for common query paths
CREATE INDEX IF NOT EXISTS idx_stories_slug          ON stories(slug);
-- Case-insensitive uniqueness for generated slugs (P3.1).
CREATE UNIQUE INDEX IF NOT EXISTS idx_stories_slug_ci    ON stories(lower(slug));
CREATE INDEX IF NOT EXISTS idx_stories_author_id     ON stories(author_id);
CREATE INDEX IF NOT EXISTS idx_chapters_story_id     ON chapters(story_id);
CREATE INDEX IF NOT EXISTS idx_chapters_media_asset_id ON chapters(media_asset_id);
CREATE INDEX IF NOT EXISTS idx_stories_deleted_at    ON stories(deleted_at);
CREATE INDEX IF NOT EXISTS idx_chapters_deleted_at   ON chapters(deleted_at);
CREATE INDEX IF NOT EXISTS idx_media_assets_deleted_at ON media_assets(deleted_at);
