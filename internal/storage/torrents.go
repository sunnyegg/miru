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
	FilesJSON      string
}

func (s *Store) InsertTorrentJob(job TorrentJob) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(
		`INSERT INTO torrent_jobs(info_hash, source, dest_dir, name, status, bytes_completed, bytes_total, bytes_uploaded, error, files_json, created_at, updated_at)
		 VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.InfoHash, job.Source, job.DestDir, job.Name, job.Status, job.BytesCompleted, job.BytesTotal, job.BytesUploaded, job.Error, job.FilesJSON, now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UpdateTorrentJob(job TorrentJob) error {
	_, err := s.db.Exec(
		`UPDATE torrent_jobs
		 SET info_hash = ?, name = ?, status = ?, bytes_completed = ?, bytes_total = ?, bytes_uploaded = ?, error = ?, files_json = ?, updated_at = ?
		 WHERE id = ?`,
		job.InfoHash, job.Name, job.Status, job.BytesCompleted, job.BytesTotal, job.BytesUploaded, job.Error, job.FilesJSON,
		time.Now().UTC().Format(time.RFC3339), job.ID,
	)
	return err
}

func scanTorrentJob(scanner interface {
	Scan(dest ...any) error
}) (TorrentJob, error) {
	var job TorrentJob
	err := scanner.Scan(
		&job.ID, &job.InfoHash, &job.Source, &job.DestDir, &job.Name, &job.Status,
		&job.BytesCompleted, &job.BytesTotal, &job.BytesUploaded, &job.Error, &job.FilesJSON,
	)
	return job, err
}

func (s *Store) TorrentJobByID(id int64) (TorrentJob, error) {
	row := s.db.QueryRow(
		`SELECT id, COALESCE(info_hash, ''), source, dest_dir, COALESCE(name, ''), status,
		        bytes_completed, bytes_total, bytes_uploaded, COALESCE(error, ''), COALESCE(files_json, '')
		 FROM torrent_jobs
		 WHERE id = ?`,
		id,
	)
	job, err := scanTorrentJob(row)
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
		        bytes_completed, bytes_total, bytes_uploaded, COALESCE(error, ''), COALESCE(files_json, '')
		 FROM torrent_jobs
		 ORDER BY id DESC
		 LIMIT 1`,
	)
	job, err := scanTorrentJob(row)
	if errors.Is(err, sql.ErrNoRows) {
		return TorrentJob{}, ErrNotFound
	}
	return job, err
}

func (s *Store) ListTorrentJobs() ([]TorrentJob, error) {
	rows, err := s.db.Query(
		`SELECT id, COALESCE(info_hash, ''), source, dest_dir, COALESCE(name, ''), status,
		        bytes_completed, bytes_total, bytes_uploaded, COALESCE(error, ''), COALESCE(files_json, '')
		 FROM torrent_jobs
		 ORDER BY created_at DESC, id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []TorrentJob
	for rows.Next() {
		job, err := scanTorrentJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if jobs == nil {
		jobs = []TorrentJob{}
	}
	return jobs, rows.Err()
}

func (s *Store) NextQueuedTorrentJob() (TorrentJob, error) {
	row := s.db.QueryRow(
		`SELECT id, COALESCE(info_hash, ''), source, dest_dir, COALESCE(name, ''), status,
		        bytes_completed, bytes_total, bytes_uploaded, COALESCE(error, ''), COALESCE(files_json, '')
		 FROM torrent_jobs
		 WHERE status = 'QUEUED'
		 ORDER BY id ASC
		 LIMIT 1`,
	)
	job, err := scanTorrentJob(row)
	if errors.Is(err, sql.ErrNoRows) {
		return TorrentJob{}, ErrNotFound
	}
	return job, err
}

func (s *Store) CountDownloadingTorrentJobs() (int, error) {
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(1) FROM torrent_jobs WHERE status = 'DOWNLOADING'`,
	).Scan(&count)
	return count, err
}

func (s *Store) RecoverInterruptedDownloads() error {
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := s.db.Exec(
		`UPDATE torrent_jobs
		 SET status = 'QUEUED', error = '', updated_at = ?
		 WHERE status = 'DOWNLOADING'`,
		now,
	); err != nil {
		return err
	}
	_, err := s.db.Exec(
		`UPDATE torrent_jobs
		 SET status = 'FAILED', error = 'interrupted by restart', updated_at = ?
		 WHERE status = 'PAUSED'`,
		now,
	)
	return err
}
