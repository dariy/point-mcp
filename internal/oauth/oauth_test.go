package oauth_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/dariy/point-mcp/internal/oauth"
)

func newTestProvider() (*oauth.Provider, *http.ServeMux) {
	p := oauth.New(oauth.Config{
		BaseURL:  "https://example.com",
		Password: "testpassword",
	})
	mux := http.NewServeMux()
	p.Register(mux)
	return p, mux
}

func registerClient(t *testing.T, mux *http.ServeMux, redirectURI string) string {
	t.Helper()
	body := `{"redirect_uris":["` + redirectURI + `"]}`
	req := httptest.NewRequest("POST", "/oauth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]any
	json.Unmarshal(rr.Body.Bytes(), &resp)
	return resp["client_id"].(string)
}

func pkce(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// TestHappyPath covers register → authorize → token → bearer validation.
func TestHappyPath(t *testing.T) {
	p, mux := newTestProvider()
	const redirectURI = "https://example.com/callback"
	clientID := registerClient(t, mux, redirectURI)

	verifier := "testverifier_abcdefghijklmnopqrstuvwxyz01234567"
	challenge := pkce(verifier)

	// GET /oauth/authorize — should render login page
	authURL := "/oauth/authorize?response_type=code&client_id=" + clientID +
		"&redirect_uri=" + url.QueryEscape(redirectURI) +
		"&state=xyz&code_challenge=" + url.QueryEscape(challenge) +
		"&code_challenge_method=S256"
	req := httptest.NewRequest("GET", authURL, nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("authorize GET: expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Header().Get("Content-Type"), "text/html") {
		t.Fatal("authorize GET: expected HTML response")
	}

	// POST /oauth/authorize — submit correct password
	form := url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"state":                 {"xyz"},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
		"submitted_password":    {"testpassword"},
	}
	req = httptest.NewRequest("POST", "/oauth/authorize", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusFound {
		t.Fatalf("authorize POST: expected 302, got %d: %s", rr.Code, rr.Body.String())
	}
	location := rr.Header().Get("Location")
	redirURL, err := url.Parse(location)
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	code := redirURL.Query().Get("code")
	if code == "" {
		t.Fatal("missing code in redirect")
	}
	if redirURL.Query().Get("state") != "xyz" {
		t.Fatalf("state mismatch: got %q", redirURL.Query().Get("state"))
	}

	// POST /oauth/token — exchange code for tokens
	tokenForm := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"code_verifier": {verifier},
	}
	req = httptest.NewRequest("POST", "/oauth/token", strings.NewReader(tokenForm.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("token: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var tokenResp map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &tokenResp); err != nil {
		t.Fatalf("decode token response: %v", err)
	}
	accessToken, _ := tokenResp["access_token"].(string)
	refreshToken, _ := tokenResp["refresh_token"].(string)
	if accessToken == "" {
		t.Fatal("missing access_token")
	}
	if refreshToken == "" {
		t.Fatal("missing refresh_token")
	}
	if tokenResp["token_type"] != "Bearer" {
		t.Fatalf("token_type: got %v", tokenResp["token_type"])
	}

	// RequireBearer — valid token should pass through
	protected := p.RequireBearer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req = httptest.NewRequest("GET", "/mcp", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	rr = httptest.NewRecorder()
	protected.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("RequireBearer with valid token: expected 200, got %d", rr.Code)
	}

	// RequireBearer — missing token should 401
	req = httptest.NewRequest("GET", "/mcp", nil)
	rr = httptest.NewRecorder()
	protected.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("RequireBearer without token: expected 401, got %d", rr.Code)
	}

	// Refresh token grant
	refreshForm := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}
	req = httptest.NewRequest("POST", "/oauth/token", strings.NewReader(refreshForm.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("refresh: expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var refreshResp map[string]any
	json.Unmarshal(rr.Body.Bytes(), &refreshResp)
	if refreshResp["access_token"] == "" {
		t.Fatal("refresh: missing access_token")
	}
}

// TestPKCEFailure ensures a wrong code_verifier is rejected.
func TestPKCEFailure(t *testing.T) {
	_, mux := newTestProvider()
	const redirectURI = "https://example.com/callback"
	clientID := registerClient(t, mux, redirectURI)

	verifier := "correctverifier_abcdefghijklmnopqrstuvwxyz01234"
	challenge := pkce(verifier)

	form := url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"state":                 {"s"},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
		"submitted_password":    {"testpassword"},
	}
	req := httptest.NewRequest("POST", "/oauth/authorize", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	location := rr.Header().Get("Location")
	redirURL, _ := url.Parse(location)
	code := redirURL.Query().Get("code")

	// Submit wrong verifier
	tokenForm := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"client_id":     {clientID},
		"redirect_uri":  {redirectURI},
		"code_verifier": {"wrongverifier_abcdefghijklmnopqrstuvwxyz0123"},
	}
	req = httptest.NewRequest("POST", "/oauth/token", strings.NewReader(tokenForm.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr = httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("PKCE failure: expected 400, got %d", rr.Code)
	}
	var errResp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &errResp)
	if errResp["error"] != "invalid_grant" {
		t.Fatalf("PKCE failure: expected invalid_grant, got %q", errResp["error"])
	}
}

// TestWrongPassword ensures bad password re-renders the login page without redirecting.
func TestWrongPassword(t *testing.T) {
	_, mux := newTestProvider()
	clientID := registerClient(t, mux, "https://example.com/cb")

	form := url.Values{
		"client_id":             {clientID},
		"redirect_uri":          {"https://example.com/cb"},
		"state":                 {"s"},
		"code_challenge":        {"challenge"},
		"code_challenge_method": {"S256"},
		"submitted_password":    {"wrongpassword"},
	}
	req := httptest.NewRequest("POST", "/oauth/authorize", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("wrong password: expected 200 (re-render), got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Wrong password") {
		t.Fatal("wrong password: expected error message in response")
	}
}

// TestRegisterValidation ensures missing redirect_uris returns 400.
func TestRegisterValidation(t *testing.T) {
	_, mux := newTestProvider()
	req := httptest.NewRequest("POST", "/oauth/register", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("register without redirect_uris: expected 400, got %d", rr.Code)
	}
	var errResp map[string]string
	json.Unmarshal(rr.Body.Bytes(), &errResp)
	if errResp["error"] != "invalid_client_metadata" {
		t.Fatalf("expected invalid_client_metadata, got %q", errResp["error"])
	}
}
