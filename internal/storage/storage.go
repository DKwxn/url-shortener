package storage

import (
    "encoding/json"
    "errors"
    "os"
    "sync"
)

type Storage struct {
    mu    sync.Mutex
    file  *os.File
    store map[string]string
}

func NewStorage(filename string) (*Storage, error) {
    file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0755)
    if err != nil {
        return nil, err
    }

    s := &Storage{
        file:  file,
        store: make(map[string]string),
    }

    if err := s.load(); err != nil && !errors.Is(err, os.ErrNotExist) {
        return nil, err
    }

    return s, nil
}

func (s *Storage) load() error {
    decoder := json.NewDecoder(s.file)
    if err := decoder.Decode(&s.store); err != nil {
        // If the file is empty or contains invalid JSON, initialize an empty store
        if errors.Is(err, os.ErrNotExist) || err.Error() == "EOF" {
            s.store = make(map[string]string)
            return nil
        }
        return err
    }
    return nil
}

func (s *Storage) save() error {
    s.file.Truncate(0)
    s.file.Seek(0, 0)
    encoder := json.NewEncoder(s.file)
    return encoder.Encode(s.store)
}

func (s *Storage) Store(shortURL, originalURL string) error {
    s.mu.Lock()
    defer s.mu.Unlock()

    s.store[shortURL] = originalURL
    return s.save()
}

func (s *Storage) Load(shortURL string) (string, bool) {
    s.mu.Lock()
    defer s.mu.Unlock()

    originalURL, ok := s.store[shortURL]
    return originalURL, ok
}
