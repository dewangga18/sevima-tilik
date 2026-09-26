package handler

import (
	"context"
	"github.com/sevima/tilik-api/internal/domain"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sevima/tilik-api/internal/service"
)

func TestDemoLoginDisabledAndInvalidRole(t *testing.T) {
	for _, test := range []struct {
		name    string
		enabled bool
		role    string
		status  int
	}{
		{"disabled", false, "student", http.StatusNotFound},
		{"invalid role", true, "superuser", http.StatusBadRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			h := NewAuthHandler(service.NewAuthService(nil, test.enabled))
			req := httptest.NewRequest(http.MethodPost, "/api/auth/demo-login", strings.NewReader(`{"role":"`+test.role+`"}`))
			response := httptest.NewRecorder()
			h.DemoLogin(response, req)
			if response.Code != test.status {
				t.Fatalf("want %d, got %d", test.status, response.Code)
			}
		})
	}
}

func TestStudentRouteRoles(t *testing.T) {
	routes := []struct{ method, path string }{
		{"POST", "/api/diagnostic/start"}, {"POST", "/api/diagnostic/submit"}, {"POST", "/api/diagnostic/answer"},
		{"GET", "/api/diagnostic/latest"}, {"GET", "/api/diagnostic/history"}, {"GET", "/api/diagnostic/attempt-id"},
	}
	for _, route := range routes {
		for _, role := range []domain.Role{domain.RoleStudent, domain.RoleTeacher, domain.RoleAdmin, "unknown", ""} {
			t.Run(route.path+"/"+string(role), func(t *testing.T) {
				called := false
				endpoint := requireStudent(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true; w.WriteHeader(http.StatusNoContent) }))
				req := httptest.NewRequest(route.method, route.path, strings.NewReader("invalid body"))
				if role != "" {
					req = req.WithContext(context.WithValue(req.Context(), UserContextKey, &domain.User{ID: "test-user", Role: role}))
				}
				response := httptest.NewRecorder()
				endpoint.ServeHTTP(response, req)
				want := http.StatusForbidden
				if role == "" {
					want = http.StatusUnauthorized
				}
				if role == domain.RoleStudent {
					want = http.StatusNoContent
				}
				if response.Code != want || called != (role == domain.RoleStudent) {
					t.Fatalf("status %d, invoked %v", response.Code, called)
				}
				if role != "" && role != domain.RoleStudent && !strings.Contains(response.Body.String(), "Fitur ini hanya tersedia untuk siswa") {
					t.Fatal("missing safe permission error")
				}
			})
		}
	}
}
