package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ListenAddr      string
	Upstream        string
	DatabaseURL     string
	LocalDataDir    string
	DeploymentID    string
	NewAPIImage     string
	TapImage        string
	MaxBodyBytes    int64
	ProxyOnly       bool
	DevUserID       int
	DenyUserIDs     map[int]struct{}
	AdminKey        string
	RetentionDays   int
}

func Load() Config {
	maxBody := int64(32 << 20)
	if v := os.Getenv("CORPUS_TAP_MAX_BODY_BYTES"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			maxBody = n
		}
	}
	devUser := 0
	if v := os.Getenv("CORPUS_TAP_DEV_USER_ID"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			devUser = n
		}
	}
	retention := 90
	if v := os.Getenv("CORPUS_TAP_RETENTION_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			retention = n
		}
	}
	localDir := os.Getenv("CORPUS_TAP_LOCAL_DATA_DIR")
	if localDir == "" && os.Getenv("CORPUS_TAP_S3_BUCKET") == "" {
		localDir = "./data"
	}
	deny := parseIntSet(os.Getenv("CORPUS_TAP_DENY_USER_IDS"))
	return Config{
		ListenAddr:    envOr("CORPUS_TAP_LISTEN", ":8443"),
		Upstream:      os.Getenv("CORPUS_TAP_UPSTREAM"),
		DatabaseURL:   os.Getenv("CORPUS_TAP_DATABASE_URL"),
		LocalDataDir:  localDir,
		DeploymentID:  os.Getenv("CORPUS_TAP_DEPLOYMENT_ID"),
		NewAPIImage:   envOr("CORPUS_TAP_NEWAPI_IMAGE", "unknown"),
		TapImage:      envOr("CORPUS_TAP_IMAGE", "corpus-tap:dev"),
		MaxBodyBytes:  maxBody,
		ProxyOnly:     strings.EqualFold(os.Getenv("CORPUS_TAP_MODE"), "proxy-only"),
		DevUserID:     devUser,
		DenyUserIDs:   deny,
		AdminKey:      os.Getenv("CORPUS_TAP_ADMIN_KEY"),
		RetentionDays: retention,
	}
}

func (c Config) Valid() error {
	if c.Upstream == "" {
		return errMissing("CORPUS_TAP_UPSTREAM")
	}
	return nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseIntSet(s string) map[int]struct{} {
	out := make(map[int]struct{})
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if n, err := strconv.Atoi(part); err == nil {
			out[n] = struct{}{}
		}
	}
	return out
}

type missingEnv string

func (m missingEnv) Error() string { return "missing required env: " + string(m) }

func errMissing(key string) error { return missingEnv(key) }
