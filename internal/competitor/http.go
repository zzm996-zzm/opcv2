package competitor

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/zzm/opcv2/internal/auth"
	"github.com/zzm/opcv2/internal/membership"
	"github.com/zzm/opcv2/internal/platform/httpapi"
)

type Application interface {
	CreateScan(ctx context.Context, input CreateScanInput) (Scan, error)
	ListScans(ctx context.Context, userID int64, limit int) ([]Scan, error)
	GetScan(ctx context.Context, userID, id int64) (Scan, error)
	RetryScan(ctx context.Context, userID, id int64) (Scan, error)
	CreateWatchItem(ctx context.Context, input CreateWatchItemInput) (WatchItem, error)
	DeleteWatchItem(ctx context.Context, userID, id int64) error
	StartWatchItemScan(ctx context.Context, userID, id int64) (Scan, error)
	GetMonitoring(ctx context.Context, userID int64, limit int) (MonitoringSnapshot, error)
}

type HTTPHandler struct {
	app Application
}

func NewHTTPHandler(app Application) *HTTPHandler {
	return &HTTPHandler{app: app}
}

func (h *HTTPHandler) Register(router *gin.RouterGroup) {
	router.POST("/competitor/scans", h.createScan)
	router.GET("/competitor/scans", h.listScans)
	router.GET("/competitor/scans/:id", h.getScan)
	router.POST("/competitor/scans/:id/retry", h.retryScan)
	router.POST("/competitor/monitoring/watchlist", h.createWatchItem)
	router.DELETE("/competitor/monitoring/watchlist/:id", h.deleteWatchItem)
	router.POST("/competitor/monitoring/watchlist/:id/scan", h.startWatchItemScan)
	router.GET("/competitor/monitoring", h.monitoring)
}

func (h *HTTPHandler) createScan(c *gin.Context) {
	var request CreateScanInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	if !validCreateScanInput(request) {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	scan, err := h.app.CreateScan(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, scan)
}

func validCreateScanInput(input CreateScanInput) bool {
	return len(normalizeStrings(input.Targets)) > 0 && strings.TrimSpace(input.Focus) != ""
}

func (h *HTTPHandler) listScans(c *gin.Context) {
	limit := queryLimit(c)
	if limit == 0 {
		return
	}
	scans, err := h.app.ListScans(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"scans": httpapi.EnsureSlice(scans)})
}

func (h *HTTPHandler) getScan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		httpapi.BadRequest(c, "invalid_scan_id")
		return
	}
	scan, err := h.app.GetScan(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, scan)
}

func (h *HTTPHandler) retryScan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		httpapi.BadRequest(c, "invalid_scan_id")
		return
	}
	scan, err := h.app.RetryScan(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, scan)
}

func (h *HTTPHandler) monitoring(c *gin.Context) {
	limit := queryLimit(c)
	if limit == 0 {
		return
	}
	snapshot, err := h.app.GetMonitoring(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), limit)
	if err != nil {
		writeError(c, err)
		return
	}
	snapshot.Watchlist = httpapi.EnsureSlice(snapshot.Watchlist)
	snapshot.Events = httpapi.EnsureSlice(snapshot.Events)
	c.JSON(http.StatusOK, snapshot)
}

func (h *HTTPHandler) createWatchItem(c *gin.Context) {
	var request CreateWatchItemInput
	if err := c.ShouldBindJSON(&request); err != nil {
		httpapi.BadRequest(c, "invalid_request")
		return
	}
	if strings.TrimSpace(request.Name) == "" {
		httpapi.BadRequest(c, "invalid_watch_item")
		return
	}
	request.UserID = c.GetInt64(auth.UserIDContextKey)
	item, err := h.app.CreateWatchItem(c.Request.Context(), request)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *HTTPHandler) deleteWatchItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		httpapi.BadRequest(c, "invalid_watch_item")
		return
	}
	if err := h.app.DeleteWatchItem(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func (h *HTTPHandler) startWatchItemScan(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		httpapi.BadRequest(c, "invalid_watch_item")
		return
	}
	scan, err := h.app.StartWatchItemScan(c.Request.Context(), c.GetInt64(auth.UserIDContextKey), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, scan)
}

func queryLimit(c *gin.Context) int {
	limit, ok := httpapi.QueryLimit(c, 20, 100)
	if !ok {
		return 0
	}
	return limit
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrScanNotFound):
		httpapi.Error(c, http.StatusNotFound, "scan_not_found")
	case errors.Is(err, ErrServiceNotReady):
		httpapi.Error(c, http.StatusInternalServerError, "service_not_ready")
	case errors.Is(err, ErrInvalidWatchItem):
		httpapi.BadRequest(c, "invalid_watch_item")
	case errors.Is(err, ErrWatchItemNotFound):
		httpapi.Error(c, http.StatusNotFound, "watch_item_not_found")
	case errors.Is(err, membership.ErrQuotaExceeded):
		httpapi.Error(c, http.StatusPaymentRequired, "quota_exceeded")
	case errors.Is(err, membership.ErrQuotaNotFound):
		httpapi.Error(c, http.StatusInternalServerError, "quota_not_configured")
	default:
		httpapi.Error(c, http.StatusInternalServerError, "internal_error")
	}
}
