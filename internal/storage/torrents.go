package storage

import (
	"database/sql"
	"errors"
	"time"
)

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

func (s *Store) TorrentJobByID(id int64) (TorrentJob, error) {
	row := s.db.QueryRow(
		`SELECT id, COALESCE(info_hash, ''), source, dest_dir, COALESCE(name, ''), status,
		        bytes_completed, bytes_total, bytes_uploaded, COALESCE(error, '')
		 FROM torrent_jobs
		 WHERE id = ?`,
		id,
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

func (s *Store) DeleteTorrentJob(id int64) error {
	result, err := s.db.Exec(`DELETE FROM torrent_jobs WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
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
		 WHERE status IN ('DOWNLOADING', 'PAUSED')`,
		time.Now().UTC().Format(time.RFC3339),
	)
	return err
}
