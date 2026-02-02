package main

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrNotFound      = errors.New("short code not found")
	ErrCodeTaken     = errors.New("custom code already in use")
	ErrInvalidURL    = errors.New("invalid URL")
)

type Link struct {
	Code      string    `json:"code"`
	LongURL   string    `json:"long_url"`
	ShortURL  string    `json:"short_url"`
	Clicks    int64     `json:"clicks"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

func (l *Link) IsExpired() bool {
	return l.ExpiresAt != nil && time.Now().After(*l.ExpiresAt)
}

type Store struct {
	mu    sync.RWMutex
	links map[string]*Link
}

func newStore() *Store {
	return &Store{links: make(map[string]*Link)}
}

func (s *Store) Save(link *Link) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.links[link.Code]; exists {
		return ErrCodeTaken
	}
	s.links[link.Code] = link
	return nil
}

func (s *Store) Get(code string) (*Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.links[code]
	if !ok {
		return nil, ErrNotFound
	}
	return l, nil
}

func (s *Store) IncrClick(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if l, ok := s.links[code]; ok {
		l.Clicks++
	}
}

func (s *Store) Delete(code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.links[code]; !ok {
		return ErrNotFound
	}
	delete(s.links, code)
	return nil
}

func (s *Store) All() []*Link {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Link, 0, len(s.links))
	for _, l := range s.links {
		out = append(out, l)
	}
	return out
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.links)
}
// v2-0
// v6-0
// v9-2
