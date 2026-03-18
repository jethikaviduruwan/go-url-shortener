package main

import (
	"sort"
	"sync"
	"time"
)

type ClickEvent struct {
	Code      string
	Timestamp time.Time
	Referer   string
	UserAgent string
}

type AnalyticsEngine struct {
	mu     sync.RWMutex
	events []*ClickEvent
	store  *Store
}

func newAnalyticsEngine(store *Store) *AnalyticsEngine {
	return &AnalyticsEngine{store: store}
}

func (a *AnalyticsEngine) Record(ev *ClickEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.events = append(a.events, ev)
	a.store.IncrClick(ev.Code)
}

type LinkStat struct {
	Code     string `json:"code"`
	LongURL  string `json:"long_url"`
	Clicks   int64  `json:"clicks"`
	LastSeen *time.Time `json:"last_seen,omitempty"`
}

func (a *AnalyticsEngine) TopLinks(n int) []LinkStat {
	links := a.store.All()
	stats := make([]LinkStat, 0, len(links))

	a.mu.RLock()
	lastSeen := make(map[string]time.Time)
	for _, ev := range a.events {
		if t, ok := lastSeen[ev.Code]; !ok || ev.Timestamp.After(t) {
			lastSeen[ev.Code] = ev.Timestamp
		}
	}
	a.mu.RUnlock()

	for _, l := range links {
		s := LinkStat{Code: l.Code, LongURL: l.LongURL, Clicks: l.Clicks}
		if t, ok := lastSeen[l.Code]; ok {
			s.LastSeen = &t
		}
		stats = append(stats, s)
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Clicks > stats[j].Clicks
	})
	if n > 0 && len(stats) > n {
		stats = stats[:n]
	}
	return stats
}

func (a *AnalyticsEngine) TotalClicks() int64 {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return int64(len(a.events))
}
// v3-1
// v7-0
// v10-1
// v16-0
// v22-0
// v28-1
