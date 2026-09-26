package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
	"github.com/sevima/tilik-api/internal/service"
)

func TestDiagnosticDatabaseErrorsAreSafe(t *testing.T) {
	sqlDB, err := sql.Open("pgx", "postgres://unused")
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.Close()
	db := &repository.DB{DB: sqlDB}
	h := NewDiagnosticHandler(service.NewDiagnosticService(repository.NewCurriculumRepository(db), repository.NewAssessmentRepository(db)))
	for _, test := range []struct {
		name, body, message string
		handle              http.HandlerFunc
	}{
		{"start", `{}`, "Gagal memulai tes diagnostik. Silakan coba lagi.", h.Start},
		{"submit", `{"assessment_id":"test"}`, "Gagal mengevaluasi asesmen. Silakan coba lagi.", h.Submit},
		{"latest", "", "Gagal memuat asesmen", h.GetLatest},
		{"skills", "", "Gagal memuat daftar skill", h.GetSkills},
	} {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			req = req.WithContext(context.WithValue(req.Context(), UserContextKey, &domain.User{ID: "student", GradeLevel: 4}))
			response := httptest.NewRecorder()
			test.handle(response, req)
			var body APIResponse
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if response.Code != http.StatusInternalServerError || body.Error != test.message {
				t.Fatalf("unsafe/unexpected response: %d %s", response.Code, response.Body.String())
			}
		})
	}
}
