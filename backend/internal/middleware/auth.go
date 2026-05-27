package middleware

import (
	"context"
	"net/http"
	"strings"

	"tattoo-consultation/internal/auth"
)

type contextKey string

const ClaimsKey contextKey = "claims"

func AuthRequired(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, `{"error":"missing token"}`, 401)
				return
			}
			claims, err := auth.ValidateToken(header[7:], jwtSecret)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, 401)
				return
			}
			ctx := context.WithValue(r.Context(), ClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(ClaimsKey).(*auth.Claims)
		if !ok || claims == nil || claims.Role != "admin" {
			http.Error(w, `{"error":"admin required"}`, 403)
			return
		}
		next.ServeHTTP(w, r)
	})
}
