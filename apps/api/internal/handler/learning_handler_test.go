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

func TestLearningDatabaseErrorsAreSafe(t *testing.T) {
	sqlDB, err := sql.Open("pgx", "postgres://unused")
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.Close()
	db := &repository.DB{DB: sqlDB}
	curriculum := repository.NewCurriculumRepository(db)
	diagnostic := service.NewDiagnosticService(curriculum, repository.NewAssessmentRepository(db))
	h := NewLearningHandler(service.NewLearningService(repository.NewLearningRepository(db), curriculum, diagnostic))
	for _, test := range []struct {
		name, body string
		handle     http.HandlerFunc
	}{
		{"progress", "", h.Progress}, {"lesson", "", h.Lesson}, {"session", "", h.Session},
		{"start", `{"source_assessment_id":"source","skill_id":"mul_basic","request_id":"retry"}`, h.Start},
		{"lesson-complete", `{"session_id":"session"}`, h.CompleteLesson},
		{"answer", `{"session_id":"session","question_id":"q","student_answer":"10"}`, h.Answer},
	} {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.body))
			req.SetPathValue("id", "session")
			req.SetPathValue("skill_id", "mul_basic")
			req = req.WithContext(context.WithValue(req.Context(), UserContextKey, &domain.User{ID: "student", Role: domain.RoleStudent}))
			response := httptest.NewRecorder()
			test.handle(response, req)
			var body APIResponse
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if response.Code != http.StatusInternalServerError || body.Error != "Data belajar belum bisa diproses. Silakan coba lagi." {
				t.Fatalf("unexpected public error: %d %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestLearningWritePayloadValidation(t *testing.T) {
	h := NewLearningHandler(nil)
	for name, handle := range map[string]http.HandlerFunc{"start": h.Start, "lesson-complete": h.CompleteLesson, "answer": h.Answer} {
		for _, body := range []string{"null", "[]", "{", `{"unknown":true}`, `{} {}`, `{"unknown":"` + strings.Repeat("a", 17000) + `"}`} {
			t.Run(name+"/"+body[:min(len(body), 20)], func(t *testing.T) {
				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
				req = req.WithContext(context.WithValue(req.Context(), UserContextKey, &domain.User{ID: "student", Role: domain.RoleStudent}))
				response := httptest.NewRecorder()
				handle(response, req)
				if response.Code != http.StatusBadRequest {
					t.Fatalf("invalid payload accepted: %d", response.Code)
				}
			})
		}
	}
}
