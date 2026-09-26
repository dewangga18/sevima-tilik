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
		{"history", "", "Gagal memuat riwayat belajar. Silakan coba lagi.", h.GetHistory},
		{"answer", `{"assessment_id":"test","question_id":"q","student_answer":"option"}`, "Jawaban belum bisa disimpan. Silakan coba lagi.", h.Answer},
		{"detail", "", "Gagal memuat asesmen. Silakan coba lagi.", h.GetByID},
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

func TestProgressiveRequestValidation(t *testing.T) {
	h := NewDiagnosticHandler(nil)
	for _, test := range []struct {
		name, body string
		handle     http.HandlerFunc
	}{
		{"malformed start", "{", h.Start}, {"unknown start", `{"unknown":1}`, h.Start}, {"trailing start", `{} {}`, h.Start},
		{"malformed answer", "{", h.Answer}, {"unknown answer", `{"unexpected":true}`, h.Answer}, {"trailing answer", `{} {}`, h.Answer},
		{"oversized answer", `{"student_answer":"` + strings.Repeat("a", 17000) + `"}`, h.Answer},
	} {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			req = req.WithContext(context.WithValue(req.Context(), UserContextKey, &domain.User{ID: "test", Role: domain.RoleStudent, GradeLevel: 4}))
			response := httptest.NewRecorder()
			test.handle(response, req)
			if response.Code != http.StatusBadRequest || strings.Contains(response.Body.String(), "student_answer") {
				t.Fatalf("invalid request: %d %s", response.Code, response.Body.String())
			}
		})
	}
}
