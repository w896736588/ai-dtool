package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRecorderTokenAuth_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RecorderTokenAuthMiddleware())
	r.POST("/api/e2e/record/by_token/x", func(c *gin.Context) { c.String(200, "ok") })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("POST", "/api/e2e/record/by_token/x", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("期望 401，实际 %d", w.Code)
	}
}