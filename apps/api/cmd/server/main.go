package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sevima/tilik-api/internal/config"
	"github.com/sevima/tilik-api/internal/handler"
	"github.com/sevima/tilik-api/internal/repository"
	"github.com/sevima/tilik-api/internal/service"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Connecting to PostgreSQL database...")
	demoEnabled := cfg.AppEnv == "development"
	db, err := repository.Connect(ctx, cfg.DatabaseURL, demoEnabled)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()
	log.Println("Database connection, migrations, and seeds completed successfully.")

	// Repositories
	userRepo := repository.NewUserRepository(db)
	curriculumRepo := repository.NewCurriculumRepository(db)
	assessmentRepo := repository.NewAssessmentRepository(db)

	// Services
	authService := service.NewAuthService(userRepo, demoEnabled)
	diagService := service.NewDiagnosticService(curriculumRepo, assessmentRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	diagHandler := handler.NewDiagnosticHandler(diagService)

	mux := http.NewServeMux()

	// Public healthcheck
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"env":    cfg.AppEnv,
		})
	})

	// Auth routes
	mux.HandleFunc("POST /api/auth/login", authHandler.Login)
	if demoEnabled {
		mux.HandleFunc("POST /api/auth/demo-login", authHandler.DemoLogin)
	}
	mux.HandleFunc("POST /api/auth/logout", authHandler.Logout)
	mux.Handle("GET /api/auth/me", authHandler.AuthMiddleware(http.HandlerFunc(authHandler.Me)))

	// Curriculum routes
	mux.HandleFunc("GET /api/skills", diagHandler.GetSkills)

	// Diagnostic routes (protected)
	mux.Handle("POST /api/diagnostic/start", authHandler.AuthMiddleware(http.HandlerFunc(diagHandler.Start)))
	mux.Handle("POST /api/diagnostic/submit", authHandler.AuthMiddleware(http.HandlerFunc(diagHandler.Submit)))
	mux.Handle("GET /api/diagnostic/latest", authHandler.AuthMiddleware(http.HandlerFunc(diagHandler.GetLatest)))
	mux.Handle("GET /api/diagnostic/history", authHandler.AuthMiddleware(http.HandlerFunc(diagHandler.GetHistory)))
	mux.Handle("GET /api/diagnostic/{id}", authHandler.AuthMiddleware(http.HandlerFunc(diagHandler.GetByID)))

	// Middleware chain: Logging -> CORS
	wrappedMux := loggingMiddleware(corsMiddleware(cfg.AllowedOrigin)(mux))

	addr := "0.0.0.0:" + cfg.Port
	server := &http.Server{
		Addr:         addr,
		Handler:      wrappedMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Tilik API server listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down API server gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	log.Println("API server stopped.")
}

func corsMiddleware(origin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
