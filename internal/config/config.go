package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr        string
	ShutdownTimeout time.Duration
	CacheEntries    int
	WorkerTTL       time.Duration
	LogLevel        string
}

func Load() Config {
	c := Config{HTTPAddr: ":8080", ShutdownTimeout: 10 * time.Second, CacheEntries: 10000, WorkerTTL: 30 * time.Second, LogLevel: "info"}
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		c.HTTPAddr = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		c.LogLevel = strings.ToLower(v)
	}
	if v := os.Getenv("CACHE_ENTRIES"); v != "" {
		if n, e := strconv.Atoi(v); e == nil {
			c.CacheEntries = n
		}
	}
	return c
}
func (c Config) Validate() error {
	if c.HTTPAddr == "" {
		return fmt.Errorf("http address is required")
	}
	if c.CacheEntries < 1 || c.CacheEntries > 1000000 {
		return fmt.Errorf("cache entries out of range")
	}
	if c.ShutdownTimeout <= 0 {
		return fmt.Errorf("shutdown timeout must be positive")
	}
	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("invalid log level")
	}
	return nil
}
func Redact(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + "..." + value[len(value)-2:]
}
