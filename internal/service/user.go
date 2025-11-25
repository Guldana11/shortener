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

func sign(data string) string {
	h := hmac.New(sha256.New, secretKey)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func verify(data, sig string) bool {
	expected := sign(data)
	return hmac.Equal([]byte(sig), []byte(expected))
}
