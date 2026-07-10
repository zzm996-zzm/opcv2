package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Error(c *gin.Context, status int, code string) {
	c.JSON(status, gin.H{"error": code})
}

func BadRequest(c *gin.Context, code string) {
	Error(c, http.StatusBadRequest, code)
}

func QueryLimit(c *gin.Context, defaultLimit, maxLimit int) (int, bool) {
	limit := defaultLimit
	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			BadRequest(c, "invalid_limit")
			return 0, false
		}
		limit = parsed
	}
	if maxLimit > 0 && limit > maxLimit {
		limit = maxLimit
	}
	return limit, true
}

func QueryOffset(c *gin.Context) (int, bool) {
	offset := 0
	if value := c.Query("offset"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			BadRequest(c, "invalid_offset")
			return 0, false
		}
		offset = parsed
	}
	return offset, true
}

func EnsureSlice[T any](values []T) []T {
	if values == nil {
		return []T{}
	}
	return values
}
