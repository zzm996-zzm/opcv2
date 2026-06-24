package httpserver

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/analysis"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/content"
	"github.com/zzm/opcv2/internal/crm"
	"github.com/zzm/opcv2/internal/leads"
	"github.com/zzm/opcv2/internal/membership"
	"github.com/zzm/opcv2/internal/projects"
)

type HealthChecks struct {
	Ready                  func() bool
	Logger                 *slog.Logger
	AllowedOrigins         []string
	ExpensiveEndpointLimit int
}

func NewRouter(
	checks HealthChecks,
	authHandler *auth.HTTPHandler,
	membershipHandler *membership.HTTPHandler,
	analysisHandler *analysis.HTTPHandler,
	projectsHandler *projects.HTTPHandler,
	leadsHandler *leads.HTTPHandler,
	crmHandler *crm.HTTPHandler,
	contentHandler *content.HTTPHandler,
) http.Handler {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(corsMiddleware(checks.AllowedOrigins))
	if checks.Logger != nil {
		router.Use(requestLogger(checks.Logger))
	}
	if checks.ExpensiveEndpointLimit > 0 {
		router.Use(expensiveEndpointLimiter(checks.ExpensiveEndpointLimit, time.Minute))
	}

	router.GET("/health/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/health/ready", func(c *gin.Context) {
		if checks.Ready != nil && !checks.Ready() {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	api := router.Group("/api/v1")
	if authHandler != nil {
		authHandler.Register(api)
	}
	if contentHandler != nil {
		contentHandler.RegisterPublic(api)
	}
	if authHandler != nil && membershipHandler != nil {
		protected := api.Group("")
		protected.Use(authHandler.RequireAccessToken())
		membershipHandler.Register(protected)
	}
	if authHandler != nil && analysisHandler != nil {
		protected := api.Group("")
		protected.Use(authHandler.RequireAccessToken())
		analysisHandler.Register(protected)
	}
	if authHandler != nil && projectsHandler != nil {
		protected := api.Group("")
		protected.Use(authHandler.RequireAccessToken())
		projectsHandler.Register(protected)
	}
	if authHandler != nil && leadsHandler != nil {
		protected := api.Group("")
		protected.Use(authHandler.RequireAccessToken())
		leadsHandler.Register(protected)
	}
	if authHandler != nil && crmHandler != nil {
		protected := api.Group("")
		protected.Use(authHandler.RequireAccessToken())
		crmHandler.Register(protected)
	}
	if authHandler != nil && contentHandler != nil {
		protected := api.Group("")
		protected.Use(authHandler.RequireAccessToken())
		contentHandler.RegisterAdmin(protected)
	}

	return router
}

func corsMiddleware(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = true
		}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowed[origin] {
			header := c.Writer.Header()
			header.Set("Access-Control-Allow-Origin", origin)
			header.Set("Vary", "Origin")
			header.Set("Access-Control-Allow-Credentials", "true")
			header.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("request_complete",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"user_id", c.GetInt64(auth.UserIDContextKey),
		)
	}
}

type fixedWindowLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	buckets map[string]rateBucket
}

type rateBucket struct {
	count   int
	resetAt time.Time
}

func expensiveEndpointLimiter(limit int, window time.Duration) gin.HandlerFunc {
	limiter := &fixedWindowLimiter{
		limit:   limit,
		window:  window,
		buckets: map[string]rateBucket{},
	}
	return func(c *gin.Context) {
		if !isExpensiveEndpoint(c.Request.Method, c.FullPath(), c.Request.URL.Path) {
			c.Next()
			return
		}
		key := rateLimitKey(c)
		if !limiter.allow(key, time.Now()) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limited",
				"message": "Too many expensive requests. Please try again later.",
			})
			return
		}
		c.Next()
	}
}

func (l *fixedWindowLimiter) allow(key string, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	bucket := l.buckets[key]
	if bucket.resetAt.IsZero() || !now.Before(bucket.resetAt) {
		l.buckets[key] = rateBucket{count: 1, resetAt: now.Add(l.window)}
		return true
	}
	if bucket.count >= l.limit {
		return false
	}
	bucket.count++
	l.buckets[key] = bucket
	return true
}

func isExpensiveEndpoint(method, fullPath, requestPath string) bool {
	if method != http.MethodPost {
		return false
	}
	path := fullPath
	if path == "" {
		path = requestPath
	}
	switch path {
	case "/api/v1/analysis/direction",
		"/api/v1/projects/matches",
		"/api/v1/leads/tasks",
		"/api/v1/crm/customers/:id/follow-up-copy":
		return true
	default:
		return false
	}
}

func rateLimitKey(c *gin.Context) string {
	userID := c.GetInt64(auth.UserIDContextKey)
	if userID > 0 {
		return c.FullPath() + ":user:" + strconv.FormatInt(userID, 10)
	}
	return c.FullPath() + ":ip:" + c.ClientIP()
}
