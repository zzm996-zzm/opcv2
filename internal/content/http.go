package content

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
)

type Application interface {
	ListArticles(ctx context.Context) ([]Article, error)
	GetArticle(ctx context.Context, slug string) (Article, error)
	CreateArticle(ctx context.Context, userID int64, input ArticleInput) (Article, error)
	ListTools(ctx context.Context) ([]Tool, error)
	UpsertTool(ctx context.Context, userID int64, input ToolInput) (Tool, error)
	GetCommunityConfig(ctx context.Context) (CommunityConfig, error)
	UpdateCommunityConfig(ctx context.Context, userID int64, input CommunityConfigInput) (CommunityConfig, error)
	GetBrand(ctx context.Context) (BrandContent, error)
	UpsertBrandMetric(ctx context.Context, userID int64, input BrandMetricInput) (BrandMetric, error)
	UpsertBrandCase(ctx context.Context, userID int64, input BrandCaseInput) (BrandCase, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) RegisterPublic(router *gin.RouterGroup) {
	router.GET("/content/articles", h.listArticles)
	router.GET("/content/articles/:slug", h.getArticle)
	router.GET("/content/tools", h.listTools)
	router.GET("/content/community", h.getCommunityConfig)
	router.GET("/content/brand", h.getBrand)
}

func (h *HTTPHandler) RegisterAdmin(router *gin.RouterGroup) {
	router.POST("/admin/content/articles", h.createArticle)
	router.POST("/admin/content/tools", h.upsertTool)
	router.PUT("/admin/content/community", h.updateCommunityConfig)
	router.POST("/admin/content/brand/metrics", h.upsertBrandMetric)
	router.POST("/admin/content/brand/cases", h.upsertBrandCase)
}

func (h *HTTPHandler) listArticles(c *gin.Context) {
	articles, err := h.app.ListArticles(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	if articles == nil {
		articles = []Article{}
	}
	c.JSON(http.StatusOK, gin.H{"articles": articles})
}

func (h *HTTPHandler) getArticle(c *gin.Context) {
	article, err := h.app.GetArticle(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, article)
}

func (h *HTTPHandler) createArticle(c *gin.Context) {
	var request ArticleInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	article, err := h.app.CreateArticle(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, article)
}

func (h *HTTPHandler) listTools(c *gin.Context) {
	tools, err := h.app.ListTools(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	if tools == nil {
		tools = []Tool{}
	}
	c.JSON(http.StatusOK, gin.H{"tools": tools})
}

func (h *HTTPHandler) upsertTool(c *gin.Context) {
	var request ToolInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	tool, err := h.app.UpsertTool(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, tool)
}

func (h *HTTPHandler) getCommunityConfig(c *gin.Context) {
	config, err := h.app.GetCommunityConfig(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, config)
}

func (h *HTTPHandler) updateCommunityConfig(c *gin.Context) {
	var request CommunityConfigInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	config, err := h.app.UpdateCommunityConfig(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, config)
}

func (h *HTTPHandler) getBrand(c *gin.Context) {
	brand, err := h.app.GetBrand(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	if brand.Metrics == nil {
		brand.Metrics = []BrandMetric{}
	}
	if brand.Cases == nil {
		brand.Cases = []BrandCase{}
	}
	c.JSON(http.StatusOK, brand)
}

func (h *HTTPHandler) upsertBrandMetric(c *gin.Context) {
	var request BrandMetricInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	metric, err := h.app.UpsertBrandMetric(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, metric)
}

func (h *HTTPHandler) upsertBrandCase(c *gin.Context) {
	var request BrandCaseInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	brandCase, err := h.app.UpsertBrandCase(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, brandCase)
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_content_input"})
	case errors.Is(err, ErrAdminRequired):
		c.JSON(http.StatusForbidden, gin.H{"error": "admin_required"})
	case errors.Is(err, ErrArticleNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "article_not_found"})
	case errors.Is(err, ErrToolNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "tool_not_found"})
	case errors.Is(err, ErrServiceNotReady):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}
