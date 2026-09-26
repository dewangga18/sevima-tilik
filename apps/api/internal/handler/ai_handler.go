package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
	"github.com/sevima/tilik-api/internal/service"
)

type AIHandler struct {
	settings  *service.AISettingsService
	generator *service.AIGeneratorService
}

func NewAIHandler(settings *service.AISettingsService, generator *service.AIGeneratorService) *AIHandler {
	return &AIHandler{settings: settings, generator: generator}
}
func decodeAI(w http.ResponseWriter, r *http.Request, value any) bool {
	w.Header().Set("Cache-Control", "no-store")
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10))
	d.DisallowUnknownFields()
	if d.Decode(value) != nil {
		WriteError(w, 400, "Input AI tidak valid")
		return false
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		WriteError(w, 400, "Input AI tidak valid")
		return false
	}
	return true
}
func (h *AIHandler) Settings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	v, err := h.settings.View(r.Context())
	if err != nil {
		aiError(w, err)
		return
	}
	WriteJSON(w, 200, v)
}
func (h *AIHandler) SaveSettings(w http.ResponseWriter, r *http.Request) {
	var input struct {
		APIKey string `json:"api_key"`
		Model  string `json:"model"`
	}
	if !decodeAI(w, r, &input) {
		return
	}
	v, err := h.settings.Save(r.Context(), input.APIKey, input.Model)
	if err != nil {
		aiError(w, err)
		return
	}
	WriteJSON(w, 200, v)
}
func (h *AIHandler) DeleteSettings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	v, err := h.settings.Delete(r.Context())
	if err != nil {
		aiError(w, err)
		return
	}
	WriteJSON(w, 200, v)
}
func (h *AIHandler) Generate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	user, ok := r.Context().Value(UserContextKey).(*domain.User)
	if !ok || user == nil {
		WriteError(w, 401, "Sesi tidak valid")
		return
	}
	if user.Role != domain.RoleAdmin && user.Role != domain.RoleTeacher {
		WriteError(w, 403, "Generator hanya tersedia untuk guru dan admin")
		return
	}
	var input service.AIGenerateInput
	if !decodeAI(w, r, &input) {
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 28*time.Second)
	defer cancel()
	response, err := h.generator.Generate(ctx, user.ID, input)
	if err != nil {
		aiError(w, err)
		return
	}
	WriteJSON(w, 200, response)
}
func aiError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrAIInput):
		WriteError(w, 400, "Input tidak valid. Pilih skill kelas 4, level 1–2, dan tujuan soal yang tersedia.")
	case errors.Is(err, repository.ErrAIRequestConflict):
		WriteError(w, 409, "Permintaan sedang diproses atau ID dipakai untuk input lain. Tunggu lalu coba lagi.")
	case errors.Is(err, service.ErrAIUnavailable):
		WriteError(w, 503, "Gemini belum siap. Minta admin mengatur API key dan kunci enkripsi storage.")
	case errors.Is(err, service.ErrAIProvider):
		WriteError(w, 502, "Gemini menolak key atau model. Admin perlu memeriksa pengaturan AI.")
	case errors.Is(err, service.ErrAIOutput):
		WriteError(w, 502, "Hasil Gemini tidak lolos validasi. Tidak ada draft disimpan; coba lagi.")
	case errors.Is(err, service.ErrAIQuota):
		WriteError(w, 503, "Gemini sedang tidak tersedia atau kuota habis. Coba lagi nanti.")
	case errors.Is(err, service.ErrAIRate):
		w.Header().Set("Retry-After", "60")
		WriteError(w, 429, "Batas permintaan AI tercapai. Tunggu satu menit lalu coba lagi.")
	case errors.Is(err, context.DeadlineExceeded):
		WriteError(w, 504, "Gemini melewati batas waktu. Coba lagi dengan permintaan yang sama.")
	default:
		WriteError(w, 500, "Operasi AI belum bisa disimpan. Silakan coba lagi.")
	}
}
