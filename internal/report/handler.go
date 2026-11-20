package report

import (
    "encoding/json"
    "log"
    "net/http"
)

type Handler struct {
    repo *Repository
}

func NewHandler(repo *Repository) *Handler {
    return &Handler{repo: repo}
}

// RegisterRoutes attaches the report endpoints to the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
    mux.HandleFunc("GET /reports/monthly", h.monthly)
}

func (h *Handler) monthly(w http.ResponseWriter, r *http.Request) {
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

    report, err := h.repo.Monthly(r.Context(), start)
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