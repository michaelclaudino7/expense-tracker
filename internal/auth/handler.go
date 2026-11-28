package auth

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/michaelclaudino7/expense-tracker/internal/user"
)

// dummyHash is compared against when the email does not exist, so a login
// for an unknown email takes about as long as one for a known email.
var dummyHash, _ = HashPassword("dummy-password-for-timing")

type Handler struct {
	users  *user.Repository
	tokens *TokenManager
}

func NewHandler(users *user.Repository, tokens *TokenManager) *Handler {
	return &Handler{users: users, tokens: tokens}
}

// RegisterRoutes attaches the public auth endpoints to the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/register", h.register)
	mux.HandleFunc("POST /auth/login", h.login)
}

type authResponse struct {
	Token string    `json:"token"`
	User  user.User `json:"user"`
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	c, ok := decodeCredentials(w, r)
	if !ok {
		return
	}
	if msg := c.Validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	hash, err := HashPassword(c.Password)
	if err != nil {
		log.Printf("failed to hash password: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	u, err := h.users.Create(r.Context(), c.Email, hash)
	if err != nil {
		if errors.Is(err, user.ErrEmailTaken) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		log.Printf("internal error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.respondWithToken(w, http.StatusCreated, u)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	c, ok := decodeCredentials(w, r)
	if !ok {
		return
	}
	c.Normalize()
	if c.Email == "" || c.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	const invalid = "invalid email or password"

	u, err := h.users.GetByEmail(r.Context(), c.Email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			CheckPassword(dummyHash, c.Password)
			writeError(w, http.StatusUnauthorized, invalid)
			return
		}
		log.Printf("internal error: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if !CheckPassword(u.PasswordHash, c.Password) {
		writeError(w, http.StatusUnauthorized, invalid)
		return
	}

	h.respondWithToken(w, http.StatusOK, u)
}

func (h *Handler) respondWithToken(w http.ResponseWriter, status int, u user.User) {
	token, err := h.tokens.Generate(u.ID)
	if err != nil {
		log.Printf("failed to generate token: %v", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, status, authResponse{Token: token, User: u})
}

func decodeCredentials(w http.ResponseWriter, r *http.Request) (Credentials, bool) {
	var c Credentials

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return Credentials{}, false
	}
	return c, true
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
