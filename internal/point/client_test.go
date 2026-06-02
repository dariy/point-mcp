package point

import (
	"net/http"
	"testing"
	"time"
)

func TestNewClientTimeout(t *testing.T) {
	c := New("http://localhost:8000", "key", nil)
	if c.http.Timeout != 30*time.Second {
		t.Errorf("expected 30s timeout, got %v", c.http.Timeout)
	}

	customClient := &http.Client{Timeout: 5 * time.Second}
	c2 := New("http://localhost:8000", "key", customClient)
	if c2.http.Timeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", c2.http.Timeout)
	}
}
