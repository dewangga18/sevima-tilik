package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/service"
)

// adminBodyLimit matches the auth limit; admin payloads carry only short
// identity and placement fields, never curriculum or assessment content.
const adminBodyLimit = 4 << 10

type AdminHandler struct {
	adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.adminService.ListUsers(r.Context())
	if err != nil {
		log.Printf("admin list users: %v", err)
		WriteError(w, http.StatusInternalServerError, "Data akun belum bisa dimuat. Silakan coba lagi.")
		return
	}
	WriteJSON(w, http.StatusOK, users)
}

type createUserRequest struct {
	Email      string      `json:"email"`
	Name       string      `json:"name"`
	Role       domain.Role `json:"role"`
	GradeLevel int         `json:"grade_level"`
	Password   string      `json:"password"`
}

func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if !decodeAdminBody(w, r, &req) {
		return
	}

	user, err := h.adminService.CreateUser(r.Context(), service.NewUserInput{
		Email:      req.Email,
		Name:       req.Name,
		Role:       req.Role,
		GradeLevel: req.GradeLevel,
		Password:   req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAdminInputInvalid):
			WriteError(w, http.StatusBadRequest, "Data akun tidak valid. Periksa email, nama, role, kelas, dan panjang kata sandi.")
		case errors.Is(err, service.ErrAdminEmailTaken):
			WriteError(w, http.StatusConflict, "Email sudah digunakan akun lain")
		default:
			log.Printf("admin create user: %v", err)
			WriteError(w, http.StatusInternalServerError, "Akun belum bisa dibuat. Silakan coba lagi.")
		}
		return
	}
	WriteJSON(w, http.StatusCreated, user)
}

type updateUserRequest struct {
	Role     *domain.Role `json:"role"`
	IsActive *bool        `json:"is_active"`
}

func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	actor := authenticatedUser(r)
	var req updateUserRequest
	if !decodeAdminBody(w, r, &req) {
		return
	}

	user, err := h.adminService.UpdateUser(r.Context(), actor.ID, r.PathValue("id"), service.UpdateUserInput{
		Role:     req.Role,
		IsActive: req.IsActive,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAdminInputInvalid):
			WriteError(w, http.StatusBadRequest, "Perubahan akun tidak valid. Pilih role yang tersedia dan setidaknya satu perubahan.")
		case errors.Is(err, service.ErrAdminSelfChange):
			WriteError(w, http.StatusConflict, "Akun sendiri tidak dapat diubah role atau dinonaktifkan")
		case errors.Is(err, service.ErrAdminUserNotFound):
			WriteError(w, http.StatusNotFound, "Akun tidak ditemukan")
		default:
			log.Printf("admin update user %s: %v", r.PathValue("id"), err)
			WriteError(w, http.StatusInternalServerError, "Akun belum bisa diperbarui. Silakan coba lagi.")
		}
		return
	}
	WriteJSON(w, http.StatusOK, user)
}

func (h *AdminHandler) ListClasses(w http.ResponseWriter, r *http.Request) {
	classes, err := h.adminService.ListClasses(r.Context())
	if err != nil {
		log.Printf("admin list classes: %v", err)
		WriteError(w, http.StatusInternalServerError, "Data kelas belum bisa dimuat. Silakan coba lagi.")
		return
	}
	WriteJSON(w, http.StatusOK, classes)
}

type createClassRequest struct {
	Name       string `json:"name"`
	GradeLevel int    `json:"grade_level"`
}

func (h *AdminHandler) CreateClass(w http.ResponseWriter, r *http.Request) {
	var req createClassRequest
	if !decodeAdminBody(w, r, &req) {
		return
	}

	class, err := h.adminService.CreateClass(r.Context(), service.NewClassInput{Name: req.Name, GradeLevel: req.GradeLevel})
	if err != nil {
		if errors.Is(err, service.ErrAdminInputInvalid) {
			WriteError(w, http.StatusBadRequest, "Data kelas tidak valid. Periksa nama dan kelas.")
			return
		}
		log.Printf("admin create class: %v", err)
		WriteError(w, http.StatusInternalServerError, "Kelas belum bisa dibuat. Silakan coba lagi.")
		return
	}
	WriteJSON(w, http.StatusCreated, class)
}

