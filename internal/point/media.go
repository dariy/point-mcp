package point

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

// MediaFilter holds optional query parameters for ListMedia.
type MediaFilter struct {
	Page    int
	PerPage int
}

func (f MediaFilter) query() string {
	v := url.Values{}
	if f.Page > 0 {
		v.Set("page", strconv.Itoa(f.Page))
	}
	if f.PerPage > 0 {
		v.Set("per_page", strconv.Itoa(f.PerPage))
	}
	if qs := v.Encode(); qs != "" {
		return "?" + qs
	}
	return ""
}

func (c *Client) ListMedia(filter MediaFilter) (MediaList, error) {
	return get[MediaList](c, "/api/media"+filter.query())
}

// UploadFile uploads a local file to the media library via multipart form.
func (c *Client) UploadFile(filePath string) (MediaItem, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return MediaItem{}, err
	}
	defer f.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return MediaItem{}, err
	}
	if _, err = io.Copy(part, f); err != nil {
		return MediaItem{}, err
	}
	if err = w.Close(); err != nil {
		return MediaItem{}, err
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/media/upload", &buf)
	if err != nil {
		return MediaItem{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return MediaItem{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		var errResp ErrorResponse
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		return MediaItem{}, &APIError{StatusCode: resp.StatusCode, Detail: errResp.Detail}
	}
	return decode[MediaItem](resp)
}

func (c *Client) AnalyzeImageByID(id int64) (MediaAnalysis, error) {
	return post[MediaAnalysis](c, fmt.Sprintf("/api/media/%d/analyze", id), nil)
}
