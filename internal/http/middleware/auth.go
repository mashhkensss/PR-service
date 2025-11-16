package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/mashhkensss/PR-service/internal/http/httperror"
	"github.com/mashhkensss/PR-service/internal/http/response"
)

type Claims struct {
	Subject   string `json:"sub"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"exp"`
}

type Authorization struct {
	adminSecret []byte
	userSecret  []byte
}

func NewAuthorization(adminSecret, userSecret []byte) *Authorization {
	return &Authorization{adminSecret: adminSecret, userSecret: userSecret}
}

func (a *Authorization) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := a.verifyToken(r.Header.Get("Authorization"))
		if err != nil || claims.Role != "admin" {
			status, resp := httperror.Unauthorized()
			response.ErrorResponse(w, status, resp)
			return
		}
		ctx := contextWithClaims(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Authorization) RequireUserOrAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := a.verifyToken(r.Header.Get("Authorization"))
		if err != nil {
			status, resp := httperror.Unauthorized()
			response.ErrorResponse(w, status, resp)
			return
		}
		ctx := contextWithClaims(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (a *Authorization) verifyToken(header string) (Claims, error) {
	var empty Claims
	const prefix = "Bearer "

	if header == "" || len(header) <= len(prefix) {
		return empty, errors.New("missing token")
	}

	if header[:len(prefix)] != prefix {
		return empty, errors.New("invalid header")
	}

	token := header[len(prefix):]
	parts := splitToken(token)
	if len(parts) != 3 {
		return empty, errors.New("invalid token")
	}

	if err := verifyHS256(parts[0], parts[1], parts[2], a.adminSecret, a.userSecret); err != nil {
		return empty, err
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return empty, err
	}

	var claims Claims

	if err := json.Unmarshal(payload, &claims); err != nil {
		return empty, err
	}

	if claims.ExpiresAt != 0 {
		exp := time.Unix(claims.ExpiresAt, 0)
		if time.Now().After(exp) {
			return empty, errors.New("token expired")
		}
	}

	return claims, nil
}

func splitToken(token string) []string {
	parts := make([]string, 0, 3)
	start := 0

	for i := 0; i < len(token); i++ {
		if token[i] == '.' {
			parts = append(parts, token[start:i])
			start = i + 1
		}
	}
	parts = append(parts, token[start:])
	return parts
}

func verifyHS256(header, payload, signature string, adminSecret, userSecret []byte) error {
	data := header + "." + payload
	sig, err := base64.RawURLEncoding.DecodeString(signature)

	if err != nil {
		return err
	}

	if len(adminSecret) > 0 && validHMAC(data, sig, adminSecret) {
		return nil
	}
	if len(userSecret) > 0 && validHMAC(data, sig, userSecret) {
		return nil
	}

	return errors.New("invalid signature")
}

func validHMAC(data string, signature []byte, secret []byte) bool {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(data))
	expected := mac.Sum(nil)

	return hmac.Equal(expected, signature)
}
