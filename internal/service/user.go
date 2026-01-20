// Package service содержит вспомогательные сервисы для работы с пользователями,
// в частности работу с пользовательскими cookie для идентификации.
package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

var secretKey = []byte("super-secret-key")

const cookieName = "user_id"

// GenerateUserCookie создаёт новый HTTP cookie для пользователя.
//
// Cookie содержит уникальный идентификатор пользователя и HMAC-подпись для проверки целостности.
// Возвращается cookie с флагами HttpOnly и Path="/".
func GenerateUserCookie() *http.Cookie {
	userID := uuid.New().String()
	sig := sign(userID)
	value := userID + "|" + sig

	return &http.Cookie{
		Name:     cookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	}
}

// ValidateUserCookie проверяет cookie пользователя из HTTP-запроса.
//
// Возвращает userID, если cookie валидна. В случае ошибки возвращает пустую строку и ошибку.
// Возможные ошибки:
// - cookie отсутствует
// - неверный формат cookie
// - неверная подпись (подделка cookie)
func ValidateUserCookie(r *http.Request) (string, error) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return "", err
	}

	parts := strings.SplitN(c.Value, "|", 2)
	if len(parts) != 2 {
		return "", errors.New("invalid cookie format")
	}

	userID, sig := parts[0], parts[1]
	if !verify(userID, sig) {
		return "", errors.New("invalid cookie signature")
	}
	return userID, nil
}

// sign создаёт HMAC-SHA256 подпись для переданной строки.
func sign(data string) string {
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// verify проверяет соответствие подписи данных HMAC-подписи.
func verify(data, sig string) bool {
	expected := sign(data)
	return hmac.Equal([]byte(sig), []byte(expected))
}
