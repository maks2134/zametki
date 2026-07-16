package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr    string
	MongoURI    string
	MongoDB     string
	JWTSecret   string
	JWTTTL      time.Duration
	CORSOrigins []string
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
		MongoURI:    envOr("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:     envOr("MONGO_DB", "zametka"),
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
	if c.MongoURI == "" {
		return fmt.Errorf("MONGO_URI is required")
	}
	if c.MongoDB == "" {
		return fmt.Errorf("MONGO_DB is required")
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
