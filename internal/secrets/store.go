package secrets

import (
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/zalando/go-keyring"
)

const (
	service     = "miru"
	anilistUser = "anilist_token"
)

var ErrNotFound = errors.New("secret not found")

type Store interface {
	Get() (string, error)
	Set(value string) error
	Delete() error
}

type KeyringStore struct {
	filePath string
}

func New(fileFallback string) *KeyringStore {
	return &KeyringStore{filePath: fileFallback}
}

func (s *KeyringStore) Get() (string, error) {
	value, err := keyring.Get(service, anilistUser)
	if err == nil && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value), nil
	}
	return s.readFile()
}

func (s *KeyringStore) Set(value string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return errors.New("empty token")
	}
	if err := keyring.Set(service, anilistUser, value); err == nil {
		return nil
	}
	// ponytail: file 0600 if secret service is missing
	return os.WriteFile(s.filePath, []byte(value+"\n"), 0o600)
}

func (s *KeyringStore) Delete() error {
	_ = keyring.Delete(service, anilistUser)
	if s.filePath == "" {
		return nil
	}
	err := os.Remove(s.filePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *KeyringStore) readFile() (string, error) {
	if s.filePath == "" {
		return "", ErrNotFound
	}
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNotFound
		}
		return "", err
	}
	value := strings.TrimSpace(string(data))
	if value == "" {
		return "", ErrNotFound
	}
	return value, nil
}

type MemoryStore struct {
	mu    sync.Mutex
	value string
}

func (s *MemoryStore) Get() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.value == "" {
		return "", ErrNotFound
	}
	return s.value, nil
}

func (s *MemoryStore) Set(value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.value = strings.TrimSpace(value)
	return nil
}

func (s *MemoryStore) Delete() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.value = ""
	return nil
}
