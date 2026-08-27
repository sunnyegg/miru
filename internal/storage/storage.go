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

const currentVersion = 3

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

type Anime struct {
	AnilistID     int
	TitleRomaji   string
	TitleEnglish  string
	CoverImage    string
	TotalEpisodes int
	Status        string
	Synopsis      string
}

func (s *Store) UpsertAnime(a Anime) error {
	_, err := s.db.Exec(
		`INSERT INTO anime_cache(
			anilist_id, title_romaji, title_english, cover_image, total_episodes, status, synopsis, updated_at
		) VALUES(?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(anilist_id) DO UPDATE SET
			title_romaji = excluded.title_romaji,
			title_english = excluded.title_english,
			cover_image = excluded.cover_image,
			total_episodes = excluded.total_episodes,
			status = excluded.status,
			synopsis = excluded.synopsis,
			updated_at = excluded.updated_at`,
		a.AnilistID, a.TitleRomaji, a.TitleEnglish, a.CoverImage, a.TotalEpisodes, a.Status, a.Synopsis,
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

type Episode struct {
	ID              int64
	AnilistID       sql.NullInt64
	EpisodeNumber   sql.NullInt64
	FilePath        string
	DisplayTitle    string
	DownloadedBytes int64
	Status          string
	TitleRomaji     string
	TitleEnglish    string
	CoverImage      string
}

func (s *Store) InsertEpisode(e Episode) (int64, error) {
	res, err := s.db.Exec(
		`INSERT INTO episode_downloads(
			anilist_id, episode_number, file_path, display_title, downloaded_bytes, status
		) VALUES(?, ?, ?, ?, ?, ?)`,
		nullInt(e.AnilistID), nullInt(e.EpisodeNumber), e.FilePath, e.DisplayTitle, e.DownloadedBytes, e.Status,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) GetEpisode(id int64) (Episode, error) {
	row := s.db.QueryRow(
		`SELECT e.id, e.anilist_id, e.episode_number, e.file_path, e.display_title,
		        e.downloaded_bytes, e.status,
		        COALESCE(a.title_romaji, ''), COALESCE(a.title_english, ''), COALESCE(a.cover_image, '')
		 FROM episode_downloads e
		 LEFT JOIN anime_cache a ON a.anilist_id = e.anilist_id
		 WHERE e.id = ?`,
		id,
	)
	var e Episode
	err := row.Scan(
		&e.ID, &e.AnilistID, &e.EpisodeNumber, &e.FilePath, &e.DisplayTitle,
		&e.DownloadedBytes, &e.Status, &e.TitleRomaji, &e.TitleEnglish, &e.CoverImage,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Episode{}, ErrNotFound
	}
	return e, err
}

func (s *Store) EpisodeByPath(path string) (Episode, error) {
	row := s.db.QueryRow(
		`SELECT e.id, e.anilist_id, e.episode_number, e.file_path, e.display_title,
		        e.downloaded_bytes, e.status,
		        COALESCE(a.title_romaji, ''), COALESCE(a.title_english, ''), COALESCE(a.cover_image, '')
		 FROM episode_downloads e
		 LEFT JOIN anime_cache a ON a.anilist_id = e.anilist_id
		 WHERE e.file_path = ?`,
		path,
	)
	var e Episode
	err := row.Scan(
		&e.ID, &e.AnilistID, &e.EpisodeNumber, &e.FilePath, &e.DisplayTitle,
		&e.DownloadedBytes, &e.Status, &e.TitleRomaji, &e.TitleEnglish, &e.CoverImage,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Episode{}, ErrNotFound
	}
	return e, err
}

func (s *Store) ListEpisodes() ([]Episode, error) {
	rows, err := s.db.Query(
		`SELECT e.id, e.anilist_id, e.episode_number, e.file_path, e.display_title,
		        e.downloaded_bytes, e.status,
		        COALESCE(a.title_romaji, ''), COALESCE(a.title_english, ''), COALESCE(a.cover_image, '')
		 FROM episode_downloads e
		 LEFT JOIN anime_cache a ON a.anilist_id = e.anilist_id
		 ORDER BY e.created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Episode
	for rows.Next() {
		var e Episode
		if err := rows.Scan(
			&e.ID, &e.AnilistID, &e.EpisodeNumber, &e.FilePath, &e.DisplayTitle,
			&e.DownloadedBytes, &e.Status, &e.TitleRomaji, &e.TitleEnglish, &e.CoverImage,
		); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	if out == nil {
		out = []Episode{}
	}
	return out, rows.Err()
}

func (s *Store) BindEpisode(id int64, anilistID int, episodeNumber int) error {
	_, err := s.db.Exec(
		`UPDATE episode_downloads SET anilist_id = ?, episode_number = ? WHERE id = ?`,
		anilistID, episodeNumber, id,
	)
	return err
}

func (s *Store) HasSynced(anilistID int, episodeNumber int) (bool, error) {
	var n int
	err := s.db.QueryRow(
		`SELECT COUNT(1) FROM sync_events WHERE anilist_id = ? AND episode_number = ?`,
		anilistID, episodeNumber,
	).Scan(&n)
	return n > 0, err
}

func (s *Store) RecordSync(anilistID int, episodeNumber int) error {
	_, err := s.db.Exec(
		`INSERT INTO sync_events(anilist_id, episode_number) VALUES(?, ?)
		 ON CONFLICT(anilist_id, episode_number) DO NOTHING`,
		anilistID, episodeNumber,
	)
	return err
}

type TorrentJob struct {
	ID             int64
	InfoHash       string
	Source         string
	DestDir        string
	Name           string
	Status         string
	BytesCompleted int64
	BytesTotal     int64
	BytesUploaded  int64
	Error          string
}

func (s *Store) InsertTorrentJob(job TorrentJob) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(
		`INSERT INTO torrent_jobs(info_hash, source, dest_dir, name, status, bytes_completed, bytes_total, bytes_uploaded, error, created_at, updated_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.InfoHash, job.Source, job.DestDir, job.Name, job.Status, job.BytesCompleted, job.BytesTotal, job.BytesUploaded, job.Error, now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UpdateTorrentJob(job TorrentJob) error {
	_, err := s.db.Exec(
		`UPDATE torrent_jobs
		 SET info_hash = ?, name = ?, status = ?, bytes_completed = ?, bytes_total = ?, bytes_uploaded = ?, error = ?, updated_at = ?
		 WHERE id = ?`,
		job.InfoHash, job.Name, job.Status, job.BytesCompleted, job.BytesTotal, job.BytesUploaded, job.Error,
		time.Now().UTC().Format(time.RFC3339), job.ID,
	)
	return err
}

func (s *Store) LatestTorrentJob() (TorrentJob, error) {
	row := s.db.QueryRow(
		`SELECT id, COALESCE(info_hash, ''), source, dest_dir, COALESCE(name, ''), status,
		        bytes_completed, bytes_total, bytes_uploaded, COALESCE(error, '')
		 FROM torrent_jobs
		 ORDER BY id DESC
		 LIMIT 1`,
	)
	var job TorrentJob
	err := row.Scan(
		&job.ID, &job.InfoHash, &job.Source, &job.DestDir, &job.Name, &job.Status,
		&job.BytesCompleted, &job.BytesTotal, &job.BytesUploaded, &job.Error,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return TorrentJob{}, ErrNotFound
	}
	return job, err
}

func (s *Store) ListTorrentJobs() ([]TorrentJob, error) {
	rows, err := s.db.Query(
		`SELECT id, COALESCE(info_hash, ''), source, dest_dir, COALESCE(name, ''), status,
		        bytes_completed, bytes_total, bytes_uploaded, COALESCE(error, '')
		 FROM torrent_jobs
		 ORDER BY created_at DESC, id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []TorrentJob
	for rows.Next() {
		var job TorrentJob
		if err := rows.Scan(
			&job.ID, &job.InfoHash, &job.Source, &job.DestDir, &job.Name, &job.Status,
			&job.BytesCompleted, &job.BytesTotal, &job.BytesUploaded, &job.Error,
		); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if jobs == nil {
		jobs = []TorrentJob{}
	}
	return jobs, rows.Err()
}

func (s *Store) FailInterruptedDownloads() error {
	_, err := s.db.Exec(
		`UPDATE torrent_jobs
		 SET status = 'FAILED', error = 'interrupted by restart', updated_at = ?
		 WHERE status IN ('DOWNLOADING', 'PAUSED', 'SEEDING')`,
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

func nullInt(v sql.NullInt64) any {
	if !v.Valid {
		return nil
	}
	return v.Int64
}
