package auth_test

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pecodez/couchdev/internal/auth"
)

func tokenHash(secret string) []byte {
	sum := sha256.Sum256([]byte(secret))
	return sum[:]
}

func tokenHashHex(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func TestParseTokenHash_Valid(t *testing.T) {
	h, err := auth.ParseTokenHash(tokenHashHex("secret"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(h) != 32 {
		t.Fatalf("expected 32 bytes, got %d", len(h))
	}
}

func TestParseTokenHash_InvalidHex(t *testing.T) {
	_, err := auth.ParseTokenHash("not-hex!!!")
	if err == nil {
		t.Fatal("expected error for non-hex input, got nil")
	}
}

func TestParseTokenHash_WrongLength(t *testing.T) {
	// 31 bytes = 62 hex chars
	short := hex.EncodeToString(make([]byte, 31))
	_, err := auth.ParseTokenHash(short)
	if err == nil {
		t.Fatal("expected error for wrong-length hash, got nil")
	}
}

func makeMiddleware(secret string) func(http.Handler) http.Handler {
	return auth.Middleware(tokenHash(secret))
}

func callMiddleware(t *testing.T, secret, authHeader string) *httptest.ResponseRecorder {
	t.Helper()
	reached := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	})
	mw := makeMiddleware(secret)(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	w := httptest.NewRecorder()
	mw.ServeHTTP(w, req)
	_ = reached
	return w
}

func TestMiddleware_ValidToken(t *testing.T) {
	w := callMiddleware(t, "secret", "Bearer secret")
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestMiddleware_MissingHeader(t *testing.T) {
	w := callMiddleware(t, "secret", "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestMiddleware_WrongToken(t *testing.T) {
	w := callMiddleware(t, "secret", "Bearer wrong")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestMiddleware_MalformedBearer(t *testing.T) {
	w := callMiddleware(t, "secret", "Basic abc")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
