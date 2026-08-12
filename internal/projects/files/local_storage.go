package files

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	root string
}

func NewLocalStorage(root string) (*LocalStorage, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, errors.New("local project file storage path is empty")
	}
	if err := os.MkdirAll(root, 0o750); err != nil {
		return nil, err
	}
	absolute, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &LocalStorage{root: absolute}, nil
}

func (s *LocalStorage) path(key string) (string, error) {
	clean := filepath.Clean(key)
	if clean == "." || filepath.IsAbs(clean) || strings.HasPrefix(clean, ".."+string(filepath.Separator)) || clean == ".." {
		return "", errors.New("invalid project file storage key")
	}
	path := filepath.Join(s.root, clean)
	if !strings.HasPrefix(path, s.root+string(filepath.Separator)) {
		return "", errors.New("invalid project file storage key")
	}
	return path, nil
}

func (s *LocalStorage) Put(_ context.Context, key string, content []byte) error {
	path, err := s.path(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return os.WriteFile(path, content, 0o640)
}

func (s *LocalStorage) Get(_ context.Context, key string) ([]byte, error) {
	path, err := s.path(key)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path)
}

func (s *LocalStorage) Delete(_ context.Context, key string) error {
	path, err := s.path(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
