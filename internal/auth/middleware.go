package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey struct{}

// UserIDFromContext returns the authenticated user id set by the Middleware.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(contextKey{}).(int64)
	return id, ok
}

// Middleware rejects requests without a valid Bearer token and stores
// the user id in the request context for the handlers.
func (m *TokenManager) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		scheme, token, found := strings.Cut(header, " ")
		if !found || !strings.EqualFold(scheme, "Bearer") || strings.TrimSpace(token) == "" {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "missing or malformed Authorization header")
			return
		}

		userID, err := m.Parse(strings.TrimSpace(token))
		if err != nil {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}

		ctx := context.WithValue(r.Context(), contextKey{}, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
