package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthChecks struct {
	Ready func() bool
}

func NewRouter(checks HealthChecks) http.Handler {
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

	return router
}
