package api

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"linxo-reader/internal/config"
	"linxo-reader/models"
)

// TransactionFetcher is the function signature used to retrieve transactions.
type TransactionFetcher func(cfg *config.Config) ([]models.Transaction, error)

// Server is the HTTP API server.
type Server struct {
	cfg     *config.Config
	router  *gin.Engine
	fetcher TransactionFetcher
}

// NewServer creates and configures the HTTP server.
func NewServer(cfg *config.Config, fetcher TransactionFetcher) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(apiKeyAuth(cfg.APIKey))

	s := &Server{cfg: cfg, router: r, fetcher: fetcher}
	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	s.router.GET("/showbanks", s.handleShowBanks)
}

// ServeHTTP exposes the router as an http.Handler for testing.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Run starts listening on the given address.
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) handleShowBanks(c *gin.Context) {
	items, err := s.fetcher(s.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Returning %d transactions", len(items))
	c.JSON(http.StatusOK, items)
}

func apiKeyAuth(validKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-Api-Key")
		if key == "" {
			auth := c.GetHeader("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				key = strings.TrimPrefix(auth, "Bearer ")
			}
		}
		if key != validKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
