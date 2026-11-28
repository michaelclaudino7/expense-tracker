package transaction

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/michaelclaudino7/expense-tracker/internal/auth"
)

type Handler struct {
	repo *Repository
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{repo: repo}
}

// RegisterRoutes attaches all transaction endpoints to the given mux.
// Every route is wrapped by the authentication middleware.
func (h *Handler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	protect := func(f http.HandlerFunc) http.Handler { return authMW(f) }

	mux.Handle("POST /transactions", protect(h.create))
	mux.Handle("GET /transactions", protect(h.list))
	mux.Handle("GET /transactions/{id}", protect(h.get))
	mux.Handle("PUT /transactions/{id}", protect(h.update))
	mux.Handle("DELETE /transactions/{id}", protect(h.delete))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}
	in, ok := decodeInput(w, r)
	if !ok {
		return
	}

	tx, err := h.repo.Create(r.Context(), userID, in)
	if err != nil {
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, tx)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}
	filter, ok := parseFilter(w, r)
	if !ok {
		return
	}
	pagination, ok := parsePagination(w, r)
	if !ok {
		return
	}

	transactions, total, err := h.repo.List(r.Context(), userID, filter, pagination)
	if err != nil {
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, newPagedResult(transactions, total, pagination))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	tx, err := h.repo.Get(r.Context(), userID, id)
	if err != nil {
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tx)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	in, ok := decodeInput(w, r)
	if !ok {
		return
	}

	tx, err := h.repo.Update(r.Context(), userID, id, in)
	if err != nil {
		writeRepoError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tx)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUser(w, r)
	if !ok {
		return
	}
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	if err := h.repo.Delete(r.Context(), userID, id); err != nil {
		writeRepoError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// currentUser reads the user id set by the auth middleware.
// It should never fail on protected routes, but we fail closed if it does.
func currentUser(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated")
		return 0, false
	}
	return id, true
}

// parseFilter reads the optional query params: from, to (YYYY-MM-DD) and category_id.
func parseFilter(w http.ResponseWriter, r *http.Request) (Filter, bool) {
	var f Filter
	q := r.URL.Query()

	if v := q.Get("from"); v != "" {
		t, err := time.Parse(dateLayout, v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "from must be a valid date in the format YYYY-MM-DD")
			return Filter{}, false
		}
		f.From = &t
	}
	if v := q.Get("to"); v != "" {
		t, err := time.Parse(dateLayout, v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "to must be a valid date in the format YYYY-MM-DD")
			return Filter{}, false
		}
		f.To = &t
	}
	if v := q.Get("category_id"); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil || id <= 0 {
			writeError(w, http.StatusBadRequest, "category_id must be a positive integer")
			return Filter{}, false
		}
		f.CategoryID = &id
	}
	return f, true
}

// parseID reads the {id} path value. On failure it writes a 400 and returns false.
func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

// decodeInput reads and validates the JSON body. On failure it writes a 400 and returns false.
func decodeInput(w http.ResponseWriter, r *http.Request) (Input, bool) {
	var in Input

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&in); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return Input{}, false
	}
	if msg := in.Validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return Input{}, false
	}
	return in, true
}

// writeRepoError maps our named errors to HTTP status codes.
func writeRepoError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrCategoryNotFound):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		log.Printf("internal error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
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
