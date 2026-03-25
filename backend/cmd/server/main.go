package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"ridex/backend/internal/auth"
	"ridex/backend/internal/config"
	"ridex/backend/internal/handlers"
	"ridex/backend/internal/middleware"
	"ridex/backend/internal/store"
)

func main() {
	if err := config.LoadEnvFile(".env"); err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}

	port := getEnv("PORT", "8080")
	jwtSecret := getEnv("JWT_SECRET", "ridex-dev-secret-change-me")
	dbPath := getEnv("DB_PATH", filepath.Join("data", "ridex.db"))
	frontendOrigin := getEnv("FRONTEND_ORIGIN", "http://localhost:5173")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := getEnv("SMTP_PORT", "587")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := os.Getenv("SMTP_FROM")
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	googleRedirectURI := getEnv("GOOGLE_REDIRECT_URI", "http://localhost:8080/api/auth/oauth/google/callback")
	githubClientID := os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	githubRedirectURI := getEnv("GITHUB_REDIRECT_URI", "http://localhost:8080/api/auth/oauth/github/callback")

	st, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}
	defer st.DB.Close()

	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
	mailer := auth.NewSMTPMailer(smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom)
	googleOAuth := &auth.GoogleOAuth{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURI:  googleRedirectURI,
	}
	githubOAuth := &auth.GitHubOAuth{
		ClientID:     githubClientID,
		ClientSecret: githubClientSecret,
		RedirectURI:  githubRedirectURI,
	}
	if mailer == nil {
		log.Printf("warning: SMTP email delivery is not configured; verification and password reset emails will fail until SMTP env vars are set")
	}
	if !googleOAuth.Enabled() {
		log.Printf("warning: Google OAuth is not configured; Google sign-in will be unavailable until GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET are set")
	}
	if !githubOAuth.Enabled() {
		log.Printf("warning: GitHub OAuth is not configured; GitHub sign-in will be unavailable until GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET are set")
	}
	authHandler := handlers.NewAuthHandler(st, jwtManager, mailer, googleOAuth, githubOAuth, frontendOrigin)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/signup", authHandler.Signup)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.HandleFunc("/api/auth/verify-email", authHandler.VerifyEmail)
	mux.HandleFunc("/api/auth/resend-verification", authHandler.ResendVerification)
	mux.HandleFunc("/api/auth/forgot-password", authHandler.ForgotPassword)
	mux.HandleFunc("/api/auth/reset-password", authHandler.ResetPassword)
	mux.HandleFunc("/api/auth/oauth/google/start", authHandler.GoogleOAuthStart)
	mux.HandleFunc("/api/auth/oauth/google/callback", authHandler.GoogleOAuthCallback)
	mux.HandleFunc("/api/auth/oauth/github/start", authHandler.GitHubOAuthStart)
	mux.HandleFunc("/api/auth/oauth/github/callback", authHandler.GitHubOAuthCallback)
	mux.Handle("/api/auth/me", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(authHandler.Me)))
	mux.Handle("/api/users/", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(authHandler.PublicProfile)))
	mux.Handle("/api/ride-requests", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(authHandler.RideRequests)))
	mux.Handle("/api/ride-requests/feed", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(authHandler.RideRequestFeed)))
	mux.Handle("/api/published-rides", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(authHandler.PublishedRides)))
	mux.Handle("/api/ride-requests/", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(authHandler.RideRequestByID)))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      withCORS(frontendOrigin, mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("RideX auth server running on http://localhost:%s", port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func withCORS(origin string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
