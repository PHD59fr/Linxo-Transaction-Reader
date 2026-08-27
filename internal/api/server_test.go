package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"linxo-reader/internal/config"
	"linxo-reader/models"
)

// stubFetcher returns a fixed list of transactions for testing.
func stubFetcher(ctx context.Context) ([]models.Transaction, error) {
	return []models.Transaction{
		{From: "TEST SHOP", Category: "Shopping", Amount: "-25.00", Date: "01/01/2026", Note: ""},
		{From: "SALARY", Category: "Income", Amount: "3000.00", Date: "02/01/2026", Note: "Jan"},
	}, nil
}

// failingFetcher always returns an error.
func failingFetcher(ctx context.Context) ([]models.Transaction, error) {
	return nil, fmt.Errorf("fetch failed: connection timeout")
}

func newTestServer(apiKey string, fetcher TransactionFetcher) *Server {
	cfg := &config.Config{
		APIKey: apiKey,
		Port:   "0",
	}
	if fetcher == nil {
		fetcher = stubFetcher
	}
	return NewServer(cfg, fetcher)
}

// apiKeyAuth tests ----------------------------------------------------------

func TestApiKeyAuth_MissingKey(t *testing.T) {
	srv := newTestServer("secret123", nil)

	req := httptest.NewRequest(http.MethodGet, "/showbanks", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestApiKeyAuth_WrongKey(t *testing.T) {
	srv := newTestServer("secret123", nil)

	req := httptest.NewRequest(http.MethodGet, "/showbanks", nil)
	req.Header.Set("X-Api-Key", "wrong-key")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestApiKeyAuth_ValidXAPIKey(t *testing.T) {
	srv := newTestServer("secret123", nil)

	req := httptest.NewRequest(http.MethodGet, "/showbanks", nil)
	req.Header.Set("X-Api-Key", "secret123")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestApiKeyAuth_ValidBearerToken(t *testing.T) {
	srv := newTestServer("secret123", nil)

	req := httptest.NewRequest(http.MethodGet, "/showbanks", nil)
	req.Header.Set("Authorization", "Bearer secret123")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestApiKeyAuth_EmptyBearerToken(t *testing.T) {
	srv := newTestServer("secret123", nil)

	req := httptest.NewRequest(http.MethodGet, "/showbanks", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestApiKeyAuth_MalformedAuthHeader(t *testing.T) {
	srv := newTestServer("secret123", nil)

	req := httptest.NewRequest(http.MethodGet, "/showbanks", nil)
	req.Header.Set("Authorization", "Basic abc123")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// /showbanks endpoint tests -------------------------------------------------

func TestShowBanks_Success(t *testing.T) {
	srv := newTestServer("secret123", stubFetcher)

	req := httptest.NewRequest(http.MethodGet, "/showbanks", nil)
	req.Header.Set("X-Api-Key", "secret123")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var got []models.Transaction
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(got))
	}
	if got[0].From != "TEST SHOP" {
		t.Errorf("expected From=%q, got %q", "TEST SHOP", got[0].From)
	}
	if got[1].Amount != "3000.00" {
		t.Errorf("expected Amount=%q, got %q", "3000.00", got[1].Amount)
	}
}

func TestShowBanks_FetchError(t *testing.T) {
	srv := newTestServer("secret123", failingFetcher)

	req := httptest.NewRequest(http.MethodGet, "/showbanks", nil)
	req.Header.Set("X-Api-Key", "secret123")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["error"] == "" {
		t.Error("expected error message in response body")
	}
}

func TestShowBanks_Unauthorized(t *testing.T) {
	srv := newTestServer("secret123", stubFetcher)

	req := httptest.NewRequest(http.MethodGet, "/showbanks", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestShowBanks_UnknownRoute(t *testing.T) {
	srv := newTestServer("secret123", stubFetcher)

	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	req.Header.Set("X-Api-Key", "secret123")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestShowBanks_MethodNotAllowed(t *testing.T) {
	srv := newTestServer("secret123", stubFetcher)

	req := httptest.NewRequest(http.MethodPost, "/showbanks", nil)
	req.Header.Set("X-Api-Key", "secret123")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// Health check tests --------------------------------------------------------

func TestHealth_Success(t *testing.T) {
	srv := newTestServer("secret123", nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", body["status"])
	}
}

func TestHealth_NoAuthRequired(t *testing.T) {
	srv := newTestServer("secret123", nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 without auth, got %d", w.Code)
	}
}

// Cache tests ---------------------------------------------------------------

func TestShowBanks_CacheHit(t *testing.T) {
	callCount := 0
	countingFetcher := func(ctx context.Context) ([]models.Transaction, error) {
		callCount++
		return stubFetcher(ctx)
	}
	srv := newTestServer("secret123", countingFetcher)

	// First call: fetches from upstream.
	req := httptest.NewRequest(http.MethodGet, "/showbanks", nil)
	req.Header.Set("X-Api-Key", "secret123")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("first call: expected 200, got %d", w.Code)
	}
	if callCount != 1 {
		t.Fatalf("expected 1 fetch, got %d", callCount)
	}

	// Second call: should hit cache.
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/showbanks", nil)
	req2.Header.Set("X-Api-Key", "secret123")
	srv.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("second call: expected 200, got %d", w2.Code)
	}
	if callCount != 1 {
		t.Fatalf("expected still 1 fetch (cache hit), got %d", callCount)
	}
}
