package torrentx

import (
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type FileView struct {
	Path           string `json:"path"`
	Length         int64  `json:"length"`
	BytesCompleted int64  `json:"bytesCompleted"`
	Selected       bool   `json:"selected"`
	IsVideo        bool   `json:"isVideo"`
}

type ContentsView struct {
	Name       string     `json:"name"`
	BytesTotal int64      `json:"bytesTotal"`
	Files      []FileView `json:"files"`
}

func isVideoPath(path string) bool {
	return videoExt[strings.ToLower(filepath.Ext(path))]
}

func contentsFromInfo(info *metainfo.Info) ContentsView {
	infoFiles := info.UpvertedFiles()
	files := make([]FileView, 0, len(infoFiles))
	for _, infoFile := range infoFiles {
		path := infoFile.DisplayPath(info)
		files = append(files, FileView{
			Path:     path,
			Length:   infoFile.Length,
			Selected: true,
			IsVideo:  isVideoPath(path),
		})
	}
	return ContentsView{
		Name:       info.BestName(),
		BytesTotal: info.TotalLength(),
		Files:      files,
	}
}

func contentsFromTorrent(t *torrent.Torrent) ContentsView {
	info := t.Info()
	if info == nil {
		return ContentsView{Name: t.Name()}
	}
	return contentsFromInfo(info)
}

func encodeFiles(files []FileView) string {
	if len(files) == 0 {
		return ""
	}
	raw, err := json.Marshal(files)
	if err != nil {
		return ""
	}
	return string(raw)
}

func decodeFiles(raw string) []FileView {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var files []FileView
	if err := json.Unmarshal([]byte(raw), &files); err != nil {
		return nil
	}
	for index := range files {
		files[index].IsVideo = isVideoPath(files[index].Path)
		files[index].Selected = true
	}
	return files
}

func selectedBytesTotal(files []FileView) int64 {
	var total int64
	for _, file := range files {
		total += file.Length
	}
	return total
}

func selectedPathSet(files []FileView) map[string]struct{} {
	if len(files) == 0 {
		return nil
	}
	wanted := make(map[string]struct{}, len(files))
	for _, file := range files {
		path := strings.TrimSpace(file.Path)
		if path == "" {
			continue
		}
		wanted[path] = struct{}{}
	}
	if len(wanted) == 0 {
		return nil
	}
	return wanted
}

func fileWanted(torrentFile *torrent.File, wanted map[string]struct{}) bool {
	if len(wanted) == 0 {
		return true
	}
	if _, ok := wanted[torrentFile.DisplayPath()]; ok {
		return true
	}
	_, ok := wanted[torrentFile.Path()]
	return ok
}

func applyFileSelection(t *torrent.Torrent, files []FileView) (int64, error) {
	if t.Info() == nil {
		return 0, errors.New("torrent metadata is unavailable")
	}
	wanted := selectedPathSet(files)
	var total int64
	matched := 0
	for _, torrentFile := range t.Files() {
		if !fileWanted(torrentFile, wanted) {
			torrentFile.SetPriority(torrent.PiecePriorityNone)
			continue
		}
		torrentFile.Download()
		total += torrentFile.Length()
		matched++
	}
	if len(wanted) > 0 && matched == 0 {
		return 0, errors.New("none of the selected files are in this torrent")
	}
	return total, nil
}

func fileViewsFromTorrent(t *torrent.Torrent, selected []FileView) []FileView {
	if t.Info() == nil {
		return selected
	}
	wanted := selectedPathSet(selected)
	files := make([]FileView, 0, len(t.Files()))
	for _, torrentFile := range t.Files() {
		if !fileWanted(torrentFile, wanted) {
			continue
		}
		path := torrentFile.DisplayPath()
		files = append(files, FileView{
			Path:           path,
			Length:         torrentFile.Length(),
			BytesCompleted: torrentFile.BytesCompleted(),
			Selected:       true,
			IsVideo:        isVideoPath(path),
		})
	}
	return files
}