func (h *AdminHandler) GetRoster(w http.ResponseWriter, r *http.Request) {
	roster, err := h.adminService.Roster(r.Context(), r.PathValue("class_id"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrAdminInputInvalid):
			WriteError(w, http.StatusBadRequest, "Permintaan kelas tidak valid")
		case errors.Is(err, service.ErrAdminClassNotFound):
			WriteError(w, http.StatusNotFound, "Kelas tidak ditemukan")
		default:
			log.Printf("admin roster %s: %v", r.PathValue("class_id"), err)
			WriteError(w, http.StatusInternalServerError, "Data peserta kelas belum bisa dimuat. Silakan coba lagi.")
		}
		return
	}
	WriteJSON(w, http.StatusOK, roster)
}

type placementRequest struct {
	UserID string `json:"user_id"`
}

func (h *AdminHandler) SetEnrollment(w http.ResponseWriter, r *http.Request) {
	var req placementRequest
	if !decodeAdminBody(w, r, &req) {
		return
	}

	if err := h.adminService.SetEnrollment(r.Context(), r.PathValue("class_id"), req.UserID, true); err != nil {
		h.writePlacementError(w, r, err, "siswa")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "Siswa ditambahkan ke kelas"})
}

func (h *AdminHandler) RemoveEnrollment(w http.ResponseWriter, r *http.Request) {
	var req placementRequest
	if !decodeAdminBody(w, r, &req) {
		return
	}

	if err := h.adminService.SetEnrollment(r.Context(), r.PathValue("class_id"), req.UserID, false); err != nil {
		h.writePlacementError(w, r, err, "siswa")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "Siswa dikeluarkan dari kelas"})
}

func (h *AdminHandler) SetAssignment(w http.ResponseWriter, r *http.Request) {
	var req placementRequest
	if !decodeAdminBody(w, r, &req) {
		return
	}

	if err := h.adminService.SetAssignment(r.Context(), r.PathValue("class_id"), req.UserID, true); err != nil {
		h.writePlacementError(w, r, err, "guru")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "Guru ditugaskan ke kelas"})
}

func (h *AdminHandler) RemoveAssignment(w http.ResponseWriter, r *http.Request) {
	var req placementRequest
	if !decodeAdminBody(w, r, &req) {
		return
	}

	if err := h.adminService.SetAssignment(r.Context(), r.PathValue("class_id"), req.UserID, false); err != nil {
		h.writePlacementError(w, r, err, "guru")
		return
	}
	WriteJSON(w, http.StatusOK, map[string]string{"message": "Guru dilepas dari kelas"})
}

func (h *AdminHandler) writePlacementError(w http.ResponseWriter, r *http.Request, err error, subject string) {
	switch {
	case errors.Is(err, service.ErrAdminInputInvalid):
		WriteError(w, http.StatusBadRequest, "Permintaan penempatan tidak valid")
	case errors.Is(err, service.ErrAdminClassNotFound):
		WriteError(w, http.StatusNotFound, "Kelas tidak ditemukan")
	case errors.Is(err, service.ErrAdminUserNotFound):
		WriteError(w, http.StatusBadRequest, "Akun dengan role "+subject+" tidak ditemukan")
	default:
		log.Printf("admin placement %s: %v", r.URL.Path, err)
		WriteError(w, http.StatusInternalServerError, "Penempatan belum bisa disimpan. Silakan coba lagi.")
	}
}

// decodeAdminBody rejects unknown fields and trailing payloads so a client
// mistake surfaces as 400 instead of being silently ignored.
func decodeAdminBody(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, adminBodyLimit))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		WriteError(w, http.StatusBadRequest, "Data permintaan tidak valid")
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		WriteError(w, http.StatusBadRequest, "Data permintaan tidak valid")
		return false
	}
	return true
}
