package files

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"path/filepath"
	"strings"
	"sync"
)

var ErrFileRejected = errors.New("file rejected")

type Metadata struct {
	Key    string
	Name   string
	MIME   string
	Size   int64
	Digest string
}

type Storage interface {
	Save(context.Context, string, io.Reader) error
	Open(context.Context, string) (io.ReadCloser, error)
	Delete(context.Context, string) error
}

type Validator struct {
	MaximumSize int64
	Extensions  map[string]struct{}
	MIMEs       map[string]struct{}
}

func (v Validator) Validate(name, mime string, size int64) error {
	if size <= 0 || size > v.MaximumSize {
		return ErrFileRejected
	}
	extension := strings.ToLower(filepath.Ext(name))
	if _, ok := v.Extensions[extension]; !ok {
		return ErrFileRejected
	}
	if _, ok := v.MIMEs[mime]; !ok {
		return ErrFileRejected
	}
	return nil
}

func Digest(reader io.Reader) (string, int64, error) {
	h := sha256.New()
	size, err := io.Copy(h, reader)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), size, nil
}

type MemoryStorage struct {
	mu    sync.RWMutex
	files map[string][]byte
}

func NewMemoryStorage() *MemoryStorage { return &MemoryStorage{files: map[string][]byte{}} }

func (m *MemoryStorage) Save(ctx context.Context, key string, reader io.Reader) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.files[key] = append([]byte(nil), data...)
	m.mu.Unlock()
	return nil
}

func (m *MemoryStorage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	m.mu.RLock()
	data, ok := m.files[key]
	m.mu.RUnlock()
	if !ok {
		return nil, errors.New("file not found")
	}
	return io.NopCloser(strings.NewReader(string(data))), nil
}

func (m *MemoryStorage) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	m.mu.Lock()
	delete(m.files, key)
	m.mu.Unlock()
	return nil
}
