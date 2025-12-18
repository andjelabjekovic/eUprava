package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

type SignedDetails struct {
	Email      string `json:"Email"`
	First_name string `json:"First_name"`
	Last_name  string `json:"Last_name"`
	Uid        string `json:"Uid"`
	User_type  string `json:"User_type"`
	jwt.StandardClaims
}

type ctxKey string

const (
	CtxUserID    ctxKey = "userId"    // token Uid (hex string)
	CtxUserType  ctxKey = "userType"  // student/cook/...
	CtxFirstName ctxKey = "firstName" // from token
	CtxLastName  ctxKey = "lastName"
)

func AuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		if !strings.HasPrefix(strings.ToLower(auth), "bearer ") {
			http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimSpace(auth[len("Bearer "):])
		if tokenString == "" {
			http.Error(w, "Empty token", http.StatusUnauthorized)
			return
		}

		secret := os.Getenv("SECRET_KEY")
		if secret == "" {
			http.Error(w, "SECRET_KEY not configured", http.StatusInternalServerError)
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &SignedDetails{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || token == nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*SignedDetails)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}
		if claims.ExpiresAt > 0 && claims.ExpiresAt < time.Now().Unix() {
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), CtxUserID, claims.Uid)
		ctx = context.WithValue(ctx, CtxUserType, claims.User_type)
		ctx = context.WithValue(ctx, CtxFirstName, claims.First_name)
		ctx = context.WithValue(ctx, CtxLastName, claims.Last_name)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(r *http.Request) (string, bool) {
	v := r.Context().Value(CtxUserID)
	s, ok := v.(string)
	return s, ok && s != ""
}

func GetUserType(r *http.Request) (string, bool) {
	v := r.Context().Value(CtxUserType)
	s, ok := v.(string)
	return s, ok && s != ""
}

func GetFullName(r *http.Request) string {
	fn, _ := r.Context().Value(CtxFirstName).(string)
	ln, _ := r.Context().Value(CtxLastName).(string)
	full := strings.TrimSpace(fn + " " + ln)
	if full == "" {
		return "Unknown"
	}
	return full
}
