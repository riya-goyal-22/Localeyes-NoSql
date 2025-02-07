package middlewares

import (
	"context"
	"encoding/json"
	"localeyes/utils"
	"net/http"
	"strings"
)

func AuthenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		excludedPaths := []string{"/login", "/signup", "/sns", "/otp", "/password/reset"}
		for _, path := range excludedPaths {
			if strings.Contains(r.URL.Path, path) {
				next.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			response := utils.NewUnauthorizedError("Missing authentication token")
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				utils.Logger.Error("ERROR: Error encoding response")
			}
			return
		}
		if !utils.ValidateTokenFunc(authHeader) {
			w.WriteHeader(http.StatusUnauthorized)
			response := utils.NewUnauthorizedError("Invalid token")
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				utils.Logger.Error("ERROR: Error encoding response")
			}
			return
		}
		claims, err := utils.ExtractClaimsFunc(authHeader)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			response := utils.NewUnauthorizedError("Invalid token")
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				utils.Logger.Error("ERROR: Error encoding response")
			}
			return
		}
		id := claims["id"]
		ctx := context.WithValue(r.Context(), "Id", id)
		ctx = context.WithValue(ctx, "Role", "user")
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func AdminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			response := utils.NewUnauthorizedError("Missing authentication token")
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				utils.Logger.Error("ERROR: Error encoding response")
			}
			return
		}
		if !utils.ValidateAdminTokenFunc(authHeader) {
			w.WriteHeader(http.StatusUnauthorized)
			response := utils.NewUnauthorizedError("Not an admin")
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				utils.Logger.Error("ERROR: Error encoding response")
			}
			return
		}
		claims, err := utils.ExtractClaimsFunc(authHeader)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			response := utils.NewUnauthorizedError("Invalid token")
			err := json.NewEncoder(w).Encode(response)
			if err != nil {
				utils.Logger.Error("ERROR: Error encoding response")
			}
			return
		}
		role := claims["sub"]
		id := claims["id"]
		ctx := context.WithValue(r.Context(), "Role", role)
		ctx = context.WithValue(ctx, "Id", id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
