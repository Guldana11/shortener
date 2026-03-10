package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Guldana11/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

func TestUserCookieMiddleware_NoCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotUserID string
	r := gin.New()
	r.Use(UserCookieMiddleware())
	r.GET("/test", func(c *gin.Context) {
		uid, _ := c.Get("userID")
		gotUserID = uid.(string)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotUserID == "" {
		t.Error("userID не установлен в контексте")
	}

	resp := rec.Result()
	defer resp.Body.Close()
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Error("cookie должен быть установлен для нового пользователя")
	}
}

func TestUserCookieMiddleware_ValidCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cookie := service.GenerateUserCookie()
	expectedUserID := cookie.Value[:36]

	var gotUserID string
	r := gin.New()
	r.Use(UserCookieMiddleware())
	r.GET("/test", func(c *gin.Context) {
		uid, _ := c.Get("userID")
		gotUserID = uid.(string)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotUserID != expectedUserID {
		t.Errorf("expected userID %s, got %s", expectedUserID, gotUserID)
	}
}

func TestUserCookieMiddleware_InvalidCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotUserID string
	r := gin.New()
	r.Use(UserCookieMiddleware())
	r.GET("/test", func(c *gin.Context) {
		uid, _ := c.Get("userID")
		gotUserID = uid.(string)
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "user", Value: "invalid-value"})
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotUserID == "" {
		t.Error("userID должен быть установлен даже для невалидной cookie")
	}

	resp := rec.Result()
	defer resp.Body.Close()
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Error("новая cookie должна быть установлена при невалидной")
	}
}
