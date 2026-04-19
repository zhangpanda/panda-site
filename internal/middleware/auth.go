package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type ctxKey int
const adminKey ctxKey = 0

var JWTSecret = func() []byte {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return []byte(s)
	}
	log.Println("⚠️  未设置 JWT_SECRET，使用随机密钥（重启后 token 失效）")
	b := make([]byte, 32)
	rand.Read(b)
	return []byte(hex.EncodeToString(b))
}()

func HashPassword(pw string) string {
	h, _ := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(h)
}

func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

func GenerateToken(username string) (string, error) {
	return jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}).SignedString(JWTSecret)
}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"unauthorized"}`, 401)
			return
		}
		parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) { return JWTSecret, nil })
		if err != nil || !parsed.Valid {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"invalid token"}`, 401)
			return
		}
		claims := parsed.Claims.(jwt.MapClaims)
		sub, _ := claims["sub"].(string)
		ctx := context.WithValue(r.Context(), adminKey, sub)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetAdmin(r *http.Request) string {
	if v, ok := r.Context().Value(adminKey).(string); ok {
		return v
	}
	return ""
}
