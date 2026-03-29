package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"linxo-reader/internal/config"
	"linxo-reader/models"
)

// stubFetcher returns a fixed list of transactions for testing.
func stubFetcher(cfg *config.Config) ([]models.Transaction, error) {
	return []models.Transaction{
		{From: "TEST SHOP", Category: "Shopping", Amount: "-25.00", Date: "01/01/2026", Note: ""},
		{From: "SALARY", Category: "Income", Amount: "3000.00", Date: "02/01/2026", Note: "Jan"},
	}, nil
}

// failingFetcher always returns an error.
func failingFetcher(cfg *config.Config) ([]models.Transaction, error) {
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
