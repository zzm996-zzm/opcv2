package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/membership"
)

type HealthChecks struct {
	Ready func() bool
}

func NewRouter(
	checks HealthChecks,
	authHandler *auth.HTTPHandler,
	membershipHandler *membership.HTTPHandler,
) http.Handler {
	router := gin.New()
	router.Use(gin.Recovery())

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
	if authHandler != nil && membershipHandler != nil {
		protected := api.Group("")
		protected.Use(authHandler.RequireAccessToken())
		membershipHandler.Register(protected)
	}

	return router
}
