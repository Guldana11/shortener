// Package middleware содержит промежуточное ПО для HTTP-сервера на Gin,
// включая логирование, работу с gzip и управление пользовательскими cookie.
package middleware

import (
	"net/http"

	"github.com/Guldana11/shortener/internal/service"
	"github.com/gin-gonic/gin"
)

// UserCookieMiddleware возвращает middleware для Gin, который обеспечивает идентификацию пользователя через cookie.
//
// Особенности работы:
// 1. Проверяет наличие валидного cookie пользователя с помощью service.ValidateUserCookie.
// 2. Если cookie отсутствует или некорректен, создаёт новый cookie через service.GenerateUserCookie и устанавливает его.
// 3. Записывает userID в контекст Gin под ключом "userID" для последующего использования в хендлерах.
// 4. Передаёт управление следующему middleware или обработчику через c.Next().
func UserCookieMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := service.ValidateUserCookie(c.Request)
		if err != nil || userID == "" {
			cookie := service.GenerateUserCookie()
			http.SetCookie(c.Writer, cookie)
			userID = cookie.Value[:36]
		}

		c.Set("userID", userID)

		c.Next()
	}
}
