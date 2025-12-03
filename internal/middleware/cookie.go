package middleware

import (
	"net/http"

	"github.com/Guldana11/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

func UserCookieMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := service.ValidateUserCookie(c.Request)
		if err != nil || userID == "" {
			cookie := service.GenerateUserCookie()
			http.SetCookie(c.Writer, cookie)
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}
