package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Storage абстрагирует хранение исходных документов
// в MVP — локальный диск, позже можно добавить S3 без правок вызывающих
type Storage interface {
	// Save сохраняет файл пользователя и возвращает относительный путь
	Save(userID int, filename string, r io.Reader) (string, error)
	// Open возвращает reader для ранее сохранённого пути
	Open(path string) (io.ReadCloser, error)
	// Delete удаляет путь, отсутствие файла не ошибка
	Delete(path string) error
}

// Local хранит файлы в <root>/<userID>/
type Local struct {
	root string
}

// NewLocal создаёт root при необходимости и возвращает локальное хранилище
func NewLocal(root string) (*Local, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create upload root: %w", err)
	}
	return &Local{root: root}, nil
}

func sanitize(name string) string {
	base := filepath.Base(name)
	base = strings.ReplaceAll(base, string(os.PathSeparator), "_")
	if base == "." || base == ".." || base == "" {
		base = "document"
	}
	return base
}

// Save записывает файл и возвращает его относительный путь
func (l *Local) Save(userID int, filename string, r io.Reader) (string, error) {
	dir := filepath.Join(l.root, fmt.Sprintf("%d", userID))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("create user dir: %w", err)
	}

	unique := fmt.Sprintf("%d_%s", time.Now().UnixNano(), sanitize(filename))
	full := filepath.Join(dir, unique)

	f, err := os.Create(full)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		_ = os.Remove(full)
		return "", fmt.Errorf("write file: %w", err)
	}

	rel, err := filepath.Rel(l.root, full)
	if err != nil {
		return "", fmt.Errorf("relative path: %w", err)
	}
	return rel, nil
}

// Open открывает сохранённый файл, защищаясь от обхода пути
func (l *Local) Open(path string) (io.ReadCloser, error) {
	full := filepath.Join(l.root, filepath.Clean("/"+path))
	rel, err := filepath.Rel(l.root, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		return nil, fmt.Errorf("invalid path")
	}
	return os.Open(full)
}

// Delete удаляет сохранённый файл
func (l *Local) Delete(path string) error {
	full := filepath.Join(l.root, filepath.Clean("/"+path))
	rel, err := filepath.Rel(l.root, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("invalid path")
	}
	if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
