package datareset

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/sunnyegg/miru/internal/paths"
)

const (
	markerSchemaVersion = 1
	markerFilename      = ".miru-reset.json"
	maxMarkerBytes      = 64 * 1024

	phasePending        = "pending"
	phaseStaged         = "staged"
	phaseCleanupPending = "cleanup_pending"
)

type StartupState struct {
	NeedsCommit    bool
	CleanupPending bool
	Blocked        bool
}

type Usage struct {
	Bytes          int64
	CleanupPending bool
}

type marker struct {
	SchemaVersion int      `json:"schemaVersion"`
	ResetID       string   `json:"resetId"`
	Phase         string   `json:"phase"`
	StagedRoots   []string `json:"stagedRoots"`
}

type resetRoot struct {
	name string
	path string
}

func Schedule(dirs paths.Dirs) error {
	roots, err := rootsFor(dirs)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(markerPath(dirs)); err == nil {
		return errors.New("a data reset is already pending")
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("check data reset marker: %w", err)
	}

	resetID, err := newResetID()
	if err != nil {
		return fmt.Errorf("create data reset ID: %w", err)
	}
	for _, root := range roots {
		if err := validateExistingRoot(root); err != nil {
			return err
		}
	}
	return writeMarker(dirs, marker{
		SchemaVersion: markerSchemaVersion,
		ResetID:       resetID,
		Phase:         phasePending,
	})
}

func Prepare(dirs paths.Dirs) (StartupState, error) {
	roots, err := rootsFor(dirs)
	if err != nil {
		return StartupState{Blocked: true}, err
	}
	current, err := readMarker(dirs, roots)
	if errors.Is(err, os.ErrNotExist) {
		return StartupState{}, nil
	}
	if err != nil {
		return StartupState{Blocked: true}, err
	}

	switch current.Phase {
	case phaseCleanupPending:
		if err := cleanup(dirs, roots, &current); err != nil {
			return StartupState{CleanupPending: true}, err
		}
		return StartupState{}, nil
	case phaseStaged:
		if err := ensureActiveRoots(roots); err != nil {
			return StartupState{Blocked: true}, err
		}
		return StartupState{NeedsCommit: true}, nil
	case phasePending:
		if err := stage(dirs, roots, &current); err != nil {
			rollbackErr := rollback(roots, current)
			if rollbackErr != nil {
				return StartupState{Blocked: true}, errors.Join(err, rollbackErr)
			}
			if removeErr := removeMarker(dirs); removeErr != nil {
				return StartupState{Blocked: true}, errors.Join(err, removeErr)
			}
			return StartupState{}, err
		}
		return StartupState{NeedsCommit: true}, nil
	default:
		return StartupState{Blocked: true}, fmt.Errorf("unsupported data reset phase %q", current.Phase)
	}
}

func Commit(dirs paths.Dirs) error {
	roots, err := rootsFor(dirs)
	if err != nil {
		return err
	}
	current, err := readMarker(dirs, roots)
	if err != nil {
		return err
	}
	if current.Phase != phaseStaged && current.Phase != phaseCleanupPending {
		return fmt.Errorf("cannot commit data reset in phase %q", current.Phase)
	}
	if current.Phase != phaseCleanupPending {
		current.Phase = phaseCleanupPending
		if err := writeMarker(dirs, current); err != nil {
			return err
		}
	}
	return cleanup(dirs, roots, &current)
}

func Measure(dirs paths.Dirs) (Usage, error) {
	roots, err := rootsFor(dirs)
	if err != nil {
		return Usage{}, err
	}

	measurePaths := make([]string, 0, len(roots)*2)
	for _, root := range roots {
		measurePaths = append(measurePaths, root.path)
	}
	current, markerErr := readMarker(dirs, roots)
	if markerErr != nil && !errors.Is(markerErr, os.ErrNotExist) {
		return Usage{}, markerErr
	}
	if markerErr == nil {
		for _, root := range roots {
			if slices.Contains(current.StagedRoots, root.name) {
				measurePaths = append(measurePaths, backupPath(root, current.ResetID))
			}
		}
	}

	var total int64
	for _, path := range measurePaths {
		size, err := pathSize(path)
		if err != nil {
			return Usage{}, err
		}
		total += size
	}
	return Usage{
		Bytes:          total,
		CleanupPending: markerErr == nil && current.Phase == phaseCleanupPending,
	}, nil
}

