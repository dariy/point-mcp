package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dariy/point-mcp/internal/point"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestTagTools(t *testing.T) {
	tag := point.TagDetail{
		Name: "Test Tag",
		Slug: "test-tag",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/test-tag") {
			json.NewEncoder(w).Encode(tag)
			return
		}

		if r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/test-tag") {
			var req point.UpdateTagRequest
			json.NewDecoder(r.Body).Decode(&req)
			updated := tag
			if req.Name != "" {
				updated.Name = req.Name
			}
			json.NewEncoder(w).Encode(updated)
			return
		}

		if r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/test-tag") {
			w.WriteHeader(http.StatusNoContent)
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

	t.Run("Get Tag", func(t *testing.T) {
		args, _ := json.Marshal(tagSlugInput{Slug: "test-tag"})
		res, err := Dispatch(ctx, "point_get_tag", args)
		if err != nil {
			t.Fatalf("Dispatch failed: %v", err)
		}
		result := res.(point.TagDetail)
		if result.Name != "Test Tag" {
			t.Errorf("expected 'Test Tag', got '%s'", result.Name)
		}
	})

	t.Run("Update Tag", func(t *testing.T) {
		args, _ := json.Marshal(updateTagInput{
			Slug: "test-tag",
			Name: "Updated Name",
		})
		res, err := Dispatch(ctx, "point_update_tag", args)
		if err != nil {
			t.Fatalf("Dispatch failed: %v", err)
		}
		result := res.(point.TagDetail)
		if result.Name != "Updated Name" {
			t.Errorf("expected 'Updated Name', got '%s'", result.Name)
		}
	})

	t.Run("Delete Tag", func(t *testing.T) {
		args, _ := json.Marshal(tagSlugInput{Slug: "test-tag"})
		_, err := Dispatch(ctx, "point_delete_tag", args)
		if err != nil {
			t.Fatalf("Dispatch failed: %v", err)
		}
	})
}
