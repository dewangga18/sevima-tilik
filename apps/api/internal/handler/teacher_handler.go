package handler

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/service"
)

type TeacherHandler struct {
	teacherService *service.TeacherService
}

func NewTeacherHandler(teacherService *service.TeacherService) *TeacherHandler {
	return &TeacherHandler{teacherService: teacherService}
}

func (h *TeacherHandler) GetClasses(w http.ResponseWriter, r *http.Request) {
	user := authenticatedUser(r)
	classes, err := h.teacherService.GetClasses(r.Context(), user.ID)
	if err != nil {
		log.Printf("teacher classes: %v", err)
		WriteError(w, http.StatusInternalServerError, "Data kelas belum bisa dimuat. Silakan coba lagi.")
		return
	}
	WriteJSON(w, http.StatusOK, classes)
}

func (h *TeacherHandler) GetClassStudents(w http.ResponseWriter, r *http.Request) {
	user := authenticatedUser(r)
	classroomID := r.PathValue("class_id")
	if classroomID == "" {
		WriteError(w, http.StatusBadRequest, "Permintaan kelas tidak valid")
		return
	}
	students, err := h.teacherService.GetClassStudents(r.Context(), user.ID, classroomID)
	if errors.Is(err, service.ErrClassForbidden) {
		WriteError(w, http.StatusForbidden, "Akses ditolak")
		return
	}
	if err != nil {
		log.Printf("class students: %v", err)
		WriteError(w, http.StatusInternalServerError, "Data siswa belum bisa dimuat. Silakan coba lagi.")
		return
	}
	WriteJSON(w, http.StatusOK, students)
}

func (h *TeacherHandler) GetStudentInsight(w http.ResponseWriter, r *http.Request) {
	user := authenticatedUser(r)
	studentID := r.PathValue("student_id")
	if studentID == "" {
		WriteError(w, http.StatusBadRequest, "Permintaan siswa tidak valid")
		return
	}
	insight, err := h.teacherService.GetStudentInsight(r.Context(), user.ID, studentID)
	if errors.Is(err, service.ErrClassForbidden) {
		WriteError(w, http.StatusForbidden, "Akses ditolak")
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		WriteError(w, http.StatusNotFound, "Siswa tidak ditemukan")
		return
	}
	if err != nil {
		log.Printf("student insight: %v", err)
		WriteError(w, http.StatusInternalServerError, "Insight siswa belum bisa dimuat. Silakan coba lagi.")
		return
	}
	WriteJSON(w, http.StatusOK, insight)
}

func authenticatedUser(r *http.Request) *domain.User {
	user, ok := r.Context().Value(UserContextKey).(*domain.User)
	if !ok || user == nil {
		return &domain.User{}
	}
	return user
}
