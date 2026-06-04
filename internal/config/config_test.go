package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dariy/point-mcp/internal/config"
)

func setenv(t *testing.T, key, val string) {
	t.Helper()
	t.Setenv(key, val)
}

func TestLoad_missingRequired(t *testing.T) {
	t.Setenv("POINT_BASE_URL", "")
	t.Setenv("POINT_API_KEY", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when both required vars are missing")
	}
}

func TestLoad_missingBaseURL(t *testing.T) {
	t.Setenv("POINT_BASE_URL", "")
	setenv(t, "POINT_API_KEY", "key")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when POINT_BASE_URL is missing")
	}
}

func TestLoad_missingAPIKey(t *testing.T) {
	setenv(t, "POINT_BASE_URL", "https://example.com")
	t.Setenv("POINT_API_KEY", "")

	_, err := config.Load()
	if err == nil {
		t.Fatal("expected error when POINT_API_KEY is missing")
	}
}

func TestLoad_defaults(t *testing.T) {
	setenv(t, "POINT_BASE_URL", "https://example.com")
	setenv(t, "POINT_API_KEY", "secret")
	t.Setenv("POINT_MCP_TRANSPORT", "")
	t.Setenv("POINT_MCP_HTTP_ADDR", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Transport != "stdio" {
		t.Errorf("expected transport=stdio, got %q", cfg.Transport)
	}
	if cfg.HTTPAddr != ":8000" {
		t.Errorf("expected httpAddr=:8000, got %q", cfg.HTTPAddr)
	}
}

func writeDotEnv(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoad_dotEnvFile(t *testing.T) {
	t.Setenv("POINT_BASE_URL", "")
	t.Setenv("POINT_API_KEY", "")
	t.Setenv("POINT_MCP_TRANSPORT", "")
	t.Setenv("POINT_MCP_HTTP_ADDR", "")

	path := writeDotEnv(t, `
# comment line
POINT_BASE_URL=https://dotenv.example.com
POINT_API_KEY=dotenv-key
POINT_MCP_TRANSPORT=http
POINT_MCP_HTTP_ADDR=:8888
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BaseURL != "https://dotenv.example.com" {
		t.Errorf("got BaseURL %q", cfg.BaseURL)
	}
	if cfg.APIKey != "dotenv-key" {
		t.Errorf("got APIKey %q", cfg.APIKey)
	}
	if cfg.Transport != "http" {
		t.Errorf("got Transport %q", cfg.Transport)
	}
	if cfg.HTTPAddr != ":8888" {
		t.Errorf("got HTTPAddr %q", cfg.HTTPAddr)
	}
}

func TestLoad_dotEnvEnvVarWins(t *testing.T) {
	t.Setenv("POINT_BASE_URL", "https://real-env.example.com")
	t.Setenv("POINT_API_KEY", "")

	path := writeDotEnv(t, `
POINT_BASE_URL=https://dotenv.example.com
POINT_API_KEY=dotenv-key
`)

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BaseURL != "https://real-env.example.com" {
		t.Errorf("env var should win: got BaseURL %q", cfg.BaseURL)
	}
	if cfg.APIKey != "dotenv-key" {
		t.Errorf("dotenv should fill missing var: got APIKey %q", cfg.APIKey)
	}
}

func TestLoad_explicit(t *testing.T) {
	setenv(t, "POINT_BASE_URL", "https://api.example.com")
	setenv(t, "POINT_API_KEY", "mykey")
	setenv(t, "POINT_MCP_TRANSPORT", "http")
	setenv(t, "POINT_MCP_HTTP_ADDR", ":8080")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BaseURL != "https://api.example.com" {
		t.Errorf("got BaseURL %q", cfg.BaseURL)
	}
	if cfg.APIKey != "mykey" {
		t.Errorf("got APIKey %q", cfg.APIKey)
	}
	if cfg.Transport != "http" {
		t.Errorf("got Transport %q", cfg.Transport)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Errorf("got HTTPAddr %q", cfg.HTTPAddr)
	}
}
