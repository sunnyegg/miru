package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const currentVersion = 5

var ErrNotFound = errors.New("not found")

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	store := &Store{db: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	var version int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return err
	}
	if version >= currentVersion {
		return nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(schemaV1); err != nil {
		return err
	}
	if version == 1 {
		if _, err := tx.Exec(schemaV2); err != nil {
			return err
		}
	}
	if version < 3 {
		if _, err := tx.Exec(schemaV3); err != nil {
			return err
		}
	}
	if version < 4 {
		if _, err := tx.Exec(schemaV4); err != nil {
			return err
		}
	}
	if version < 5 {
		if _, err := tx.Exec(schemaV5); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", currentVersion)); err != nil {
		return err
	}
	return tx.Commit()
}

const schemaV1 = `
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS anime_cache (
    anilist_id INTEGER PRIMARY KEY,
    title_romaji TEXT NOT NULL,
    title_english TEXT,
    cover_image TEXT,
    total_episodes INTEGER,
    status TEXT,
    synopsis TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS episode_downloads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    anilist_id INTEGER,
    episode_number INTEGER,
    file_path TEXT NOT NULL UNIQUE,
    display_title TEXT,
    downloaded_bytes INTEGER DEFAULT 0,
    status TEXT CHECK(status IN ('DOWNLOADING', 'COMPLETED', 'FAILED', 'PAUSED')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (anilist_id) REFERENCES anime_cache(anilist_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS episode_unique
    ON episode_downloads(anilist_id, episode_number)
    WHERE anilist_id IS NOT NULL AND episode_number IS NOT NULL;

CREATE TABLE IF NOT EXISTS torrent_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    info_hash TEXT,
    source TEXT NOT NULL,
    dest_dir TEXT NOT NULL,
    name TEXT,
    status TEXT NOT NULL CHECK(status IN ('DOWNLOADING', 'PAUSED', 'SEEDING', 'COMPLETED', 'FAILED', 'CANCELLED')),
    bytes_completed INTEGER DEFAULT 0,
    bytes_total INTEGER DEFAULT 0,
    bytes_uploaded INTEGER DEFAULT 0,
    error TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS sync_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    anilist_id INTEGER NOT NULL,
    episode_number INTEGER NOT NULL,
    synced_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(anilist_id, episode_number)
);
`

const schemaV2 = `
ALTER TABLE torrent_jobs RENAME TO torrent_jobs_v1;

CREATE TABLE torrent_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    info_hash TEXT,
    source TEXT NOT NULL,
    dest_dir TEXT NOT NULL,
    name TEXT,
    status TEXT NOT NULL CHECK(status IN ('DOWNLOADING', 'PAUSED', 'SEEDING', 'COMPLETED', 'FAILED', 'CANCELLED')),
    bytes_completed INTEGER DEFAULT 0,
    bytes_total INTEGER DEFAULT 0,
    bytes_uploaded INTEGER DEFAULT 0,
    error TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO torrent_jobs(
    id, info_hash, source, dest_dir, name, status, bytes_completed, bytes_total, bytes_uploaded, error, created_at, updated_at
)
SELECT
    id, info_hash, source, dest_dir, name, status, bytes_completed, bytes_total, 0, error, created_at, updated_at
FROM torrent_jobs_v1;

DROP TABLE torrent_jobs_v1;
`

const schemaV3 = `
CREATE TABLE IF NOT EXISTS api_cache (
    cache_key TEXT PRIMARY KEY,
    payload TEXT NOT NULL,
    fetched_at INTEGER NOT NULL
);
`

const schemaV4 = `
ALTER TABLE episode_downloads ADD COLUMN resume_position REAL NOT NULL DEFAULT 0;
`

const schemaV5 = `
ALTER TABLE torrent_jobs RENAME TO torrent_jobs_v4;

CREATE TABLE torrent_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    info_hash TEXT,
    source TEXT NOT NULL,
    dest_dir TEXT NOT NULL,
    name TEXT,
    status TEXT NOT NULL CHECK(status IN ('QUEUED', 'DOWNLOADING', 'PAUSED', 'SEEDING', 'COMPLETED', 'FAILED', 'CANCELLED')),
    bytes_completed INTEGER DEFAULT 0,
    bytes_total INTEGER DEFAULT 0,
    bytes_uploaded INTEGER DEFAULT 0,
    error TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO torrent_jobs(
    id, info_hash, source, dest_dir, name, status, bytes_completed, bytes_total, bytes_uploaded, error, created_at, updated_at
)
SELECT
    id, info_hash, source, dest_dir, name, status, bytes_completed, bytes_total, bytes_uploaded, error, created_at, updated_at
FROM torrent_jobs_v4;

DROP TABLE torrent_jobs_v4;
`

func (s *Store) GetSetting(key string) (string, error) {
	var value string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return value, err
}

func (s *Store) SetSetting(key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO settings(key, value) VALUES(?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	return err
}

func (s *Store) GetAPICache(key string, maxAge time.Duration) (string, error) {
	var payload string
	var fetchedAt int64
	err := s.db.QueryRow(
		`SELECT payload, fetched_at FROM api_cache WHERE cache_key = ?`,
		key,
	).Scan(&payload, &fetchedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if maxAge > 0 && time.Since(time.Unix(fetchedAt, 0)) > maxAge {
		return "", ErrNotFound
	}
	return payload, nil
}

func (s *Store) SetAPICache(key, payload string) error {
	_, err := s.db.Exec(
		`INSERT INTO api_cache(cache_key, payload, fetched_at) VALUES(?, ?, ?)
		 ON CONFLICT(cache_key) DO UPDATE SET
		 	payload = excluded.payload,
		 	fetched_at = excluded.fetched_at`,
		key, payload, time.Now().Unix(),
	)
	return err
}

func (s *Store) DeleteAPICache(key string) error {
	_, err := s.db.Exec(`DELETE FROM api_cache WHERE cache_key = ?`, key)
	return err
}
