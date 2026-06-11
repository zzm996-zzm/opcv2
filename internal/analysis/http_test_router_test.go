package analysis

import (
	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
)

func testRouter(app Application) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewHTTPHandler(app)
	group := router.Group("/api/v1")
	group.Use(func(c *gin.Context) {
		c.Set(auth.UserIDContextKey, int64(42))
		c.Next()
	})
	handler.Register(group)
	return router
}
