package projects

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
)

type Application interface {
	ListOpportunities(ctx context.Context, filters OpportunityFilters) ([]Opportunity, error)
	GetOpportunity(ctx context.Context, slug string) (Opportunity, error)
	ListCases(ctx context.Context, filters CaseFilters) ([]CaseStudy, error)
	GetCase(ctx context.Context, slug string) (CaseStudy, error)
	CreateMatch(ctx context.Context, input MatchInput) (MatchResult, error)
	ListMatches(ctx context.Context, userID int64, limit int) ([]MatchSession, error)
	GetMatch(ctx context.Context, userID, id int64) (MatchSession, error)
	AnswerMatch(ctx context.Context, input AnswerMatchInput) (MatchResult, error)
	CreateComparison(ctx context.Context, input CreateComparisonInput) (Comparison, error)
	GetComparison(ctx context.Context, userID, id int64) (Comparison, error)
	FavoriteMatch(ctx context.Context, userID, id int64) (Favorite, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.GET("/projects/opportunities", h.listOpportunities)
	router.GET("/projects/opportunities/:slug", h.getOpportunity)
	router.GET("/projects/cases", h.listCases)
	router.GET("/projects/cases/:slug", h.getCase)
	router.POST("/projects/matches", h.createMatch)
	router.GET("/projects/matches", h.listMatches)
	router.GET("/projects/matches/:id", h.getMatch)
	router.POST("/projects/matches/:id/answers", h.answerMatch)
	router.POST("/projects/matches/:id/favorite", h.favoriteMatch)
	router.POST("/projects/comparisons", h.createComparison)
	router.GET("/projects/comparisons/:id", h.getComparison)
}

func (h *HTTPHandler) createComparison(c *gin.Context) {
	var request CreateComparisonInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	result, err := h.app.CreateComparison(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
func (h *HTTPHandler) getComparison(c *gin.Context) {
	id, ok := matchID(c)
	if !ok {
		return
	}
	result, err := h.app.GetComparison(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) answerMatch(c *gin.Context) {
	id, ok := matchID(c)
	if !ok {
		return
	}
	var request AnswerMatchInput
	if err := c.ShouldBindJSON(&request); err != nil || len(request.Answers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	request.SessionID = id
	result, err := h.app.AnswerMatch(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) listCases(c *gin.Context) {
	items, err := h.app.ListCases(c.Request.Context(), CaseFilters{CaseType: c.Query("type"), Limit: 20})
	if err != nil {
		writeError(c, err)
		return
	}
	if items == nil {
		items = []CaseStudy{}
	}
	c.JSON(http.StatusOK, gin.H{"cases": items})
}
func (h *HTTPHandler) getCase(c *gin.Context) {
	item, err := h.app.GetCase(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *HTTPHandler) listOpportunities(c *gin.Context) {
	limit := 20
	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_limit"})
			return
		}
		limit = parsed
	}
	items, err := h.app.ListOpportunities(c.Request.Context(), OpportunityFilters{Query: c.Query("q"), Industry: c.Query("industry"), Limit: limit})
	if err != nil {
		writeError(c, err)
		return
	}
	if items == nil {
		items = []Opportunity{}
	}
	c.JSON(http.StatusOK, gin.H{"opportunities": items})
}

func (h *HTTPHandler) getOpportunity(c *gin.Context) {
	item, err := h.app.GetOpportunity(c.Request.Context(), c.Param("slug"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *HTTPHandler) createMatch(c *gin.Context) {
	var request MatchInput
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	result, err := h.app.CreateMatch(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *HTTPHandler) listMatches(c *gin.Context) {
	limit := 20
	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_limit"})
			return
		}
		limit = parsed
	}
	sessions, err := h.app.ListMatches(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"matches": sessions})
}

func (h *HTTPHandler) getMatch(c *gin.Context) {
	id, ok := matchID(c)
	if !ok {
		return
	}
	session, err := h.app.GetMatch(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *HTTPHandler) favoriteMatch(c *gin.Context) {
	id, ok := matchID(c)
	if !ok {
		return
	}
	favorite, err := h.app.FavoriteMatch(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, favorite)
}

func matchID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_match_id"})
		return 0, false
	}
	return id, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrSessionNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "match_not_found"})
	case errors.Is(err, ErrOpportunityNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "opportunity_not_found"})
	case errors.Is(err, ErrCaseNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "case_not_found"})
	case errors.Is(err, ErrInvalidMatchAnswers):
		c.JSON(http.StatusConflict, gin.H{"error": "invalid_match_answers"})
	case errors.Is(err, ErrComparisonNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "comparison_not_found"})
	case errors.Is(err, ErrInvalidComparison):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_comparison"})
	case errors.Is(err, ErrServiceNotReady):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service_not_ready"})
	case errors.Is(err, ErrInvalidAIResult):
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid_ai_result"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
	}
}
