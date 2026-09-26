package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/sevima/tilik-api/internal/service"
)

// QuestionBankHandler serves the admin question bank review. Every route here
// returns candidate payloads that include correct_answer, so all of them stay
// behind AdminMiddleware in every environment.
type QuestionBankHandler struct {
	bankService *service.QuestionBankService
}

func NewQuestionBankHandler(bankService *service.QuestionBankService) *QuestionBankHandler {
	return &QuestionBankHandler{bankService: bankService}
}

func (h *QuestionBankHandler) Summary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.bankService.Summary(r.Context())
	if err != nil {
		log.Printf("admin question bank summary: %v", err)
		WriteError(w, http.StatusInternalServerError, "Ringkasan bank soal belum bisa dimuat. Silakan coba lagi.")
		return
	}
	WriteJSON(w, http.StatusOK, summary)
}

func (h *QuestionBankHandler) ListCandidates(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	if status == "" {
		status = "draft"
	}
	limit := 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "Parameter limit tidak valid")
			return
		}
		limit = parsed
	}

	records, err := h.bankService.List(r.Context(), status, limit)
	if err != nil {
		if errors.Is(err, service.ErrBankReviewInvalid) {
			WriteError(w, http.StatusBadRequest, "Status kandidat harus draft, approved, atau rejected")
			return
		}
		log.Printf("admin list candidates %s: %v", status, err)
		WriteError(w, http.StatusInternalServerError, "Daftar kandidat belum bisa dimuat. Silakan coba lagi.")
		return
	}
	WriteJSON(w, http.StatusOK, records)
}

func (h *QuestionBankHandler) Approve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Note string `json:"note"`
	}
	if !decodeAdminBody(w, r, &req) {
		return
	}

	report, err := h.bankService.Approve(r.Context(), id, strings.TrimSpace(req.Note))
	if err != nil {
		h.writeReviewError(w, r, err)
		return
	}

	// A duplicate prompt is downgraded to draft by the import path; report that
	// honestly instead of presenting it as a successful activation.
	if len(report.Items) == 1 && report.Items[0].Status != "approved" {
		WriteJSON(w, http.StatusConflict, map[string]any{
			"activated": false,
			"report":    report,
		})
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{"activated": true, "report": report})
}

func (h *QuestionBankHandler) Reject(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req struct {
		Note string `json:"note"`
	}
	if !decodeAdminBody(w, r, &req) {
		return
	}

	record, err := h.bankService.Reject(r.Context(), id, strings.TrimSpace(req.Note))
	if err != nil {
		h.writeReviewError(w, r, err)
		return
	}
	WriteJSON(w, http.StatusOK, record)
}

func (h *QuestionBankHandler) writeReviewError(w http.ResponseWriter, r *http.Request, err error) {
	var notActivatable *service.BankNotActivatableError
	switch {
	case errors.As(err, &notActivatable):
		WriteError(w, http.StatusConflict, notActivatable.Reason)
	case errors.Is(err, service.ErrBankReviewInvalid):
		WriteError(w, http.StatusBadRequest, "Permintaan review tidak valid. Isi alasan saat menolak kandidat.")
	case errors.Is(err, service.ErrBankCandidateNotFound):
		WriteError(w, http.StatusNotFound, "Kandidat tidak ditemukan")
	case errors.Is(err, service.ErrBankAlreadyApproved):
		WriteError(w, http.StatusConflict, "Kandidat sudah disetujui sebelumnya")
	case errors.Is(err, service.ErrBankCannotRevoke):
		WriteError(w, http.StatusConflict, "Kandidat sudah disetujui dan tidak bisa ditolak dari sini")
	default:
		log.Printf("admin question bank review %s: %v", r.URL.Path, err)
		WriteError(w, http.StatusInternalServerError, "Keputusan review belum bisa disimpan. Silakan coba lagi.")
	}
}
