package report

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/michaelclaudino7/expense-tracker/internal/auth"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes attaches the report endpoints to the given mux.
// Every route is wrapped by the authentication middleware.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("GET /reports/monthly", authMW(http.HandlerFunc(h.monthly)))
}

func (h *Handler) monthly(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return
	}

	month := r.URL.Query().Get("month")
	if month == "" {
		writeError(w, http.StatusBadRequest, "month is required, in the format YYYY-MM")
		return
	}

	start, err := parseMonth(month)
	if err != nil {
		writeError(w, http.StatusBadRequest, "month must be a valid month in the format YYYY-MM")
		return
	}

	report, err := h.repo.Monthly(r.Context(), userID, start)
	if err != nil {
		log.Printf("internal error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
