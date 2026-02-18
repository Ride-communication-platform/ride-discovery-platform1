package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"ridex/backend/internal/auth"
	"ridex/backend/internal/handlers"
	"ridex/backend/internal/middleware"
	"ridex/backend/internal/store"
)

func main() {
	port := getEnv("PORT", "8080")
	jwtSecret := getEnv("JWT_SECRET", "ridex-dev-secret-change-me")
	dbPath := getEnv("DB_PATH", filepath.Join("data", "ridex.db"))
	frontendOrigin := getEnv("FRONTEND_ORIGIN", "http://localhost:5173")

	st, err := store.New(dbPath)
	if err != nil {
		log.Fatalf("failed to initialize store: %v", err)
	}
	defer st.DB.Close()

	jwtManager := auth.NewJWTManager(jwtSecret, 24*time.Hour)
	authHandler := handlers.NewAuthHandler(st, jwtManager)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/signup", authHandler.Signup)
	mux.HandleFunc("/api/auth/login", authHandler.Login)
	mux.Handle("/api/auth/me", middleware.AuthMiddleware(jwtManager)(http.HandlerFunc(authHandler.Me)))

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
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
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
