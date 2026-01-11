package main

import (
	"fmt"
	"os"
)

type Config struct {
	Port     string
	BaseURL  string
	Version  string
	CodeLen  int
}

func mustLoadConfig() Config {
	port := getEnv("PORT", "9090")
	base := getEnv("BASE_URL", fmt.Sprintf("http://localhost:%s", port))
	return Config{
		Port:    port,
		BaseURL: base,
		Version: "1.2.0",
		CodeLen: 7,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
