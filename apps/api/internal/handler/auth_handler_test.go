package handler

import (
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
