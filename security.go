package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type PasswordHasher struct{}

func NewPasswordHasher() *PasswordHasher { return &PasswordHasher{} }

func (p *PasswordHasher) HashPassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil { return "", err }
	return string(h), nil
}

func (p *PasswordHasher) VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

type JWTManager struct {
	secret []byte
	exp    time.Duration
}

func NewJWTManager(secret string, exp time.Duration) *JWTManager {
	return &JWTManager{secret: []byte(secret), exp: exp}
}

func (j *JWTManager) GenerateToken(userID int64, isAdmin bool) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"adm": isAdmin,
		"exp": time.Now().Add(j.exp).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString(j.secret)
}

func (j *JWTManager) ParseToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// Context keys

type ctxKey string

const (
	ctxUserIDKey ctxKey = "user_id"
)

func JWTAuthMiddleware(j *JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "missing bearer token"})
				return
			}
			tok := strings.TrimPrefix(auth, "Bearer ")
			claims, err := j.ParseToken(tok)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid token"})
				return
			}
			idF, ok := claims["sub"].(float64)
			if !ok {
				writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid subject"})
				return
			}
			r = r.WithContext(context.WithValue(r.Context(), ctxUserIDKey, int64(idF)))
			next.ServeHTTP(w, r)
		})
	}
}

func AdminOnlyMiddleware(userService *UserService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			v := r.Context().Value(ctxUserIDKey)
			userID, ok := v.(int64)
			if !ok || userID == 0 {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "forbidden"})
				return
			}
			u, err := userService.GetByID(r.Context(), userID)
			if err != nil || !u.IsAdmin {
				writeJSON(w, http.StatusForbidden, map[string]string{"error": "admin only"})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}