package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	BaseURL    string
	APIKey     string
	Transport  string
	HTTPAddr   string
	LogFile    string
	AuthTokens []string // MCP_AUTH_TOKENS — comma-separated, supports rotation
}

// Load reads configuration from environment variables, optionally pre-seeding
// them from .env files. Files are processed in order; real env vars win over
// file values. Defaults to loading ".env" from the working directory.
func Load(files ...string) (*Config, error) {
	if len(files) == 0 {
		files = []string{".env"}
	}
	for _, f := range files {
		if err := loadDotEnv(f); err != nil {
			return nil, fmt.Errorf("loading %s: %w", f, err)
		}
	}

	baseURL := os.Getenv("POINT_BASE_URL")
	apiKey := os.Getenv("POINT_API_KEY")

	var errs []error
	if baseURL == "" {
		errs = append(errs, errors.New("POINT_BASE_URL is required"))
	}
	if apiKey == "" {
		errs = append(errs, errors.New("POINT_API_KEY is required"))
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	transport := os.Getenv("POINT_MCP_TRANSPORT")
	if transport == "" {
		transport = "stdio"
	}

	httpAddr := os.Getenv("POINT_MCP_HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":9000"
	}

	return &Config{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		Transport:  transport,
		HTTPAddr:   httpAddr,
		LogFile:    os.Getenv("POINT_MCP_LOG_FILE"),
		AuthTokens: splitNonEmpty(os.Getenv("MCP_AUTH_TOKENS")),
	}, nil
}

func splitNonEmpty(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// loadDotEnv reads KEY=VALUE pairs from path into the process environment.
// Missing file is silently ignored. Existing env vars are never overwritten.
func loadDotEnv(path string) error {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" {
			continue
		}
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
	return scanner.Err()
}
