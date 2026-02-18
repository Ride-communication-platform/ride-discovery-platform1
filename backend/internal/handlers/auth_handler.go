package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"ridex/backend/internal/auth"
	"ridex/backend/internal/middleware"
	"ridex/backend/internal/models"
	"ridex/backend/internal/store"
)

type AuthHandler struct {
	store      *store.Store
	jwtManager *auth.JWTManager
}

func NewAuthHandler(store *store.Store, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{store: store, jwtManager: jwtManager}
}

type signupRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authSuccessResponse struct {
	Token string          `json:"token,omitempty"`
	User  *publicUserView `json:"user,omitempty"`
}

type publicUserView struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	Rating    float64 `json:"rating"`
	CreatedAt string  `json:"createdAt"`
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req signupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	if req.Name == "" || req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Name, email, and password are required")
		return
	}
	if len(req.Password) < 6 {
		writeError(w, http.StatusBadRequest, "Password must be at least 6 characters")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not secure password")
		return
	}

	_, err = h.store.CreateUser(r.Context(), req.Name, req.Email, hash)
	if err != nil {
		if errors.Is(err, store.ErrEmailExists) {
			writeError(w, http.StatusConflict, "Email already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not create user")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"message": "Account created"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, err := h.store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not login")
		return
	}

	if err := auth.ComparePassword(user.PasswordHash, req.Password); err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := h.jwtManager.Generate(user.ID, user.Email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not create token")
		return
	}

	writeJSON(w, http.StatusOK, authSuccessResponse{
		Token: token,
		User:  toPublicUser(user),
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDContextKey).(string)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	user, err := h.store.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "Invalid token")
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not get profile")
		return
	}

	writeJSON(w, http.StatusOK, map[string]*publicUserView{"user": toPublicUser(user)})
}

func toPublicUser(user *models.User) *publicUserView {
	return &publicUserView{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Rating:    user.Rating,
		CreatedAt: user.CreatedAt.UTC().Format(timeLayout),
	}
}

const timeLayout = "2006-01-02T15:04:05Z07:00"

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
