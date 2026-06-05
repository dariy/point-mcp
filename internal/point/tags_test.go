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
		if r.URL.Path != "/api/tags/test-tag" {
			t.Errorf("expected path /api/tags/test-tag, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TagDetail{
			Name: "Test Tag",
			Slug: "test-tag",
			Parent: &TagSummary{
				Name: "Parent Tag",
				Slug: "parent-tag",
			},
		})
	}))
	defer ts.Close()

	c := New(ts.URL, "key", nil)
	tag, err := c.GetTag("test-tag")
	if err != nil {
		t.Fatalf("GetTag failed: %v", err)
	}

	if tag.Name != "Test Tag" {
		t.Errorf("expected name 'Test Tag', got '%s'", tag.Name)
	}
	if tag.Parent == nil || tag.Parent.Slug != "parent-tag" {
		t.Errorf("expected parent slug 'parent-tag', got %+v", tag.Parent)
	}
}

func TestUpdateTag(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/tags/test-tag" {
			t.Errorf("expected path /api/tags/test-tag, got %s", r.URL.Path)
		}

		var req UpdateTagRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Name != "New Name" {
			t.Errorf("expected name 'New Name', got '%s'", req.Name)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(TagDetail{
			Name: req.Name,
			Slug: "test-tag",
		})
	}))
	defer ts.Close()

	c := New(ts.URL, "key", nil)
	tag, err := c.UpdateTag("test-tag", UpdateTagRequest{
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
		if r.URL.Path != "/api/tags/test-tag" {
			t.Errorf("expected path /api/tags/test-tag, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := New(ts.URL, "key", nil)
	err := c.DeleteTag("test-tag")
	if err != nil {
		t.Fatalf("DeleteTag failed: %v", err)
	}
}
