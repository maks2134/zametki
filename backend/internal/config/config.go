package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr     string
	DatabaseURL  string
	JWTSecret    string
	JWTTTL       time.Duration
	CORSOrigins  []string
}

func Load() Config {
	ttl, err := time.ParseDuration(envOr("JWT_TTL", "720h"))
	if err != nil {
		ttl = 720 * time.Hour
	}

	origins := strings.Split(envOr("CORS_ORIGINS", "http://localhost:3000"), ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return Config{
		HTTPAddr:    listenAddr(),
		DatabaseURL: envOr("DATABASE_URL", "postgres://zametka:zametka@localhost:5432/zametka?sslmode=disable"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		JWTTTL:      ttl,
		CORSOrigins: origins,
	}
}

func listenAddr() string {
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		return normalizeAddr(v)
	}
	if v := os.Getenv("PORT"); v != "" {
		return normalizeAddr(v)
	}
	return ":8080"
}

func normalizeAddr(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ":8080"
	}
	if strings.Contains(v, ":") {
		return v
	}
	return ":" + v
}

func (c Config) Validate() error {
	if c.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 bytes")
	}
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
