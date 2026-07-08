package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ridex/backend/internal/auth"
	"ridex/backend/internal/middleware"
	"ridex/backend/internal/models"
	"ridex/backend/internal/store"
)

type AuthHandler struct {
	store      *store.Store
	jwtManager *auth.JWTManager
	mailer     auth.Mailer
	googleOAuth *auth.GoogleOAuth
	githubOAuth *auth.GitHubOAuth
	frontendURL string
}

func NewAuthHandler(store *store.Store, jwtManager *auth.JWTManager, mailer auth.Mailer, googleOAuth *auth.GoogleOAuth, githubOAuth *auth.GitHubOAuth, frontendURL string) *AuthHandler {
	return &AuthHandler{
		store: store,
		jwtManager: jwtManager,
		mailer: mailer,
		googleOAuth: googleOAuth,
		githubOAuth: githubOAuth,
		frontendURL: strings.TrimRight(frontendURL, "/"),
	}
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

type verifyEmailRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

type resetPasswordRequest struct {
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"newPassword"`
}

type updateProfileRequest struct {
	Name       string  `json:"name"`
	AvatarData *string `json:"avatarData"`
	Interests  []string `json:"interests"`
}

type authSuccessResponse struct {
	Token string          `json:"token,omitempty"`
	User  *publicUserView `json:"user,omitempty"`
}

type publicUserView struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	AvatarData     string   `json:"avatarData"`
	Interests      []string `json:"interests"`
	Rating         float64  `json:"rating"`
	RatingCount    int      `json:"ratingCount"`
	TripsCompleted int      `json:"tripsCompleted"`
	EmailVerified  bool     `json:"emailVerified"`
	AuthProvider   string   `json:"authProvider"`
	CreatedAt      string   `json:"createdAt"`
}

type publicProfileView struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	AvatarData     string   `json:"avatarData"`
	Interests      []string `json:"interests"`
	Rating         float64  `json:"rating"`
	RatingCount    int      `json:"ratingCount"`
	TripsCompleted int      `json:"tripsCompleted"`
	CreatedAt      string   `json:"createdAt"`
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
	if !auth.IsValidEmail(req.Email) {
		writeError(w, http.StatusBadRequest, "Enter a valid email address")
		return
	}
	if !auth.IsStrongPassword(req.Password) {
		writeError(w, http.StatusBadRequest, "Password must be at least 8 characters and include letters and numbers")
		return
	}
	if h.mailer == nil {
		writeError(w, http.StatusInternalServerError, "Email delivery is not configured on the server")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not secure password")
		return
	}

	verificationCode, err := generateVerificationCode()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not prepare email verification")
		return
	}

	_, err = h.store.CreateUser(r.Context(), req.Name, req.Email, hash, verificationCode)
	if err != nil {
		if errors.Is(err, store.ErrEmailExists) {
			writeError(w, http.StatusConflict, "Email already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not create user")
		return
	}

	if err := h.mailer.SendVerificationCode(req.Email, req.Name, verificationCode); err != nil {
		writeError(w, http.StatusInternalServerError, "Account created, but verification email could not be sent. Please request a new code.")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message":              "Account created. Check your email for the verification code.",
		"verificationRequired": true,
	})
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
	if !auth.IsValidEmail(req.Email) {
		writeError(w, http.StatusBadRequest, "Enter a valid email address")
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
	if !user.EmailVerified {
		writeError(w, http.StatusUnauthorized, "Email not verified. Verify your email before logging in.")
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

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req verifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Code = strings.TrimSpace(req.Code)

	if req.Email == "" || req.Code == "" {
		writeError(w, http.StatusBadRequest, "Email and verification code are required")
		return
	}
	if !auth.IsValidEmail(req.Email) {
		writeError(w, http.StatusBadRequest, "Enter a valid email address")
		return
	}

	if err := h.store.VerifyUserEmail(r.Context(), req.Email, req.Code); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "Invalid verification code")
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not verify email")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Email verified. Login now."})
}

func (h *AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req verifyEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "Email is required")
		return
	}
	if !auth.IsValidEmail(req.Email) {
		writeError(w, http.StatusBadRequest, "Enter a valid email address")
		return
	}

	user, err := h.store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "No account found for that email")
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not resend verification")
		return
	}
	if user.EmailVerified {
		writeJSON(w, http.StatusOK, map[string]any{
			"message": "Email is already verified.",
		})
		return
	}
	if h.mailer == nil {
		writeError(w, http.StatusInternalServerError, "Email delivery is not configured on the server")
		return
	}

	verificationCode, err := generateVerificationCode()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not prepare email verification")
		return
	}
	if err := h.store.UpdateVerificationCode(r.Context(), req.Email, verificationCode); err != nil {
		writeError(w, http.StatusInternalServerError, "Could not resend verification")
		return
	}
	if err := h.mailer.SendVerificationCode(req.Email, user.Name, verificationCode); err != nil {
		writeError(w, http.StatusInternalServerError, "Could not send verification email")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Verification code sent to your email.",
	})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req forgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "Email is required")
		return
	}
	if !auth.IsValidEmail(req.Email) {
		writeError(w, http.StatusBadRequest, "Enter a valid email address")
		return
	}
	if h.mailer == nil {
		writeError(w, http.StatusInternalServerError, "Email delivery is not configured on the server")
		return
	}

	user, err := h.store.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusOK, map[string]string{"message": "If that email exists, a password reset code has been sent."})
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not process forgot password request")
		return
	}

	resetCode, err := generateVerificationCode()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not generate reset code")
		return
	}
	if err := h.store.SetPasswordResetCode(r.Context(), req.Email, resetCode, timeNowUTC()); err != nil {
		writeError(w, http.StatusInternalServerError, "Could not process forgot password request")
		return
	}
	if err := h.mailer.SendPasswordResetCode(req.Email, user.Name, resetCode); err != nil {
		writeError(w, http.StatusInternalServerError, "Could not send password reset email")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "If that email exists, a password reset code has been sent."})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req resetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	req.Code = strings.TrimSpace(req.Code)
	if req.Email == "" || req.Code == "" || req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "Email, reset code, and new password are required")
		return
	}
	if !auth.IsValidEmail(req.Email) {
		writeError(w, http.StatusBadRequest, "Enter a valid email address")
		return
	}
	if !auth.IsStrongPassword(req.NewPassword) {
		writeError(w, http.StatusBadRequest, "Password must be at least 8 characters and include letters and numbers")
		return
	}

	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not secure password")
		return
	}
	if err := h.store.ResetPassword(r.Context(), req.Email, req.Code, hash, timeNowUTC(), 15*time.Minute); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "Invalid or expired reset code")
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not reset password")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Password reset successful. Login with your new password."})
}

func (h *AuthHandler) GoogleOAuthStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.googleOAuth == nil || !h.googleOAuth.Enabled() {
		writeError(w, http.StatusInternalServerError, "Google sign-in is not configured")
		return
	}

	state, err := auth.GenerateStateToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not start Google sign-in")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "ridex_google_oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	})

	http.Redirect(w, r, h.googleOAuth.AuthURL(state), http.StatusFound)
}

func (h *AuthHandler) GoogleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.googleOAuth == nil || !h.googleOAuth.Enabled() {
		writeError(w, http.StatusInternalServerError, "Google sign-in is not configured")
		return
	}

	stateCookie, err := r.Cookie("ridex_google_oauth_state")
	if err != nil || stateCookie.Value == "" || r.URL.Query().Get("state") != stateCookie.Value {
		h.redirectWithAuthError(w, r, "Google sign-in state validation failed")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "ridex_google_oauth_state",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		h.redirectWithAuthError(w, r, "Google did not return an authorization code")
		return
	}

	accessToken, err := h.googleOAuth.ExchangeCode(code)
	if err != nil {
		h.redirectWithAuthError(w, r, "Google token exchange failed")
		return
	}

	userInfo, err := h.googleOAuth.FetchUserInfo(accessToken)
	if err != nil {
		h.redirectWithAuthError(w, r, "Google user profile fetch failed")
		return
	}
	if !userInfo.VerifiedEmail {
		h.redirectWithAuthError(w, r, "Google account email is not verified")
		return
	}

	user, err := h.store.GetUserByEmail(r.Context(), userInfo.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			user, err = h.store.CreateOAuthUser(r.Context(), userInfo.Name, userInfo.Email, "google")
			if err != nil {
				h.redirectWithAuthError(w, r, "Could not create Google account")
				return
			}
		} else {
			h.redirectWithAuthError(w, r, "Could not complete Google sign-in")
			return
		}
	} else if !user.EmailVerified {
		_ = h.store.VerifyUserByEmail(r.Context(), user.Email)
		user.EmailVerified = true
	}

	token, err := h.jwtManager.Generate(user.ID, user.Email)
	if err != nil {
		h.redirectWithAuthError(w, r, "Could not create session token")
		return
	}

	http.Redirect(w, r, h.frontendURL+"/?token="+url.QueryEscape(token), http.StatusFound)
}

func (h *AuthHandler) GitHubOAuthStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.githubOAuth == nil || !h.githubOAuth.Enabled() {
		writeError(w, http.StatusInternalServerError, "GitHub sign-in is not configured")
		return
	}

	state, err := auth.GenerateStateToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not start GitHub sign-in")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "ridex_github_oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
	})

	http.Redirect(w, r, h.githubOAuth.AuthURL(state), http.StatusFound)
}

func (h *AuthHandler) GitHubOAuthCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	if h.githubOAuth == nil || !h.githubOAuth.Enabled() {
		writeError(w, http.StatusInternalServerError, "GitHub sign-in is not configured")
		return
	}

	stateCookie, err := r.Cookie("ridex_github_oauth_state")
	if err != nil || stateCookie.Value == "" || r.URL.Query().Get("state") != stateCookie.Value {
		h.redirectWithAuthError(w, r, "GitHub sign-in state validation failed")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "ridex_github_oauth_state",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		h.redirectWithAuthError(w, r, "GitHub did not return an authorization code")
		return
	}

	accessToken, err := h.githubOAuth.ExchangeCode(code)
	if err != nil {
		h.redirectWithAuthError(w, r, "GitHub token exchange failed")
		return
	}

	userInfo, err := h.githubOAuth.FetchUserInfo(accessToken)
	if err != nil {
		h.redirectWithAuthError(w, r, "GitHub user profile fetch failed")
		return
	}
	if !userInfo.VerifiedEmail {
		h.redirectWithAuthError(w, r, "GitHub account email is not verified")
		return
	}

	user, err := h.store.GetUserByEmail(r.Context(), userInfo.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			user, err = h.store.CreateOAuthUser(r.Context(), userInfo.Name, userInfo.Email, "github")
			if err != nil {
				h.redirectWithAuthError(w, r, "Could not create GitHub account")
				return
			}
		} else {
			h.redirectWithAuthError(w, r, "Could not complete GitHub sign-in")
			return
		}
	} else if !user.EmailVerified {
		_ = h.store.VerifyUserByEmail(r.Context(), user.Email)
		user.EmailVerified = true
	}

	token, err := h.jwtManager.Generate(user.ID, user.Email)
	if err != nil {
		h.redirectWithAuthError(w, r, "Could not create session token")
		return
	}

	http.Redirect(w, r, h.frontendURL+"/?token="+url.QueryEscape(token), http.StatusFound)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDContextKey).(string)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	if r.Method == http.MethodPut {
		h.updateMe(w, r, userID)
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

func (h *AuthHandler) PublicProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID := strings.TrimPrefix(r.URL.Path, "/api/users/")
	userID = strings.TrimSuffix(userID, "/profile")
	userID = strings.Trim(userID, "/")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "User ID is required")
		return
	}

	user, err := h.store.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "User not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not get public profile")
		return
	}

	writeJSON(w, http.StatusOK, map[string]*publicProfileView{"user": toPublicProfile(user)})
}

func (h *AuthHandler) updateMe(w http.ResponseWriter, r *http.Request, userID string) {
	var req updateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	currentUser, err := h.store.GetUserByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "Invalid token")
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not update profile")
		return
	}

	name := currentUser.Name
	if trimmed := strings.TrimSpace(req.Name); trimmed != "" {
		if len([]rune(trimmed)) > 80 {
			writeError(w, http.StatusBadRequest, "Name must be 80 characters or fewer")
			return
		}
		name = trimmed
	}

	avatarData := currentUser.AvatarData
	if req.AvatarData != nil {
		avatarData = strings.TrimSpace(*req.AvatarData)
		if avatarData != "" && !strings.HasPrefix(avatarData, "data:image/") {
			writeError(w, http.StatusBadRequest, "Avatar must be a valid image")
			return
		}
		if len(avatarData) > 2_000_000 {
			writeError(w, http.StatusBadRequest, "Avatar image is too large")
			return
		}
	}

	interests := currentUser.Interests
	if req.Interests != nil {
		interests = make([]string, 0, len(req.Interests))
		seen := map[string]struct{}{}
		for _, interest := range req.Interests {
			trimmed := strings.TrimSpace(interest)
			if trimmed == "" {
				continue
			}
			if len([]rune(trimmed)) > 40 {
				writeError(w, http.StatusBadRequest, "Each interest must be 40 characters or fewer")
				return
			}
			key := strings.ToLower(trimmed)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			interests = append(interests, trimmed)
		}
		if len(interests) > 10 {
			writeError(w, http.StatusBadRequest, "You can save up to 10 interests")
			return
		}
	}

	user, err := h.store.UpdateUserProfile(r.Context(), userID, name, avatarData, interests)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "Invalid token")
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not update profile")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Profile updated successfully.",
		"user":    toPublicUser(user),
	})
}

func toPublicUser(user *models.User) *publicUserView {
	return &publicUserView{
		ID:             user.ID,
		Name:           user.Name,
		Email:          user.Email,
		AvatarData:     user.AvatarData,
		Interests:      user.Interests,
		Rating:         user.Rating,
		RatingCount:    user.RatingCount,
		TripsCompleted: user.TripsCompleted,
		EmailVerified:  user.EmailVerified,
		AuthProvider:   user.AuthProvider,
		CreatedAt:      user.CreatedAt.UTC().Format(timeLayout),
	}
}

func toPublicProfile(user *models.User) *publicProfileView {
	return &publicProfileView{
		ID:             user.ID,
		Name:           user.Name,
		AvatarData:     user.AvatarData,
		Interests:      user.Interests,
		Rating:         user.Rating,
		RatingCount:    user.RatingCount,
		TripsCompleted: user.TripsCompleted,
		CreatedAt:      user.CreatedAt.UTC().Format(timeLayout),
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

func generateVerificationCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}

func timeNowUTC() time.Time {
	return time.Now().UTC()
}

func (h *AuthHandler) redirectWithAuthError(w http.ResponseWriter, r *http.Request, message string) {
	http.Redirect(w, r, h.frontendURL+"/?auth_error="+url.QueryEscape(message), http.StatusFound)
}
