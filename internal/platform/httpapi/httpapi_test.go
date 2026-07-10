package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWriteErrorKeepsFlatErrorShape(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/error", func(c *gin.Context) {
		Error(c, http.StatusBadRequest, "invalid_request")
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/error", nil))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if strings.TrimSpace(recorder.Body.String()) != `{"error":"invalid_request"}` {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestQueryLimitDefaultsAndCapsLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	var defaulted int
	var capped int
	router.GET("/default", func(c *gin.Context) {
		var ok bool
		defaulted, ok = QueryLimit(c, 20, 100)
		if !ok {
			t.Fatal("QueryLimit returned false for empty query")
		}
		c.Status(http.StatusNoContent)
	})
	router.GET("/capped", func(c *gin.Context) {
		var ok bool
		capped, ok = QueryLimit(c, 20, 100)
		if !ok {
			t.Fatal("QueryLimit returned false for high query")
		}
		c.Status(http.StatusNoContent)
	})

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/default", nil))
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/capped?limit=500", nil))

	if defaulted != 20 {
		t.Fatalf("defaulted = %d, want 20", defaulted)
	}
	if capped != 100 {
		t.Fatalf("capped = %d, want 100", capped)
	}
}

func TestQueryLimitRejectsInvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/invalid", func(c *gin.Context) {
		if _, ok := QueryLimit(c, 20, 100); ok {
			t.Fatal("QueryLimit returned true for invalid query")
		}
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/invalid?limit=0", nil))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d body=%s", recorder.Code, recorder.Body.String())
	}
	if strings.TrimSpace(recorder.Body.String()) != `{"error":"invalid_limit"}` {
		t.Fatalf("body = %s", recorder.Body.String())
	}
}

func TestQueryOffsetDefaultsAndParsesOffset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	var defaulted int
	var parsed int
	router.GET("/default", func(c *gin.Context) {
		defaulted, _ = QueryOffset(c)
		c.Status(http.StatusNoContent)
	})
	router.GET("/parsed", func(c *gin.Context) {
		parsed, _ = QueryOffset(c)
		c.Status(http.StatusNoContent)
	})

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/default", nil))
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/parsed?offset=40", nil))

	if defaulted != 0 || parsed != 40 {
		t.Fatalf("defaulted/parsed = %d/%d", defaulted, parsed)
	}
}

func TestQueryOffsetRejectsNegativeOffset(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/invalid", func(c *gin.Context) {
		QueryOffset(c)
	})

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/invalid?offset=-1", nil))

	if recorder.Code != http.StatusBadRequest || strings.TrimSpace(recorder.Body.String()) != `{"error":"invalid_offset"}` {
		t.Fatalf("status/body = %d/%s", recorder.Code, recorder.Body.String())
	}
}

func TestEnsureSliceReturnsEmptySliceForNil(t *testing.T) {
	values := EnsureSlice[string](nil)

	if values == nil || len(values) != 0 {
		t.Fatalf("values = %#v, want non-nil empty slice", values)
	}
}
