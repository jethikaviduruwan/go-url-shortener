package main

import (
	"net/url"
	"strings"
)

// AllowedSchemes defines the URL schemes we accept.
var AllowedSchemes = []string{"http", "https"}

// ValidateURL checks whether a raw URL string is safe to shorten.
func ValidateURL(raw string) error {
	if raw == "" {
		return ErrInvalidURL
	}
	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return ErrInvalidURL
	}
	scheme := strings.ToLower(u.Scheme)
	for _, s := range AllowedSchemes {
		if scheme == s {
			return nil
		}
	}
	return ErrInvalidURL
}

// SanitizeURL strips tracking parameters from common platforms.
func SanitizeURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	q := u.Query()
	trackers := []string{"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content", "fbclid", "gclid"}
	changed := false
	for _, t := range trackers {
		if q.Has(t) {
			q.Del(t)
			changed = true
		}
	}
	if changed {
		u.RawQuery = q.Encode()
	}
	return u.String()
}
// v4-0
// v8-0
// v12-0
