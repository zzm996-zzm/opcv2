package httpserver

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/account"
	"github.com/zzm/opcv2/internal/analysis"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/competitor"
	"github.com/zzm/opcv2/internal/content"
	"github.com/zzm/opcv2/internal/copilot"
	"github.com/zzm/opcv2/internal/crm"
	"github.com/zzm/opcv2/internal/dashboard"
	"github.com/zzm/opcv2/internal/enterprise"
	"github.com/zzm/opcv2/internal/geo"
	"github.com/zzm/opcv2/internal/growth"
	"github.com/zzm/opcv2/internal/home"
	"github.com/zzm/opcv2/internal/leads"
	"github.com/zzm/opcv2/internal/learning"
	"github.com/zzm/opcv2/internal/membership"
	"github.com/zzm/opcv2/internal/notifications"
	"github.com/zzm/opcv2/internal/projects"
	"github.com/zzm/opcv2/internal/sandbox"
	"github.com/zzm/opcv2/internal/support"
	"github.com/zzm/opcv2/internal/tasks"
)

type HealthChecks struct {
	Ready                  func() bool
	Logger                 *slog.Logger
	AllowedOrigins         []string
	ExpensiveEndpointLimit int
}

type Handlers struct {
	Auth          *auth.HTTPHandler
	Account       *account.HTTPHandler
	Notifications *notifications.HTTPHandler
	Home          *home.HTTPHandler
	Membership    *membership.HTTPHandler
	Analysis      *analysis.HTTPHandler
	Projects      *projects.HTTPHandler
	Leads         *leads.HTTPHandler
	CRM           *crm.HTTPHandler
	Content       *content.HTTPHandler
	Support       *support.HTTPHandler
	Sandbox       *sandbox.HTTPHandler
	Tasks         *tasks.HTTPHandler
	Dashboard     *dashboard.HTTPHandler
	Geo           *geo.HTTPHandler
	Enterprise    *enterprise.HTTPHandler
	Growth        *growth.HTTPHandler
	Competitor    *competitor.HTTPHandler
	Learning      *learning.HTTPHandler
	Copilot       *copilot.HTTPHandler
}

func NewRouter(checks HealthChecks, handlers Handlers) http.Handler {
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
	if handlers.Auth != nil {
		handlers.Auth.Register(api)
	}
	if handlers.Content != nil {
		handlers.Content.RegisterPublic(api)
	}
	if handlers.Learning != nil {
		handlers.Learning.RegisterPublic(api)
	}
	if handlers.Enterprise != nil {
		handlers.Enterprise.RegisterPublic(api)
	}
	if handlers.Auth != nil && handlers.Account != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Account.Register(protected)
	}
	if handlers.Auth != nil && handlers.Notifications != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Notifications.Register(protected)
	}
	if handlers.Auth != nil && handlers.Home != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Home.Register(protected)
	}
	if handlers.Auth != nil && handlers.Membership != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Membership.Register(protected)
	}
	if handlers.Auth != nil && handlers.Analysis != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Analysis.Register(protected)
	}
	if handlers.Projects != nil {
		handlers.Projects.RegisterPublic(api)
	}
	if handlers.Auth != nil && handlers.Projects != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Projects.RegisterProtected(protected)
	}
	if handlers.Auth != nil && handlers.Leads != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Leads.Register(protected)
	}
	if handlers.Auth != nil && handlers.CRM != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.CRM.Register(protected)
	}
	if handlers.Auth != nil && handlers.Content != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Content.RegisterProtected(protected)
		handlers.Content.RegisterAdmin(protected)
	}
	if handlers.Auth != nil && handlers.Support != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Support.Register(protected)
	}
	if handlers.Auth != nil && handlers.Sandbox != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Sandbox.Register(protected)
	}
	if handlers.Auth != nil && handlers.Tasks != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Tasks.Register(protected)
	}
	if handlers.Auth != nil && handlers.Dashboard != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Dashboard.Register(protected)
	}
	if handlers.Auth != nil && handlers.Geo != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Geo.Register(protected)
	}
	if handlers.Auth != nil && handlers.Enterprise != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Enterprise.Register(protected)
	}
	if handlers.Auth != nil && handlers.Growth != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Growth.Register(protected)
	}
	if handlers.Auth != nil && handlers.Competitor != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Competitor.Register(protected)
	}
	if handlers.Auth != nil && handlers.Learning != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Learning.RegisterProtected(protected)
	}
	if handlers.Auth != nil && handlers.Copilot != nil {
		protected := api.Group("")
		protected.Use(handlers.Auth.RequireAccessToken())
		handlers.Copilot.Register(protected)
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
			header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
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
		"/api/v1/project-matches",
		"/api/v1/project-matches/:id/answer",
		"/api/v1/project-matches/:id/generate",
		"/api/v1/projects/matches",
		"/api/v1/leads/tasks",
		"/api/v1/crm/customers/:id/follow-up-copy",
		"/api/v1/sandbox/sessions/:id/run",
		"/api/v1/copilot/threads/:id/messages",
		"/api/v1/copilot/threads/:id/messages/stream":
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
