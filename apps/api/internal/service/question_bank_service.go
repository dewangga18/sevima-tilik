package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
)

var (
	ErrBankCandidateNotFound = errors.New("question candidate not found")
	ErrBankReviewInvalid     = errors.New("invalid question bank review request")
	ErrBankAlreadyApproved   = errors.New("candidate is already approved")
	ErrBankCannotRevoke      = errors.New("approved candidate cannot be revoked here")
)

// BankNotActivatable carries the admin-facing reason a candidate cannot be
// activated yet, so the handler can show the rule instead of a generic failure.
type BankNotActivatableError struct {
	Reason string
}

func (e *BankNotActivatableError) Error() string {
	return "candidate cannot be activated: " + e.Reason
}

// maxCandidateListLimit bounds one review page so a full draft bank cannot
// force an unbounded response.
const maxCandidateListLimit = 100

// QuestionBankService exposes the already-validated import path to admins.
// Approvals deliberately reuse RegisterQuestionBank so activation keeps the
// same guarantees as the local import command: transaction, advisory lock,
// content-hash immutability, duplicate-prompt downgrade, and refusal to
// overwrite a runtime question id.
type QuestionBankService struct {
	db *repository.DB
}

func NewQuestionBankService(db *repository.DB) *QuestionBankService {
	return &QuestionBankService{db: db}
}

func (s *QuestionBankService) Summary(ctx context.Context) (*domain.QuestionBankSummary, error) {
	summary, err := s.db.QuestionBankSummary(ctx)
	if err != nil {
		return nil, fmt.Errorf("question bank summary: %w", err)
	}
	return summary, nil
}

// List returns stored candidates for one review status. Only drafts are
// reviewable through the UI; approved and rejected rows stay readable for
// audit but cannot be re-decided.
func (s *QuestionBankService) List(ctx context.Context, status string, limit int) ([]domain.CandidateReviewView, error) {
	if status != "draft" && status != "approved" && status != "rejected" {
		return nil, ErrBankReviewInvalid
	}
	if limit <= 0 || limit > maxCandidateListLimit {
		limit = maxCandidateListLimit
	}
	records, err := s.db.ListCandidates(ctx, status, limit)
	if err != nil {
		return nil, fmt.Errorf("list candidates: %w", err)
	}

	views := make([]domain.CandidateReviewView, 0, len(records))
	skillReady := make(map[string]bool)
	for _, record := range records {
		view := domain.CandidateReviewView{QuestionCandidateRecord: record}
		ready, cached := skillReady[record.Candidate.Skill]
		if !cached {
			ready, err = s.db.SkillReadyForActivation(ctx, record.Candidate.Skill)
			if err != nil {
				return nil, err
			}
			skillReady[record.Candidate.Skill] = ready
		}
		view.Eligible, view.BlockedReason = activationBlock(record.Candidate, ready)
		views = append(views, view)
	}
	return views, nil
}

// activationBlock mirrors the guards inside RegisterQuestionBank so the UI can
// state the same rule the import path enforces.
func activationBlock(candidate domain.QuestionCandidate, skillReady bool) (bool, string) {
	switch {
	case candidate.Grade != 4:
		return false, fmt.Sprintf("Kelas %d belum punya konten dan alur belajar", candidate.Grade)
	case candidate.Difficulty > 2:
		return false, fmt.Sprintf("Level %d di luar jangkauan aktivasi saat ini (1–2)", candidate.Difficulty)
	case !skillReady:
		return false, "Skill ini belum punya micro lesson, jadi soal belum bisa diaktifkan"
	}
	return true, ""
}

func (s *QuestionBankService) Reject(ctx context.Context, id, note string) (*domain.QuestionCandidateRecord, error) {
	id = strings.TrimSpace(id)
	note = strings.TrimSpace(note)
	if id == "" || note == "" {
		return nil, ErrBankReviewInvalid
	}
	if err := s.db.SetCandidateStatus(ctx, id, "rejected", note); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBankCandidateNotFound
		}
		if errors.Is(err, repository.ErrCandidateStatusConflict) {
			return nil, ErrBankCannotRevoke
		}
		return nil, fmt.Errorf("reject candidate: %w", err)
	}
	return s.reload(ctx, id)
}

// Approve activates a draft candidate. The stored payload and stored content
// hash are reused verbatim: re-parsing from the source file could produce a
// different hash and trip the immutability guard for the wrong reason.
func (s *QuestionBankService) Approve(ctx context.Context, id, note string) (*domain.BankImportReport, error) {
	id = strings.TrimSpace(id)
	note = strings.TrimSpace(note)
	if id == "" {
		return nil, ErrBankReviewInvalid
	}

	record, err := s.db.CandidateByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load candidate: %w", err)
	}
	if record == nil {
		return nil, ErrBankCandidateNotFound
	}
	if record.Status == "approved" {
		return nil, ErrBankAlreadyApproved
	}
	if record.Status == "rejected" {
		return nil, fmt.Errorf("%w: kandidat yang sudah ditolak perlu ditinjau ulang di luar dashboard", ErrBankReviewInvalid)
	}

	ready, err := s.db.SkillReadyForActivation(ctx, record.Candidate.Skill)
	if err != nil {
		return nil, err
	}
	if eligible, reason := activationBlock(record.Candidate, ready); !eligible {
		return nil, &BankNotActivatableError{Reason: reason}
	}

	options := make([]string, 0, len(record.Candidate.Options))
	for _, option := range record.Candidate.Options {
		options = append(options, option.Text)
	}
	entry := domain.BankEntry{
		Candidate:  record.Candidate,
		Hash:       record.ContentHash,
		SourceFile: record.SourceFile,
		Payload:    record.Payload,
		Status:     "approved",
		Note:       reviewNote(note),
		Question: domain.Question{
			ID:            record.ID,
			SkillID:       record.Candidate.Skill,
			Difficulty:    record.Candidate.Difficulty,
			Prompt:        record.Candidate.Prompt,
			Options:       options,
			AnswerKey:     candidateAnswer(record.Candidate),
			Explanation:   record.Candidate.Explanation,
			Misconception: "",
		},
	}

	report, err := s.db.RegisterQuestionBank(ctx, []domain.BankEntry{entry}, true)
	if err != nil {
		return nil, fmt.Errorf("approve candidate: %w", err)
	}
	return report, nil
}

// UsedInAssessment lets the review UI warn that a question already reached a
// student, which is why its content must stay immutable.
func (s *QuestionBankService) UsedInAssessment(ctx context.Context, id string) (bool, error) {
	used, err := s.db.QuestionUsedInAssessment(ctx, id)
	if err != nil {
		return false, fmt.Errorf("question usage: %w", err)
	}
	return used, nil
}

func (s *QuestionBankService) reload(ctx context.Context, id string) (*domain.QuestionCandidateRecord, error) {
	record, err := s.db.CandidateByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("reload candidate: %w", err)
	}
	if record == nil {
		return nil, ErrBankCandidateNotFound
	}
	return record, nil
}

func reviewNote(note string) string {
	if note == "" {
		return "Disetujui lewat dashboard admin."
	}
	return "Disetujui lewat dashboard admin. " + note
}
