package point

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTag(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/tags/1" {
			t.Errorf("expected path /api/tags/1, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TagDetail{
			ID:   1,
			Name: "Test Tag",
			Slug: "test-tag",
			Parents: []TagSummary{
				{ID: 2, Name: "Parent Tag", Slug: "parent-tag"},
			},
		})
	}))
	defer ts.Close()

	c := New(ts.URL, "key", nil)
	tag, err := c.GetTag(1)
	if err != nil {
		t.Fatalf("GetTag failed: %v", err)
	}

	if tag.Name != "Test Tag" {
		t.Errorf("expected name 'Test Tag', got '%s'", tag.Name)
	}
	if len(tag.Parents) == 0 || tag.Parents[0].Slug != "parent-tag" {
		t.Errorf("expected parent slug 'parent-tag', got %+v", tag.Parents)
	}
}

func TestUpdateTag(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/tags/1" {
			t.Errorf("expected path /api/tags/1, got %s", r.URL.Path)
		}

		var req UpdateTagRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Name != "New Name" {
			t.Errorf("expected name 'New Name', got '%s'", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TagDetail{
			ID:   1,
			Name: req.Name,
			Slug: "test-tag",
		})
	}))
	defer ts.Close()

	c := New(ts.URL, "key", nil)
	tag, err := c.UpdateTag(1, UpdateTagRequest{
		Name: "New Name",
	})
	if err != nil {
		t.Fatalf("UpdateTag failed: %v", err)
	}

	if tag.Name != "New Name" {
		t.Errorf("expected name 'New Name', got '%s'", tag.Name)
	}
}

func TestDeleteTag(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/api/tags/1" {
			t.Errorf("expected path /api/tags/1, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := New(ts.URL, "key", nil)
	err := c.DeleteTag(1)
	if err != nil {
		t.Fatalf("DeleteTag failed: %v", err)
	}
}

func TestGeocodeTag(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/tags/1/geocode" {
			t.Errorf("expected path /api/tags/1/geocode, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TagLocation{
			Latitude:  45.507,
			Longitude: -73.554,
		})
	}))
	defer ts.Close()

	c := New(ts.URL, "key", nil)
	loc, err := c.GeocodeTag(1)
	if err != nil {
		t.Fatalf("GeocodeTag failed: %v", err)
	}

	if loc.Latitude != 45.507 {
		t.Errorf("expected lat 45.507, got %f", loc.Latitude)
	}
}
