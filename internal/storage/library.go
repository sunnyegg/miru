package storage

import (
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
	"time"
)

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
	TotalEpisodes   int
	MediaStatus     string
	ResumePosition  float64
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
		        COALESCE(a.title_romaji, ''), COALESCE(a.title_english, ''), COALESCE(a.cover_image, ''),
		        e.resume_position
		 FROM episode_downloads e
		 LEFT JOIN anime_cache a ON a.anilist_id = e.anilist_id
		 WHERE e.id = ?`,
		id,
	)
	var e Episode
	err := row.Scan(
		&e.ID, &e.AnilistID, &e.EpisodeNumber, &e.FilePath, &e.DisplayTitle,
		&e.DownloadedBytes, &e.Status, &e.TitleRomaji, &e.TitleEnglish, &e.CoverImage,
		&e.ResumePosition,
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

func (s *Store) EpisodeByDisplayTitle(title string) (Episode, error) {
	if title == "" {
		return Episode{}, ErrNotFound
	}
	row := s.db.QueryRow(
		`SELECT e.id, e.anilist_id, e.episode_number, e.file_path, e.display_title,
		        e.downloaded_bytes, e.status,
		        COALESCE(a.title_romaji, ''), COALESCE(a.title_english, ''), COALESCE(a.cover_image, '')
		 FROM episode_downloads e
		 LEFT JOIN anime_cache a ON a.anilist_id = e.anilist_id
		 WHERE e.display_title = ?
		 ORDER BY CASE WHEN e.anilist_id IS NULL THEN 1 ELSE 0 END, e.id
		 LIMIT 1`,
		title,
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

func (s *Store) HasEpisodeNumber(anilistID, episodeNumber int, excludeID int64) (bool, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(1) FROM episode_downloads
		 WHERE anilist_id = ? AND episode_number = ? AND id != ?`,
		anilistID, episodeNumber, excludeID,
	).Scan(&count)
	return count > 0, err
}

func (s *Store) HealSingleEpisodeNumbers() error {
	_, err := s.db.Exec(
		`UPDATE episode_downloads
		 SET episode_number = 1
		 WHERE anilist_id IS NOT NULL
		   AND (episode_number IS NULL OR episode_number = 0)
		   AND anilist_id IN (SELECT anilist_id FROM anime_cache WHERE total_episodes = 1)
		   AND NOT EXISTS (
		     SELECT 1 FROM episode_downloads other
		     WHERE other.anilist_id = episode_downloads.anilist_id
		       AND other.episode_number = 1
		       AND other.id != episode_downloads.id
		   )`,
	)
	return err
}

func (s *Store) ListEpisodes() ([]Episode, error) {
	if err := s.HealSingleEpisodeNumbers(); err != nil {
		return nil, err
	}
	rows, err := s.db.Query(
		`SELECT e.id, e.anilist_id, e.episode_number, e.file_path, e.display_title,
		        e.downloaded_bytes, e.status,
		        COALESCE(a.title_romaji, ''), COALESCE(a.title_english, ''), COALESCE(a.cover_image, ''),
		        COALESCE(a.total_episodes, 0), COALESCE(a.status, '')
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
			&e.TotalEpisodes, &e.MediaStatus,
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

func (s *Store) DeleteEpisodesByFilePrefix(prefix string) error {
	if prefix == "" {
		return nil
	}
	escaped := strings.ReplaceAll(prefix, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `%`, `\%`)
	escaped = strings.ReplaceAll(escaped, `_`, `\_`)
	childPattern := escaped + string(filepath.Separator) + "%"
	_, err := s.db.Exec(
		`DELETE FROM episode_downloads WHERE file_path = ? OR file_path LIKE ? ESCAPE '\'`,
		prefix,
		childPattern,
	)
	return err
}

func (s *Store) BindEpisode(id int64, anilistID int, episodeNumber int) error {
	_, err := s.db.Exec(
		`UPDATE episode_downloads SET anilist_id = ?, episode_number = ? WHERE id = ?`,
		anilistID, episodeNumber, id,
	)
	return err
}

func (s *Store) SetResumePosition(id int64, seconds float64) error {
	_, err := s.db.Exec(
		`UPDATE episode_downloads SET resume_position = ? WHERE id = ?`,
		seconds, id,
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

func nullInt(v sql.NullInt64) any {
	if !v.Valid {
		return nil
	}
	return v.Int64
}
