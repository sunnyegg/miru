package storage

import (
	"database/sql"
	"errors"
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
