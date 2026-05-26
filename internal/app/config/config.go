package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AdminAPIKey string
	TokenSecret string

	WyzieAPIKey string
	WyzieBaseURL string
	WyzieSource  string

	GoAIBaseURL string
	GoAIAPIKey  string

	CacheDir string
	CacheTTL time.Duration

	MaxBatchChars int
	UserAgent     string

	Port     string
	Addr     string
}

func Load() Config {
	ttl := 48 * time.Hour
	if s := strings.TrimSpace(os.Getenv("GOSUBS_CACHE_TTL")); s != "" {
		if d, err := time.ParseDuration(s); err == nil && d > 0 {
			ttl = d
		}
	}
	maxBatchChars := 4000
	if s := strings.TrimSpace(os.Getenv("GOSUBS_MAX_BATCH_CHARS")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			maxBatchChars = n
		}
	}
	cacheDir := strings.TrimSpace(os.Getenv("GOSUBS_CACHE_DIR"))
	if cacheDir == "" {
		cacheDir = filepath.Join(os.TempDir(), "gosubs-cache")
	}
	return Config{
		AdminAPIKey:  strings.TrimSpace(os.Getenv("GOSUBS_ADMIN_API_KEY")),
		TokenSecret:  strings.TrimSpace(os.Getenv("GOSUBS_TOKEN_SECRET")),
		WyzieAPIKey:  strings.TrimSpace(os.Getenv("GOSUBS_WYZIE_API_KEY")),
		WyzieBaseURL: defaultString(strings.TrimSpace(os.Getenv("GOSUBS_WYZIE_BASE_URL")), "https://sub.wyzie.io"),
		WyzieSource:  defaultString(strings.TrimSpace(os.Getenv("GOSUBS_WYZIE_SOURCE")), "all"),
		GoAIBaseURL:  strings.TrimSpace(os.Getenv("GOSUBS_GOAI_BASE_URL")),
		GoAIAPIKey:   strings.TrimSpace(os.Getenv("GOSUBS_GOAI_API_KEY")),
		CacheDir:     cacheDir,
		CacheTTL:     ttl,
		MaxBatchChars: maxBatchChars,
		UserAgent:    defaultString(strings.TrimSpace(os.Getenv("GOSUBS_USER_AGENT")), "GoSubs/1.0"),
		Port:         strings.TrimSpace(os.Getenv("PORT")),
		Addr:         strings.TrimSpace(os.Getenv("GOSUBS_HTTP_ADDR")),
	}
}

func (c Config) Validate() error {
	if c.AdminAPIKey == "" {
		return fmt.Errorf("GOSUBS_ADMIN_API_KEY is required")
	}
	if c.TokenSecret == "" {
		c.TokenSecret = c.AdminAPIKey
	}
	if c.WyzieAPIKey == "" {
		return fmt.Errorf("GOSUBS_WYZIE_API_KEY is required")
	}
	if c.GoAIBaseURL == "" {
		return fmt.Errorf("GOSUBS_GOAI_BASE_URL is required")
	}
	if c.GoAIAPIKey == "" {
		return fmt.Errorf("GOSUBS_GOAI_API_KEY is required")
	}
	return nil
}

func (c Config) HTTPAddr() string {
	if c.Addr != "" {
		return c.Addr
	}
	if c.Port != "" {
		return ":" + c.Port
	}
	return ":8090"
}

func defaultString(v, fallback string) string {
	if v != "" {
		return v
	}
	return fallback
}
