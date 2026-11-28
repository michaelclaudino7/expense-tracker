package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddleware(t *testing.T) {
	m := NewTokenManager("test-secret", time.Hour)
	valid, _ := m.Generate(7)

	var gotID int64
	var gotOK bool
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID, gotOK = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	handler := m.Middleware(next)

	tests := []struct {
		name       string
		header     string
		wantStatus int
		wantCalled bool
	}{
		{"valid token", "Bearer " + valid, http.StatusOK, true},
		{"lowercase scheme", "bearer " + valid, http.StatusOK, true},
		{"no header", "", http.StatusUnauthorized, false},
		{"wrong scheme", "Basic " + valid, http.StatusUnauthorized, false},
		{"token missing", "Bearer ", http.StatusUnauthorized, false},
		{"invalid token", "Bearer not-a-token", http.StatusUnauthorized, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK = 0, false

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if gotOK != tt.wantCalled {
				t.Errorf("next called = %v, want %v", gotOK, tt.wantCalled)
			}
			if tt.wantCalled && gotID != 7 {
				t.Errorf("user id = %d, want 7", gotID)
			}
		})
	}
}
