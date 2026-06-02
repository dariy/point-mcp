package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dariy/point-mcp/internal/point"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestReplaceInPost(t *testing.T) {
	now := time.Now().Truncate(time.Second)
	post := point.Post{
		ID:        1,
		Title:     "Test Post",
		Content:   "Hello World World",
		CSS:       "body { color: red; }",
		Excerpt:   StringPtr("This is an excerpt"),
		UpdatedAt: now,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		// Handle GetPostByID
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/1") {
			json.NewEncoder(w).Encode(post)
			return
		}

		// Handle UpdatePost
		if r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/1") {
			var req point.UpdatePostRequest
			json.NewDecoder(r.Body).Decode(&req)
			
			updated := post
			if req.Content != "" {
				updated.Content = req.Content
			}
			if req.CSS != "" {
				updated.CSS = req.CSS
			}
			if req.Excerpt != "" {
				updated.Excerpt = &req.Excerpt
			}
			updated.UpdatedAt = now.Add(time.Minute) // Simulate update

			json.NewEncoder(w).Encode(struct {
				Post        point.Post `json:"post"`
				CSSWarnings []string   `json:"css_warnings"`
			}{Post: updated})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := point.New(ts.URL, "key", nil)
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "point-mcp",
		Version: "0.0.1",
	}, nil)
	Register(s, c)

	ctx := context.Background()

	t.Run("Basic Replacement", func(t *testing.T) {
		args, _ := json.Marshal(replaceInPostInput{
			ID:        1,
			Field:     "content",
			OldString: "Hello",
			NewString: "Hi",
		})
		res, err := Dispatch(ctx, "point_replace_in_post", args)
		if err != nil {
			t.Fatalf("Dispatch failed: %v", err)
		}
		result := res.(postWriteResult)
		if result.Post.Content != "Hi World World" {
			t.Errorf("expected 'Hi World World', got '%s'", result.Post.Content)
		}
	})

	t.Run("Multiple Matches Fail", func(t *testing.T) {
		args, _ := json.Marshal(replaceInPostInput{
			ID:            1,
			Field:         "content",
			OldString:     "World",
			NewString:     "Earth",
			AllowMultiple: false,
		})
		_, err := Dispatch(ctx, "point_replace_in_post", args)
		if err == nil || !strings.Contains(err.Error(), "found 2 occurrences") {
			t.Errorf("expected error about multiple occurrences, got: %v", err)
		}
	})

	t.Run("Multiple Matches Success", func(t *testing.T) {
		args, _ := json.Marshal(replaceInPostInput{
			ID:            1,
			Field:         "content",
			OldString:     "World",
			NewString:     "Earth",
			AllowMultiple: true,
		})
		res, err := Dispatch(ctx, "point_replace_in_post", args)
		if err != nil {
			t.Fatalf("Dispatch failed: %v", err)
		}
		result := res.(postWriteResult)
		if result.Post.Content != "Hello Earth Earth" {
			t.Errorf("expected 'Hello Earth Earth', got '%s'", result.Post.Content)
		}
	})

	t.Run("Optimistic Locking Success", func(t *testing.T) {
		args, _ := json.Marshal(replaceInPostInput{
			ID:                1,
			Field:             "content",
			OldString:         "Hello",
			NewString:         "Hi",
			ExpectedUpdatedAt: now.Format(time.RFC3339),
		})
		_, err := Dispatch(ctx, "point_replace_in_post", args)
		if err != nil {
			t.Fatalf("Dispatch failed: %v", err)
		}
	})

	t.Run("Optimistic Locking Failure", func(t *testing.T) {
		args, _ := json.Marshal(replaceInPostInput{
			ID:                1,
			Field:             "content",
			OldString:         "Hello",
			NewString:         "Hi",
			ExpectedUpdatedAt: now.Add(time.Hour).Format(time.RFC3339),
		})
		_, err := Dispatch(ctx, "point_replace_in_post", args)
		if err == nil || !strings.Contains(err.Error(), "optimistic locking failure") {
			t.Errorf("expected optimistic locking failure, got: %v", err)
		}
	})

	t.Run("Idempotency", func(t *testing.T) {
		args, _ := json.Marshal(replaceInPostInput{
			ID:        1,
			Field:     "content",
			OldString: "SomethingElse",
			NewString: "World",
		})
		res, err := Dispatch(ctx, "point_replace_in_post", args)
		if err != nil {
			t.Fatalf("Dispatch failed: %v", err)
		}
		result := res.(postWriteResult)
		// Should return current post without error because "World" is already there
		if result.Post.Content != "Hello World World" {
			t.Errorf("expected 'Hello World World', got '%s'", result.Post.Content)
		}
	})
}

func StringPtr(s string) *string { return &s }
