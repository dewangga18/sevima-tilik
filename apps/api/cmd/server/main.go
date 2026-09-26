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
	"github.com/sevima/tilik-api/internal/middleware"
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

	// Periodic expired-session cleanup; failures only logged, never fatal.
	go func() {
		if _, err := userRepo.CleanupExpiredSessions(context.Background()); err != nil {
			log.Printf("Session cleanup error: %v", err)
		}
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if _, err := userRepo.CleanupExpiredSessions(context.Background()); err != nil {
				log.Printf("Session cleanup error: %v", err)
			}
		}
	}()

	// Services
	authService := service.NewAuthService(userRepo, demoEnabled)
	diagService := service.NewDiagnosticService(curriculumRepo, assessmentRepo)

	// Handlers
	isProduction := cfg.AppEnv == "production"
	authHandler := handler.NewAuthHandler(authService, isProduction)
	diagHandler := handler.NewDiagnosticHandler(diagService)
	learningService := service.NewLearningService(repository.NewLearningRepository(db), curriculumRepo, diagService)
	learningHandler := handler.NewLearningHandler(learningService)
	teacherService := service.NewTeacherService(db)
	teacherHandler := handler.NewTeacherHandler(teacherService)

	mux := http.NewServeMux()

	// Rate limiters
	// Auth endpoints: 10 requests per minute per IP
	authLimiter := middleware.NewRateLimiter(10, 1*time.Minute)
	// General API: 100 requests per minute per IP
	apiLimiter := middleware.NewRateLimiter(100, 1*time.Minute)

	// Public healthcheck (no rate limit)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status": "ok",
			"env":    cfg.AppEnv,
		})
	})

	// Auth routes (strict rate limiting)
	mux.Handle("POST /api/auth/login", authLimiter.Middleware(http.HandlerFunc(authHandler.Login)))
	if demoEnabled {
		mux.Handle("POST /api/auth/demo-login", authLimiter.Middleware(http.HandlerFunc(authHandler.DemoLogin)))
	}
	mux.Handle("POST /api/auth/logout", authLimiter.Middleware(http.HandlerFunc(authHandler.Logout)))
	mux.Handle("GET /api/auth/me", authHandler.AuthMiddleware(http.HandlerFunc(authHandler.Me)))

	// Curriculum routes (general rate limit)
	mux.Handle("GET /api/skills", apiLimiter.Middleware(http.HandlerFunc(diagHandler.GetSkills)))

	// Diagnostic routes (protected + rate limited)
	mux.Handle("POST /api/diagnostic/answer", apiLimiter.Middleware(authHandler.StudentMiddleware(http.HandlerFunc(diagHandler.Answer))))
	mux.Handle("POST /api/diagnostic/start", apiLimiter.Middleware(authHandler.StudentMiddleware(http.HandlerFunc(diagHandler.Start))))
	mux.Handle("POST /api/diagnostic/submit", apiLimiter.Middleware(authHandler.StudentMiddleware(http.HandlerFunc(diagHandler.Submit))))
	mux.Handle("GET /api/diagnostic/latest", apiLimiter.Middleware(authHandler.StudentMiddleware(http.HandlerFunc(diagHandler.GetLatest))))
	mux.Handle("GET /api/diagnostic/history", apiLimiter.Middleware(authHandler.StudentMiddleware(http.HandlerFunc(diagHandler.GetHistory))))
	mux.Handle("GET /api/diagnostic/{id}", apiLimiter.Middleware(authHandler.StudentMiddleware(http.HandlerFunc(diagHandler.GetByID))))

	mux.Handle("GET /api/learning/progress", apiLimiter.Middleware(authHandler.StudentMiddleware(http.HandlerFunc(learningHandler.Progress))))
	mux.Handle("GET /api/learning/lessons/{skill_id}", apiLimiter.Middleware(authHandler.StudentMiddleware(http.HandlerFunc(learningHandler.Lesson))))
	mux.Handle("GET /api/learning/sessions/{id}", apiLimiter.Middleware(authHandler.StudentMiddleware(http.HandlerFunc(learningHandler.Session))))
	mux.Handle("POST /api/learning/start", apiLimiter.Middleware(authHandler.StudentMiddleware(http.HandlerFunc(learningHandler.Start))))
	mux.Handle("POST /api/learning/lesson-complete", apiLimiter.Middleware(authHandler.StudentMiddleware(http.HandlerFunc(learningHandler.CompleteLesson))))
	mux.Handle("POST /api/learning/answer", apiLimiter.Middleware(authHandler.StudentMiddleware(http.HandlerFunc(learningHandler.Answer))))

	// Teacher data routes (teacher role + per-class assignment enforced in service)
	mux.Handle("GET /api/teacher/classes", apiLimiter.Middleware(authHandler.TeacherMiddleware(http.HandlerFunc(teacherHandler.GetClasses))))
	mux.Handle("GET /api/teacher/classes/{class_id}/students", apiLimiter.Middleware(authHandler.TeacherMiddleware(http.HandlerFunc(teacherHandler.GetClassStudents))))
	mux.Handle("GET /api/teacher/students/{student_id}/insight", apiLimiter.Middleware(authHandler.TeacherMiddleware(http.HandlerFunc(teacherHandler.GetStudentInsight))))

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
