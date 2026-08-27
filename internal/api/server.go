package api

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"linxo-reader/internal/config"
	"linxo-reader/models"
)

// TransactionFetcher is the function signature used to retrieve transactions.
type TransactionFetcher func(ctx context.Context) ([]models.Transaction, error)

// Server is the HTTP API server.
type Server struct {
	cfg     *config.Config
	router  *gin.Engine
	fetcher TransactionFetcher
	cache   *responseCache
	fetchMu sync.Mutex
}

type cachedResponse struct {
	data      []byte
	createdAt time.Time
}

type responseCache struct {
	ttl     time.Duration
	mu      sync.RWMutex
	entries map[string]cachedResponse
}

func newResponseCache(ttl time.Duration) *responseCache {
	return &responseCache{ttl: ttl, entries: make(map[string]cachedResponse)}
}

func (c *responseCache) get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[key]
	if !ok || time.Since(e.createdAt) > c.ttl {
		return nil, false
	}
	return e.data, true
}

func (c *responseCache) set(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cachedResponse{data: data, createdAt: time.Now()}
}

// NewServer creates and configures the HTTP server.
func NewServer(cfg *config.Config, fetcher TransactionFetcher) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(apiKeyAuth(cfg.APIKey))

	s := &Server{
		cfg:     cfg,
		router:  r,
		fetcher: fetcher,
		cache:   newResponseCache(10 * time.Minute),
	}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.router.GET("/showbanks", s.handleShowBanks)
	s.router.GET("/health", s.handleHealth)
}

// ServeHTTP exposes the router as an http.Handler for testing.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Run starts an HTTP server with sensible timeouts.
func (s *Server) Run(addr string) error {
	srv := &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 5*time.Minute + 30*time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Listening on %s", addr)
	return srv.ListenAndServe()
}

const cacheKeyTransactions = "transactions"

func (s *Server) handleShowBanks(c *gin.Context) {
	if data, ok := s.cache.get(cacheKeyTransactions); ok {
		log.Printf("Returning %d transactions (cached)", countTransactions(data))
		c.Data(http.StatusOK, "application/json", data)
		return
	}

	// Serialize fetches so only one browser session runs at a time.
	s.fetchMu.Lock()
	defer s.fetchMu.Unlock()

	// Double-check after acquiring the lock.
	if data, ok := s.cache.get(cacheKeyTransactions); ok {
		log.Printf("Returning %d transactions (cached)", countTransactions(data))
		c.Data(http.StatusOK, "application/json", data)
		return
	}

	items, err := s.fetcher(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	data, err := json.Marshal(items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "marshal error"})
		return
	}

	s.cache.set(cacheKeyTransactions, data)
	log.Printf("Returning %d transactions", len(items))
	c.Data(http.StatusOK, "application/json", data)
}

func (s *Server) handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func apiKeyAuth(validKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}
		key := c.GetHeader("X-Api-Key")
		if key == "" {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				key = strings.TrimPrefix(auth, "Bearer ")
			}
		}
		if subtle.ConstantTimeCompare([]byte(key), []byte(validKey)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

func countTransactions(data []byte) int {
	var items []models.Transaction
	if err := json.Unmarshal(data, &items); err != nil {
		return 0
	}
	return len(items)
}
