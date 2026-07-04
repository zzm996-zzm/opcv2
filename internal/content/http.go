package content

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
)

type Application interface {
	ListArticles(ctx context.Context) ([]Article, error)
	GetArticle(ctx context.Context, slug string) (Article, error)
	CreateArticle(ctx context.Context, userID int64, input ArticleInput) (Article, error)
	BookmarkArticle(ctx context.Context, userID int64, slug string) (BookmarkResult, error)
	UnbookmarkArticle(ctx context.Context, userID int64, slug string) (BookmarkResult, error)
	ListTools(ctx context.Context, filters ToolFilters) ([]Tool, error)
	GetTool(ctx context.Context, slug string) (Tool, error)
	FavoriteTool(ctx context.Context, userID int64, slug string) (FavoriteResult, error)
	UnfavoriteTool(ctx context.Context, userID int64, slug string) (FavoriteResult, error)
	UpsertTool(ctx context.Context, userID int64, input ToolInput) (Tool, error)
	GetCommunityConfig(ctx context.Context) (CommunityConfig, error)
	UpdateCommunityConfig(ctx context.Context, userID int64, input CommunityConfigInput) (CommunityConfig, error)
	CreateCommunityJoinRequest(ctx context.Context, userID int64, input CommunityJoinInput) (CommunityJoinRequest, error)
	ListHelpTopics(ctx context.Context) ([]HelpTopic, error)
	ListHelpArticles(ctx context.Context, filters HelpArticleFilters) ([]HelpArticle, error)
	GetHelpArticle(ctx context.Context, slug string) (HelpArticle, error)
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
	router.GET("/content/tools/:slug", h.getTool)
	router.GET("/content/community", h.getCommunityConfig)
	router.GET("/content/brand", h.getBrand)
	router.GET("/help/topics", h.listHelpTopics)
	router.GET("/help/articles", h.listHelpArticles)
	router.GET("/help/articles/:slug", h.getHelpArticle)
}

func (h *HTTPHandler) RegisterProtected(router *gin.RouterGroup) {
	router.POST("/content/articles/:slug/bookmark", h.bookmarkArticle)
	router.DELETE("/content/articles/:slug/bookmark", h.unbookmarkArticle)
	router.POST("/content/tools/:slug/favorite", h.favoriteTool)
	router.DELETE("/content/tools/:slug/favorite", h.unfavoriteTool)
	router.POST("/community/join-requests", h.createCommunityJoinRequest)
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

func (h *HTTPHandler) bookmarkArticle(c *gin.Context) {
	result, err := h.app.BookmarkArticle(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) unbookmarkArticle(c *gin.Context) {
	result, err := h.app.UnbookmarkArticle(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) listTools(c *gin.Context) {
	limit := queryLimit(c)
	if limit == 0 {
		return
	}
	tools, err := h.app.ListTools(c.Request.Context(), ToolFilters{
		Category: c.Query("category"),
		Query:    c.Query("q"),
		Sort:     c.Query("sort"),
		Limit:    limit,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	if tools == nil {
		tools = []Tool{}
	}
	c.JSON(http.StatusOK, gin.H{"tools": tools})
}

func (h *HTTPHandler) getTool(c *gin.Context) {
	tool, err := h.app.GetTool(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, tool)
}

func (h *HTTPHandler) favoriteTool(c *gin.Context) {
	result, err := h.app.FavoriteTool(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) unfavoriteTool(c *gin.Context) {
	result, err := h.app.UnfavoriteTool(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
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

func (h *HTTPHandler) createCommunityJoinRequest(c *gin.Context) {
	var request CommunityJoinInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	result, err := h.app.CreateCommunityJoinRequest(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) listHelpTopics(c *gin.Context) {
	topics, err := h.app.ListHelpTopics(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	if topics == nil {
		topics = []HelpTopic{}
	}
	c.JSON(http.StatusOK, gin.H{"topics": topics})
}

func (h *HTTPHandler) listHelpArticles(c *gin.Context) {
	limit := queryLimit(c)
	if limit == 0 {
		return
	}
	articles, err := h.app.ListHelpArticles(c.Request.Context(), HelpArticleFilters{
		Topic: c.Query("topic"),
		Query: c.Query("q"),
		Limit: limit,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	if articles == nil {
		articles = []HelpArticle{}
	}
	c.JSON(http.StatusOK, gin.H{"articles": articles})
}

func (h *HTTPHandler) getHelpArticle(c *gin.Context) {
	article, err := h.app.GetHelpArticle(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, article)
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
	case errors.Is(err, ErrHelpArticleNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "help_article_not_found"})
	case errors.Is(err, ErrServiceNotReady):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}

func queryLimit(c *gin.Context) int {
	limit := 20
	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_limit"})
			return 0
		}
		limit = parsed
	}
	if limit > 100 {
		return 100
	}
	return limit
}
