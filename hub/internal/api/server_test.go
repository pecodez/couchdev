package api_test

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"testing/fstest"

	"go.uber.org/zap"

	"github.com/pecodez/couchdev/internal/api"
	"github.com/pecodez/couchdev/internal/db"
	"github.com/pecodez/couchdev/internal/git"
	"github.com/pecodez/couchdev/internal/tmux"
)

const testToken = "test-token"

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func newTestServer(t *testing.T) (baseURL string) {
	t.Helper()
	sum := sha256.Sum256([]byte(testToken))
	srv := httptest.NewServer(api.New(
		sum[:],
		openTestDB(t),
		tmux.NewMock(),
		fstest.MapFS{},
		t.TempDir(),
		&git.Mock{},
		zap.NewNop(),
	))
	t.Cleanup(srv.Close)
	return srv.URL
}

func newTestServerNoAuth(t *testing.T) (baseURL string) {
	t.Helper()
	srv := httptest.NewServer(api.New(
		nil,
		openTestDB(t),
		tmux.NewMock(),
		fstest.MapFS{},
		t.TempDir(),
		&git.Mock{},
		zap.NewNop(),
	))
	t.Cleanup(srv.Close)
	return srv.URL
}

func authedDo(t *testing.T, method, url, token string, body any) *http.Response {
	t.Helper()
	var r io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, url, r)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

func authedPost(t *testing.T, url, token string, body any) *http.Response {
	return authedDo(t, http.MethodPost, url, token, body)
}

func authedGet(t *testing.T, url, token string) *http.Response {
	return authedDo(t, http.MethodGet, url, token, nil)
}

func decodeJSON(t *testing.T, resp *http.Response, dst any) {
	t.Helper()
	if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
		t.Fatalf("decode response JSON: %v", err)
	}
}

// ── Auth middleware ────────────────────────────────────────────────────────────

func TestAuth_MissingToken(t *testing.T) {
	base := newTestServer(t)
	req, _ := http.NewRequest(http.MethodGet, base+"/api/projects", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuth_WrongToken(t *testing.T) {
	base := newTestServer(t)
	resp := authedGet(t, base+"/api/projects", "wrong-token")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestAuth_ValidToken(t *testing.T) {
	base := newTestServer(t)
	resp := authedGet(t, base+"/api/projects", testToken)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAuth_Disabled(t *testing.T) {
	base := newTestServerNoAuth(t)
	req, _ := http.NewRequest(http.MethodGet, base+"/api/projects", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// ── Projects ──────────────────────────────────────────────────────────────────

func TestCreateProject_Greenfield(t *testing.T) {
	base := newTestServer(t)
	payload := map[string]string{"name": "my-app", "source_type": "greenfield"}
	resp := authedPost(t, base+"/api/projects", testToken, payload)
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, body)
	}
	var p struct {
		Name       string `json:"name"`
		SourceType string `json:"source_type"`
		RepoPath   string `json:"repo_path"`
	}
	decodeJSON(t, resp, &p)
	if p.Name != "my-app" {
		t.Errorf("name = %q, want 'my-app'", p.Name)
	}
	if p.SourceType != "greenfield" {
		t.Errorf("source_type = %q, want 'greenfield'", p.SourceType)
	}
}

func TestCreateProject_Clone(t *testing.T) {
	base := newTestServer(t)
	payload := map[string]string{
		"name":        "hub",
		"source_type": "clone",
		"repo_url":    "git@github.com:org/hub.git",
	}
	resp := authedPost(t, base+"/api/projects", testToken, payload)
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, body)
	}
	var p struct {
		Registry string `json:"registry"`
	}
	decodeJSON(t, resp, &p)
	if p.Registry != "github" {
		t.Errorf("registry = %q, want 'github'", p.Registry)
	}
}

func TestCreateProject_InvalidName(t *testing.T) {
	base := newTestServer(t)
	resp := authedPost(t, base+"/api/projects", testToken, map[string]string{"name": "bad name", "source_type": "greenfield"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateProject_MissingSourceType(t *testing.T) {
	base := newTestServer(t)
	resp := authedPost(t, base+"/api/projects", testToken, map[string]string{"name": "ok-name"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateProject_CloneMissingURL(t *testing.T) {
	base := newTestServer(t)
	resp := authedPost(t, base+"/api/projects", testToken, map[string]string{"name": "myrepo", "source_type": "clone"})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestCreateProject_Duplicate(t *testing.T) {
	base := newTestServer(t)
	payload := map[string]string{"name": "once", "source_type": "greenfield"}
	authedPost(t, base+"/api/projects", testToken, payload)
	resp := authedPost(t, base+"/api/projects", testToken, payload)
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected 409, got %d", resp.StatusCode)
	}
}

func TestListProjects_Empty(t *testing.T) {
	base := newTestServer(t)
	resp := authedGet(t, base+"/api/projects", testToken)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	var list []any
	decodeJSON(t, resp, &list)
	if len(list) != 0 {
		t.Errorf("expected empty array, got %d items", len(list))
	}
}

func TestListProjects_ReturnsCreated(t *testing.T) {
	base := newTestServer(t)
	authedPost(t, base+"/api/projects", testToken, map[string]string{"name": "listed", "source_type": "greenfield"})

	resp := authedGet(t, base+"/api/projects", testToken)
	var list []struct {
		Name string `json:"name"`
	}
	decodeJSON(t, resp, &list)
	if len(list) != 1 {
		t.Fatalf("expected 1 project, got %d", len(list))
	}
	if list[0].Name != "listed" {
		t.Errorf("name = %q, want 'listed'", list[0].Name)
	}
}

// ── Sessions ──────────────────────────────────────────────────────────────────

func TestCreateSession_ProjectNotFound(t *testing.T) {
	base := newTestServer(t)
	resp := authedPost(t, base+"/api/projects/ghost/sessions", testToken, map[string]string{"session": "s1"})
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}

func TestCreateSession_Success(t *testing.T) {
	base := newTestServer(t)
	authedPost(t, base+"/api/projects", testToken, map[string]string{"name": "proj", "source_type": "greenfield"})

	resp := authedPost(t, base+"/api/projects/proj/sessions", testToken, map[string]string{
		"session": "s1",
		"cwd":     filepath.Join(t.TempDir(), "proj"),
	})
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, body)
	}
	var s struct {
		CanonicalName string `json:"canonical_name"`
	}
	decodeJSON(t, resp, &s)
	if s.CanonicalName != "proj/s1" {
		t.Errorf("canonical_name = %q, want 'proj/s1'", s.CanonicalName)
	}
}
