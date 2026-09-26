package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
	"github.com/sevima/tilik-api/internal/service"
)

type LearningHandler struct{ learning *service.LearningService }

func NewLearningHandler(learning *service.LearningService) *LearningHandler {
	return &LearningHandler{learning: learning}
}

func learningUser(w http.ResponseWriter, r *http.Request) *domain.User {
	user, ok := r.Context().Value(UserContextKey).(*domain.User)
	if !ok || user == nil {
		WriteError(w, http.StatusUnauthorized, "Sesi tidak valid")
		return nil
	}
	return user
}
func learningError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrForbidden):
		WriteError(w, http.StatusForbidden, "Akses ditolak")
	case errors.Is(err, service.ErrLearningNotFound), errors.Is(err, service.ErrAssessmentNotFound):
		WriteError(w, http.StatusNotFound, "Aktivitas atau materi tidak ditemukan")
	case errors.Is(err, service.ErrInvalidLearningRequest):
		WriteError(w, http.StatusBadRequest, "Permintaan belajar tidak valid")
	case errors.Is(err, repository.ErrLearningBankEmpty):
		WriteError(w, http.StatusConflict, "Soal baru untuk skill ini belum cukup. Pilih rekomendasi lain atau coba setelah materi ditambah.")
	case errors.Is(err, repository.ErrLearningChanged):
		WriteError(w, http.StatusConflict, "Aktivitas atau jawaban sudah berubah. Buka kembali dari beranda.")
	default:
		log.Printf("learning operation: %v", err)
		WriteError(w, http.StatusInternalServerError, "Data belajar belum bisa diproses. Silakan coba lagi.")
	}
}
func decodeLearningRequest(w http.ResponseWriter, r *http.Request, target any) bool {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 16<<10))
	body = bytes.TrimSpace(body)
	if err != nil || len(body) == 0 || body[0] != '{' {
		WriteError(w, http.StatusBadRequest, "Permintaan belajar tidak valid")
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(target); err != nil {
		WriteError(w, http.StatusBadRequest, "Permintaan belajar tidak valid")
		return false
	}
	var extra any
	if err = decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		WriteError(w, http.StatusBadRequest, "Permintaan belajar tidak valid")
		return false
	}
	return true
}
func (h *LearningHandler) Progress(w http.ResponseWriter, r *http.Request) {
	user := learningUser(w, r)
	if user == nil {
		return
	}
	progress, err := h.learning.Progress(r.Context(), user.ID)
	if err != nil {
		learningError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, progress)
}
func (h *LearningHandler) Lesson(w http.ResponseWriter, r *http.Request) {
	if learningUser(w, r) == nil {
		return
	}
	lesson, err := h.learning.Lesson(r.Context(), r.PathValue("skill_id"))
	if err != nil {
		learningError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, lesson)
}
func (h *LearningHandler) Session(w http.ResponseWriter, r *http.Request) {
	user := learningUser(w, r)
	if user == nil {
		return
	}
	session, err := h.learning.Session(r.Context(), user.ID, r.PathValue("id"))
	if err != nil {
		learningError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, session)
}
func (h *LearningHandler) Start(w http.ResponseWriter, r *http.Request) {
	user := learningUser(w, r)
	if user == nil {
		return
	}
	var req struct {
		SourceAssessmentID string `json:"source_assessment_id"`
		SkillID            string `json:"skill_id"`
		RequestID          string `json:"request_id"`
	}
	if !decodeLearningRequest(w, r, &req) {
		return
	}
	session, err := h.learning.Start(r.Context(), user.ID, req.SourceAssessmentID, req.SkillID, req.RequestID)
	if err != nil {
		learningError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, session)
}
func (h *LearningHandler) CompleteLesson(w http.ResponseWriter, r *http.Request) {
	user := learningUser(w, r)
	if user == nil {
		return
	}
	var req struct {
		SessionID string `json:"session_id"`
	}
	if !decodeLearningRequest(w, r, &req) {
		return
	}
	session, err := h.learning.CompleteLesson(r.Context(), user.ID, req.SessionID)
	if err != nil {
		learningError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, session)
}
func (h *LearningHandler) Answer(w http.ResponseWriter, r *http.Request) {
	user := learningUser(w, r)
	if user == nil {
		return
	}
	var req struct {
		SessionID     string `json:"session_id"`
		QuestionID    string `json:"question_id"`
		StudentAnswer string `json:"student_answer"`
	}
	if !decodeLearningRequest(w, r, &req) {
		return
	}
	session, err := h.learning.Answer(r.Context(), user.ID, req.SessionID, req.QuestionID, req.StudentAnswer)
	if err != nil {
		learningError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, session)
}
