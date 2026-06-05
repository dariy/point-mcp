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
		ID:   1,
		Name: "Test Tag",
		Slug: "test-tag",
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/1") {
			json.NewEncoder(w).Encode(tag)
			return
		}

		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/slug/test-tag") {
			json.NewEncoder(w).Encode(tag)
			return
		}

		if r.Method == http.MethodPut && strings.HasSuffix(r.URL.Path, "/1") {
			var req point.UpdateTagRequest
			json.NewDecoder(r.Body).Decode(&req)
			updated := tag
			if req.Name != "" {
				updated.Name = req.Name
			}
			json.NewEncoder(w).Encode(updated)
			return
		}

		if r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/1") {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/1/geocode") {
			json.NewEncoder(w).Encode(point.TagLocation{Latitude: 10, Longitude: 20})
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

	t.Run("Get Tag by ID", func(t *testing.T) {
		args, _ := json.Marshal(getTagInput{ID: 1})
		res, err := Dispatch(ctx, "point_get_tag", args)
		if err != nil {
			t.Fatalf("Dispatch failed: %v", err)
		}
		result := res.(point.TagDetail)
		if result.ID != 1 {
			t.Errorf("expected ID 1, got %d", result.ID)
		}
	})

	t.Run("Get Tag by Slug", func(t *testing.T) {
		args, _ := json.Marshal(getTagInput{Slug: "test-tag"})
		res, err := Dispatch(ctx, "point_get_tag", args)
		if err != nil {
			t.Fatalf("Dispatch failed: %v", err)
		}
		result := res.(point.TagDetail)
		if result.Slug != "test-tag" {
			t.Errorf("expected slug 'test-tag', got '%s'", result.Slug)
		}
	})

	t.Run("Update Tag", func(t *testing.T) {
		args, _ := json.Marshal(updateTagInput{
			ID:   1,
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
		args, _ := json.Marshal(tagIDInput{ID: 1})
		_, err := Dispatch(ctx, "point_delete_tag", args)
		if err != nil {
			t.Fatalf("Dispatch failed: %v", err)
		}
	})

	t.Run("Geocode Tag", func(t *testing.T) {
		args, _ := json.Marshal(tagIDInput{ID: 1})
		res, err := Dispatch(ctx, "point_geocode_tag", args)
		if err != nil {
			t.Fatalf("Dispatch failed: %v", err)
		}
		result := res.(point.TagLocation)
		if result.Latitude != 10 {
			t.Errorf("expected lat 10, got %f", result.Latitude)
		}
	})
}
