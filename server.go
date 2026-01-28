package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type Server struct {
	cfg       Config
	store     *Store
	analytics *AnalyticsEngine
	http      *http.Server
}

func newServer(cfg Config, store *Store, analytics *AnalyticsEngine) *Server {
	s := &Server{cfg: cfg, store: store, analytics: analytics}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/shorten", s.handleShorten)
	mux.HandleFunc("/api/links", s.handleListLinks)
	mux.HandleFunc("/api/stats", s.handleStats)
	mux.HandleFunc("/api/links/", s.handleDeleteLink)
	mux.HandleFunc("/r/", s.handleRedirect)
	mux.HandleFunc("/health", s.handleHealth)
	s.http = &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      corsMiddleware(jsonContentType(mux)),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return s
}

func (s *Server) run() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("server error: %v\n", err)
		}
	}()
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.http.Shutdown(ctx)
}

type ShortenRequest struct {
	URL        string  `json:"url"`
	CustomCode string  `json:"custom_code,omitempty"`
	TTLHours   int     `json:"ttl_hours,omitempty"`
}

func (s *Server) handleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "POST only")
		return
	}
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if _, err := url.ParseRequestURI(req.URL); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid URL")
		return
	}

	code := req.CustomCode
	if code != "" {
		if !IsValidCode(code) {
			writeErr(w, http.StatusBadRequest, "custom code must be 3-32 alphanumeric chars")
			return
		}
	} else {
		var err error
		code, err = GenerateCode(s.cfg.CodeLen)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "could not generate code")
			return
		}
	}

	link := &Link{
		Code:      code,
		LongURL:   req.URL,
		ShortURL:  fmt.Sprintf("%s/r/%s", s.cfg.BaseURL, code),
		CreatedAt: time.Now(),
	}
	if req.TTLHours > 0 {
		exp := time.Now().Add(time.Duration(req.TTLHours) * time.Hour)
		link.ExpiresAt = &exp
	}
	if err := s.store.Save(link); errors.Is(err, ErrCodeTaken) {
		writeErr(w, http.StatusConflict, "code already taken")
		return
	}
	writeJSON(w, http.StatusCreated, link)
}

func (s *Server) handleRedirect(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/r/")
	link, err := s.store.Get(code)
	if errors.Is(err, ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if link.IsExpired() {
		writeErr(w, http.StatusGone, "link has expired")
		return
	}
	s.analytics.Record(&ClickEvent{
		Code:      code,
		Timestamp: time.Now(),
		Referer:   r.Referer(),
		UserAgent: r.UserAgent(),
	})
	http.Redirect(w, r, link.LongURL, http.StatusFound)
}

func (s *Server) handleListLinks(w http.ResponseWriter, r *http.Request) {
	links := s.store.All()
	writeJSON(w, http.StatusOK, map[string]any{"links": links, "total": len(links)})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"total_links":  s.store.Count(),
		"total_clicks": s.analytics.TotalClicks(),
		"top_links":    s.analytics.TopLinks(10),
	})
}

func (s *Server) handleDeleteLink(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeErr(w, http.StatusMethodNotAllowed, "DELETE only")
		return
	}
	code := strings.TrimPrefix(r.URL.Path, "/api/links/")
	if err := s.store.Delete(code); errors.Is(err, ErrNotFound) {
		writeErr(w, http.StatusNotFound, "link not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"version": s.cfg.Version,
		"links":   s.store.Count(),
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func jsonContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
// v3-2
// v7-1
