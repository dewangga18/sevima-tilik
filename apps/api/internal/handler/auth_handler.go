package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/service"
)

type contextKey string

const UserContextKey contextKey = "user"

type AuthHandler struct {
	authService    *service.AuthService
	isProduction   bool
}

func NewAuthHandler(authService *service.AuthService, isProduction bool) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		isProduction: isProduction,
	}
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User  *domain.User `json:"user"`
	Token string       `json:"token"`
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" || req.Password == "" {
		WriteError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	user, session, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			WriteError(w, http.StatusUnauthorized, "Email atau kata sandi salah")
			return
		}
		WriteError(w, http.StatusInternalServerError, "Gagal melakukan login")
		return
	}

	h.writeLoginResponse(w, user, session)
}

func (h *AuthHandler) DemoLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Role domain.Role `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	user, session, err := h.authService.DemoLogin(r.Context(), req.Role)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrDemoDisabled):
			WriteError(w, http.StatusNotFound, "Not found")
		case errors.Is(err, service.ErrInvalidDemoRole):
			WriteError(w, http.StatusBadRequest, "Role demo tidak valid")
		case errors.Is(err, service.ErrUnauthorized):
			WriteError(w, http.StatusUnauthorized, "Akun demo tidak tersedia")
		default:
			log.Printf("demo login: %v", err)
			WriteError(w, http.StatusInternalServerError, "Gagal masuk ke akun demo. Silakan coba lagi.")
		}
		return
	}
	h.writeLoginResponse(w, user, session)
}

func (h *AuthHandler) writeLoginResponse(w http.ResponseWriter, user *domain.User, session *domain.Session) {
	http.SetCookie(w, &http.Cookie{
		Name:     "tilik_session",
		Value:    session.Token,
		Path:     "/",
		Expires:  session.ExpiresAt,
		HttpOnly: true,
		Secure:   h.isProduction,
		SameSite: http.SameSiteLaxMode,
	})

	WriteJSON(w, http.StatusOK, LoginResponse{
		User:  user,
		Token: session.Token,
	})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token != "" {
		_ = h.authService.Logout(r.Context(), token)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "tilik_session",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.isProduction,
	})

	WriteJSON(w, http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserContextKey).(*domain.User)
	if !ok || user == nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	WriteJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			WriteError(w, http.StatusUnauthorized, "Sesi tidak ditemukan atau telah kedaluwarsa")
			return
		}

		user, err := h.authService.ValidateSession(r.Context(), token)
		if err != nil || user == nil {
			WriteError(w, http.StatusUnauthorized, "Sesi tidak valid")
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// StudentMiddleware authenticates first, then restricts diagnostic actions by role.
func (h *AuthHandler) StudentMiddleware(next http.Handler) http.Handler {
	return h.AuthMiddleware(requireStudent(next))
}

func requireStudent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(UserContextKey).(*domain.User)
		if !ok || user == nil {
			WriteError(w, http.StatusUnauthorized, "Sesi tidak ditemukan atau telah kedaluwarsa")
			return
		}
		if user.Role != domain.RoleStudent {
			WriteError(w, http.StatusForbidden, "Fitur ini hanya tersedia untuk siswa")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func extractToken(r *http.Request) string {
	// 1. Check Authorization header
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	// 2. Check Cookie
	if cookie, err := r.Cookie("tilik_session"); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
}
