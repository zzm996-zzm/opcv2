package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
)

type HealthChecks struct {
	Ready func() bool
}

func NewRouter(checks HealthChecks, authHandlers ...*auth.HTTPHandler) http.Handler {
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
	if len(authHandlers) > 0 && authHandlers[0] != nil {
		authHandlers[0].Register(router.Group("/api/v1"))
	}

	return router
}