func stage(dirs paths.Dirs, roots []resetRoot, current *marker) error {
	for _, root := range roots {
		backup := backupPath(root, current.ResetID)
		activeInfo, activeErr := os.Lstat(root.path)
		backupInfo, backupErr := os.Lstat(backup)
		activeExists := activeErr == nil
		backupExists := backupErr == nil
		if activeErr != nil && !errors.Is(activeErr, os.ErrNotExist) {
			return fmt.Errorf("inspect %s data root: %w", root.name, activeErr)
		}
		if backupErr != nil && !errors.Is(backupErr, os.ErrNotExist) {
			return fmt.Errorf("inspect %s data backup: %w", root.name, backupErr)
		}
		if activeExists && activeInfo.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s data root is a symbolic link", root.name)
		}
		if backupExists && backupInfo.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("%s data backup is a symbolic link", root.name)
		}
		if backupExists && !backupInfo.IsDir() {
			return fmt.Errorf("%s data backup is not a directory", root.name)
		}

		recorded := slices.Contains(current.StagedRoots, root.name)
		if backupExists {
			if err := resumeStagedRoot(dirs, root, activeExists, recorded, current); err != nil {
				return err
			}
			continue
		}
		if err := stageActiveRoot(dirs, root, backup, activeInfo, activeExists, recorded, current); err != nil {
			return err
		}
	}

	current.Phase = phaseStaged
	if err := writeMarker(dirs, *current); err != nil {
		return err
	}
	return ensureActiveRoots(roots)
}

func stageActiveRoot(
	dirs paths.Dirs,
	root resetRoot,
	backup string,
	activeInfo fs.FileInfo,
	activeExists bool,
	recorded bool,
	current *marker,
) error {
	if recorded {
		return fmt.Errorf("staged %s data backup is missing", root.name)
	}
	if !activeExists {
		return fmt.Errorf("active %s data root is missing", root.name)
	}
	if !activeInfo.IsDir() {
		return fmt.Errorf("%s data root is not a directory", root.name)
	}
	if err := os.Rename(root.path, backup); err != nil {
		return fmt.Errorf("stage %s data: %w", root.name, err)
	}
	current.StagedRoots = append(current.StagedRoots, root.name)
	return writeMarker(dirs, *current)
}

func resumeStagedRoot(
	dirs paths.Dirs,
	root resetRoot,
	activeExists bool,
	recorded bool,
	current *marker,
) error {
	if activeExists {
		empty, err := emptyDirectory(root.path)
		if err != nil {
			return fmt.Errorf("inspect recreated %s data root: %w", root.name, err)
		}
		if !empty {
			return fmt.Errorf("both active and staged %s data contain files", root.name)
		}
		if err := os.Remove(root.path); err != nil {
			return fmt.Errorf("remove recreated %s data root: %w", root.name, err)
		}
	}
	if recorded {
		return nil
	}
	current.StagedRoots = append(current.StagedRoots, root.name)
	return writeMarker(dirs, *current)
}

