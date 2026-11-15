package auth

import (
	"net/http"
	"strings"

	"github.com/JosePintos/tora/apps/api/internal/user/application/ports"
)

func AuthMiddleware(tokenProvider ports.TokenProvider, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")

		if header == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			next.ServeHTTP(w, r)
			return
		}

		userID, err := tokenProvider.ParseToken(parts[1])
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}

		ctx := WithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
