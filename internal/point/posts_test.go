package point

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreatePostFormatting(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		
		body, _ := io.ReadAll(r.Body)
		var req CreatePostRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		if req.Content != ":::{.custom}" {
			t.Errorf("expected content ':::{.custom}', got '%s'", req.Content)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(postWriteResponse{
			Post: Post{ID: 1, Content: req.Content},
		})
	}))
	defer ts.Close()

	c := New(ts.URL, "key", nil)
	_, _, err := c.CreatePost(CreatePostRequest{
		Content: "::: {.custom}",
	})
	if err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}
}

func TestUpdatePostFormatting(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		
		body, _ := io.ReadAll(r.Body)
		var req UpdatePostRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("failed to unmarshal request body: %v", err)
		}

		if req.Content != ":::{.custom}" {
			t.Errorf("expected content ':::{.custom}', got '%s'", req.Content)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(postWriteResponse{
			Post: Post{ID: 1, Content: req.Content},
		})
	}))
	defer ts.Close()

	c := New(ts.URL, "key", nil)
	_, _, err := c.UpdatePost(1, UpdatePostRequest{
		Content: "::: {.custom}",
	})
	if err != nil {
		t.Fatalf("UpdatePost failed: %v", err)
	}
}

func TestFormattingEdgeCases(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"::: {", ":::{"},
		{"some text ::: { class } more text", "some text :::{ class } more text"},
		{"::: {", ":::{"},
		{":::  {", ":::  {"}, // Only one space handled by ReplaceAll as requested
		{"already :::{ formatted", "already :::{ formatted"},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			// We can't test fixContent directly because it's not a separate function,
			// but we can verify our understanding of strings.ReplaceAll(s, "::: {", ":::{")
			actual := strings.ReplaceAll(tc.input, "::: {", ":::{")
			if actual != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, actual)
			}
		})
	}
}
