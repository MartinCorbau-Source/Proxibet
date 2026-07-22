package auth

import (
	"net/http"
	"strings"
)

func NewAuthenticationMiddleware(tokenGenerator *TokenGenerator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearerToken, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok {
				writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Le token est invalide.")
				return
			}

			userID, err := tokenGenerator.Validate(bearerToken)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Le token est invalide.")
				return
			}

			next.ServeHTTP(w, r.WithContext(ContextWithUserID(r.Context(), userID)))
		})
	}
}