func rollback(roots []resetRoot, current marker) error {
	var rollbackErr error
	for index := len(roots) - 1; index >= 0; index-- {
		root := roots[index]
		if !slices.Contains(current.StagedRoots, root.name) {
			continue
		}
		backup := backupPath(root, current.ResetID)
		if _, err := os.Lstat(backup); errors.Is(err, os.ErrNotExist) {
			rollbackErr = errors.Join(rollbackErr, fmt.Errorf("cannot restore missing %s data backup", root.name))
			continue
		} else if err != nil {
			rollbackErr = errors.Join(rollbackErr, fmt.Errorf("inspect %s rollback data: %w", root.name, err))
			continue
		}
		if _, err := os.Lstat(root.path); err == nil {
			empty, emptyErr := emptyDirectory(root.path)
			if emptyErr != nil || !empty {
				rollbackErr = errors.Join(rollbackErr, fmt.Errorf("cannot restore %s data over a non-empty root", root.name))
				continue
			}
			if removeErr := os.Remove(root.path); removeErr != nil {
				rollbackErr = errors.Join(rollbackErr, fmt.Errorf("remove empty %s data root: %w", root.name, removeErr))
				continue
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			rollbackErr = errors.Join(rollbackErr, fmt.Errorf("inspect active %s data: %w", root.name, err))
			continue
		}
		if err := os.Rename(backup, root.path); err != nil {
			rollbackErr = errors.Join(rollbackErr, fmt.Errorf("restore %s data: %w", root.name, err))
		}
	}
	return rollbackErr
}

func cleanup(dirs paths.Dirs, roots []resetRoot, current *marker) error {
	for _, root := range roots {
		if !slices.Contains(current.StagedRoots, root.name) {
			continue
		}
		backup := backupPath(root, current.ResetID)
		info, err := os.Lstat(backup)
		if errors.Is(err, os.ErrNotExist) {
			current.StagedRoots = slices.DeleteFunc(current.StagedRoots, func(name string) bool {
				return name == root.name
			})
			if err := writeMarker(dirs, *current); err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return fmt.Errorf("inspect staged %s data: %w", root.name, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("staged %s data is a symbolic link", root.name)
		}
		if !info.IsDir() {
			return fmt.Errorf("staged %s data is not a directory", root.name)
		}
		if err := os.RemoveAll(backup); err != nil {
			return fmt.Errorf("remove staged %s data: %w", root.name, err)
		}
		current.StagedRoots = slices.DeleteFunc(current.StagedRoots, func(name string) bool {
			return name == root.name
		})
		if err := writeMarker(dirs, *current); err != nil {
			return err
		}
	}
	return removeMarker(dirs)
}

func rootsFor(dirs paths.Dirs) ([]resetRoot, error) {
	candidates := []resetRoot{
		{name: "config", path: dirs.Config},
		{name: "cache", path: dirs.Cache},
		{name: "data", path: dirs.Data},
	}
	roots := make([]resetRoot, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		clean, err := validateRootPath(candidate.path)
		if err != nil {
			return nil, fmt.Errorf("invalid %s data root: %w", candidate.name, err)
		}
		candidate.path = clean
		key := clean
		if runtime.GOOS == "windows" {
			key = strings.ToLower(key)
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		roots = append(roots, candidate)
	}
	return roots, nil
}

func validateRootPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("path is empty")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	clean := filepath.Clean(abs)
	if filepath.Base(clean) != "miru" {
		return "", errors.New("path base is not miru")
	}
	parent := filepath.Dir(clean)
	if parent == clean || parent == filepath.VolumeName(clean)+string(os.PathSeparator) {
		return "", errors.New("path is directly below a filesystem root")
	}
	if home, err := os.UserHomeDir(); err == nil && samePath(clean, home) {
		return "", errors.New("path is the user home directory")
	}
	return clean, nil
}

func validateExistingRoot(root resetRoot) error {
	info, err := os.Lstat(root.path)
	if err != nil {
		return fmt.Errorf("inspect %s data root: %w", root.name, err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s data root is a symbolic link", root.name)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s data root is not a directory", root.name)
	}
	return nil
}

func ensureActiveRoots(roots []resetRoot) error {
	for _, root := range roots {
		if info, err := os.Lstat(root.path); err == nil {
			if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
				return fmt.Errorf("fresh %s data root is not a directory", root.name)
			}
			continue
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspect fresh %s data root: %w", root.name, err)
		}
		if err := os.MkdirAll(root.path, 0o700); err != nil {
			return fmt.Errorf("create fresh %s data root: %w", root.name, err)
		}
	}
	return nil
}

func readMarker(dirs paths.Dirs, roots []resetRoot) (marker, error) {
	path := markerPath(dirs)
	info, err := os.Lstat(path)
	if err != nil {
		return marker{}, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return marker{}, errors.New("data reset marker is not a regular file")
	}
	if info.Size() > maxMarkerBytes {
		return marker{}, errors.New("data reset marker is too large")
	}
	file, err := os.Open(path)
	if err != nil {
		return marker{}, err
	}
	defer file.Close()

	decoder := json.NewDecoder(io.LimitReader(file, maxMarkerBytes))
	decoder.DisallowUnknownFields()
	var current marker
	if err := decoder.Decode(&current); err != nil {
		return marker{}, fmt.Errorf("read data reset marker: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return marker{}, errors.New("data reset marker has trailing content")
	}
	if current.SchemaVersion != markerSchemaVersion {
		return marker{}, fmt.Errorf("unsupported data reset marker version %d", current.SchemaVersion)
	}
	if len(current.ResetID) != 32 {
		return marker{}, errors.New("invalid data reset ID")
	}
	if _, err := hex.DecodeString(current.ResetID); err != nil {
		return marker{}, errors.New("invalid data reset ID")
	}
	if current.Phase != phasePending && current.Phase != phaseStaged && current.Phase != phaseCleanupPending {
		return marker{}, fmt.Errorf("unsupported data reset phase %q", current.Phase)
	}
	validNames := make(map[string]struct{}, len(roots))
	for _, root := range roots {
		validNames[root.name] = struct{}{}
	}
	seen := make(map[string]struct{}, len(current.StagedRoots))
	for _, name := range current.StagedRoots {
		if _, ok := validNames[name]; !ok {
			return marker{}, fmt.Errorf("unknown staged data root %q", name)
		}
		if _, duplicate := seen[name]; duplicate {
			return marker{}, fmt.Errorf("duplicate staged data root %q", name)
		}
		seen[name] = struct{}{}
	}
	return current, nil
}

func writeMarker(dirs paths.Dirs, current marker) error {
	path := markerPath(dirs)
	temp := path + ".tmp"
	if info, err := os.Lstat(temp); err == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return errors.New("temporary data reset marker is not a regular file")
		}
		if err := os.Remove(temp); err != nil {
			return fmt.Errorf("remove stale data reset marker: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect temporary data reset marker: %w", err)
	}
	file, err := os.OpenFile(temp, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("create data reset marker: %w", err)
	}
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(temp)
		}
	}()

	encoder := json.NewEncoder(file)
	if err := encoder.Encode(current); err != nil {
		_ = file.Close()
		return fmt.Errorf("encode data reset marker: %w", err)
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return fmt.Errorf("sync data reset marker: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close data reset marker: %w", err)
	}
	if err := os.Rename(temp, path); err != nil {
		return fmt.Errorf("publish data reset marker: %w", err)
	}
	removeTemp = false
	return nil
}

func removeMarker(dirs paths.Dirs) error {
	if err := os.Remove(markerPath(dirs)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("remove data reset marker: %w", err)
	}
	return nil
}

func markerPath(dirs paths.Dirs) string {
	return filepath.Join(filepath.Dir(filepath.Clean(dirs.Config)), markerFilename)
}

func backupPath(root resetRoot, resetID string) string {
	return root.path + ".reset-" + resetID
}

func newResetID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func pathSize(path string) (int64, error) {
	var total int64
	err := filepath.WalkDir(path, func(_ string, entry fs.DirEntry, walkErr error) error {
		if errors.Is(walkErr, os.ErrNotExist) {
			return nil
		}
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("measure Miru data: %w", err)
	}
	return total, nil
}

func emptyDirectory(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return false, err
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return false, errors.New("path is not a directory")
	}
	directory, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer directory.Close()
	_, err = directory.Readdirnames(1)
	if errors.Is(err, io.EOF) {
		return true, nil
	}
	return false, err
}

func samePath(first, second string) bool {
	first = filepath.Clean(first)
	second = filepath.Clean(second)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(first, second)
	}
	return first == second
}
