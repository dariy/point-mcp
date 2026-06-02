package point

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// APIError is returned for non-2xx responses, carrying the detail message from the API.
type APIError struct {
	StatusCode int
	Detail     string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("point API error %d: %s", e.StatusCode, e.Detail)
}

// Client calls the Point REST API.
type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// New creates a Client. Pass nil to use a default HTTP client with a 30s timeout.
func New(baseURL, apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	return &Client{baseURL: baseURL, apiKey: apiKey, http: httpClient}
}

func (c *Client) do(method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	var requestBody []byte
	if body != nil {
		var err error
		requestBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(requestBody)
	}

	url := c.baseURL + path
	log.Printf("--> %s %s", method, url)
	if len(requestBody) > 0 {
		displayBody := string(requestBody)
		if len(displayBody) > 1000 {
			displayBody = displayBody[:1000] + "... (truncated)"
		}
		log.Printf("Request Body: %s", displayBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	start := time.Now()
	var resp *http.Response
	var lastErr error
	for i := 0; i < 5; i++ {
		if i > 0 {
			// Re-create body reader if it's not nil, as it's consumed by Do
			if body != nil {
				requestBody, _ := json.Marshal(body)
				req.Body = io.NopCloser(bytes.NewReader(requestBody))
			}
		}

		resp, err = c.http.Do(req)
		if err != nil {
			lastErr = err
			time.Sleep(time.Duration(i*100) * time.Millisecond)
			continue
		}

		if resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusInternalServerError || resp.StatusCode == http.StatusConflict {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			bodyStr := string(bodyBytes)
			if strings.Contains(bodyStr, "database is locked") || strings.Contains(bodyStr, "SQLITE_BUSY") {
				lastErr = fmt.Errorf("API error %d: %s", resp.StatusCode, bodyStr)
				time.Sleep(time.Duration(i*100) * time.Millisecond)
				continue
			}
			// If it's not a lock error, restore body for decoding below
			resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
		lastErr = nil
		break
	}
	duration := time.Since(start)

	if lastErr != nil {
		log.Printf("<-- ERROR %s %s (%v): %v", method, url, duration, lastErr)
		return nil, lastErr
	}

	log.Printf("<-- %d %s %s (%v)", resp.StatusCode, method, url, duration)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		var errResp ErrorResponse
		bodyBytes, _ := io.ReadAll(resp.Body)
		_ = json.Unmarshal(bodyBytes, &errResp)
		log.Printf("Response Body (Error): %s", string(bodyBytes))
		return nil, &APIError{StatusCode: resp.StatusCode, Detail: errResp.Detail}
	}
	return resp, nil
}

func decode[T any](resp *http.Response) (T, error) {
	defer resp.Body.Close()
	var v T
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return v, err
	}

	displayBody := string(bodyBytes)
	if len(displayBody) > 1000 {
		displayBody = displayBody[:1000] + "... (truncated)"
	}
	log.Printf("Response Body: %s", displayBody)

	if err := json.Unmarshal(bodyBytes, &v); err != nil {
		return v, err
	}
	return v, nil
}

func get[T any](c *Client, path string) (T, error) {
	resp, err := c.do(http.MethodGet, path, nil)
	if err != nil {
		var zero T
		return zero, err
	}
	return decode[T](resp)
}

func post[T any](c *Client, path string, body any) (T, error) {
	resp, err := c.do(http.MethodPost, path, body)
	if err != nil {
		var zero T
		return zero, err
	}
	return decode[T](resp)
}

func put[T any](c *Client, path string, body any) (T, error) {
	resp, err := c.do(http.MethodPut, path, body)
	if err != nil {
		var zero T
		return zero, err
	}
	return decode[T](resp)
}

func patch[T any](c *Client, path string, body any) (T, error) {
	resp, err := c.do(http.MethodPatch, path, body)
	if err != nil {
		var zero T
		return zero, err
	}
	return decode[T](resp)
}

func delete[T any](c *Client, path string) (T, error) {
	resp, err := c.do(http.MethodDelete, path, nil)
	if err != nil {
		var zero T
		return zero, err
	}
	return decode[T](resp)
}

// noContent performs a request where no response body is expected (e.g. DELETE with 204).
func (c *Client) noContent(method, path string, body any) error {
	resp, err := c.do(method, path, body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
