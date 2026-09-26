package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/service"
)

type DiagnosticHandler struct {
	diagService *service.DiagnosticService
}

func NewDiagnosticHandler(diagService *service.DiagnosticService) *DiagnosticHandler {
	return &DiagnosticHandler{diagService: diagService}
}

type StartDiagnosticRequest struct {
	GradeLevel int `json:"grade_level"`
}

func (h *DiagnosticHandler) Start(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserContextKey).(*domain.User)
	if !ok || user == nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	grade := user.GradeLevel
	if grade == 0 {
		grade = 4
	}

	var req StartDiagnosticRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err == nil && req.GradeLevel > 0 {
		grade = req.GradeLevel
	}

	assessment, err := h.diagService.StartDiagnostic(r.Context(), user.ID, grade)
	if err != nil {
		log.Printf("start diagnostic for user %s: %v", user.ID, err)
		WriteError(w, http.StatusInternalServerError, "Gagal memulai tes diagnostik. Silakan coba lagi.")
		return
	}

	WriteJSON(w, http.StatusOK, assessment)
}

type SubmitDiagnosticRequest struct {
	AssessmentID string                     `json:"assessment_id"`
	Answers      []service.AnswerSubmission `json:"answers"`
}

func (h *DiagnosticHandler) Submit(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserContextKey).(*domain.User)
	if !ok || user == nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req SubmitDiagnosticRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.AssessmentID == "" {
		WriteError(w, http.StatusBadRequest, "Assessment ID is required")
		return
	}

	assessment, err := h.diagService.SubmitDiagnostic(r.Context(), user.ID, req.AssessmentID, req.Answers)
	if err != nil {
		if errors.Is(err, service.ErrForbidden) {
			WriteError(w, http.StatusForbidden, "Akses ditolak")
			return
		}
		if errors.Is(err, service.ErrAssessmentNotFound) {
			WriteError(w, http.StatusNotFound, "Asesmen tidak ditemukan")
			return
		}
		if errors.Is(err, service.ErrAssessmentCompleted) {
			WriteError(w, http.StatusConflict, "Asesmen sudah selesai. Jawaban tidak dapat diubah.")
			return
		}
		log.Printf("submit diagnostic %s for user %s: %v", req.AssessmentID, user.ID, err)
		WriteError(w, http.StatusInternalServerError, "Gagal mengevaluasi asesmen. Silakan coba lagi.")
		return
	}

	WriteJSON(w, http.StatusOK, assessment)
}

func (h *DiagnosticHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserContextKey).(*domain.User)
	if !ok || user == nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	assessment, err := h.diagService.GetLatestAssessment(r.Context(), user.ID)
	if err != nil {
		log.Printf("load latest diagnostic for user %s: %v", user.ID, err)
		WriteError(w, http.StatusInternalServerError, "Gagal memuat asesmen")
		return
	}

	if assessment == nil {
		WriteJSON(w, http.StatusOK, nil)
		return
	}

	WriteJSON(w, http.StatusOK, assessment)
}

func (h *DiagnosticHandler) GetSkills(w http.ResponseWriter, r *http.Request) {
	skills, err := h.diagService.GetSkills(r.Context(), 4)
	if err != nil {
		log.Printf("load diagnostic skills: %v", err)
		WriteError(w, http.StatusInternalServerError, "Gagal memuat daftar skill")
		return
	}

	WriteJSON(w, http.StatusOK, skills)
}

func (h *DiagnosticHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserContextKey).(*domain.User)
	if !ok || user == nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	history, err := h.diagService.GetHistory(r.Context(), user.ID)
	if err != nil {
		log.Printf("load assessment history for user %s: %v", user.ID, err)
		WriteError(w, http.StatusInternalServerError, "Gagal memuat riwayat belajar. Silakan coba lagi.")
		return
	}
	WriteJSON(w, http.StatusOK, history)
}

func (h *DiagnosticHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(UserContextKey).(*domain.User)
	if !ok || user == nil {
		WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}
	assessment, err := h.diagService.GetAssessment(r.Context(), user.ID, r.PathValue("id"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrForbidden):
			WriteError(w, http.StatusForbidden, "Akses ditolak")
		case errors.Is(err, service.ErrAssessmentNotFound):
			WriteError(w, http.StatusNotFound, "Asesmen tidak ditemukan")
		default:
			log.Printf("load assessment %s for user %s: %v", r.PathValue("id"), user.ID, err)
			WriteError(w, http.StatusInternalServerError, "Gagal memuat asesmen. Silakan coba lagi.")
		}
		return
	}
	WriteJSON(w, http.StatusOK, assessment)
}
