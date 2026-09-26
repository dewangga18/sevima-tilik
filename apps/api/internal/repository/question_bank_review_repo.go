package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sevima/tilik-api/internal/domain"
)

var ErrCandidateStatusConflict = errors.New("candidate status does not allow this change")

// scanCandidate reads one stored candidate row and decodes its payload.
func scanCandidate(scan func(dest ...any) error) (*domain.QuestionCandidateRecord, error) {
	var record domain.QuestionCandidateRecord
	var payload []byte
	if err := scan(&record.ID, &record.Status, &record.ReviewNote, &record.SourceFile, &record.ContentHash, &record.RegisteredAt, &payload); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(payload, &record.Candidate); err != nil {
		return nil, fmt.Errorf("decode candidate payload %s: %w", record.ID, err)
	}
	record.Payload = json.RawMessage(payload)
	return &record, nil
}

const candidateColumns = `id,status,review_note,source_file,content_hash,registered_at,payload`

// ListCandidates returns stored candidates with a given review status. The
// status filter is validated by the service before reaching this query.
func (db *DB) ListCandidates(ctx context.Context, status string, limit int) ([]domain.QuestionCandidateRecord, error) {
	rows, err := db.QueryContext(ctx, `SELECT `+candidateColumns+` FROM question_candidates WHERE status=$1 ORDER BY registered_at DESC, id DESC LIMIT $2`, status, limit)
	if err != nil {
		return nil, fmt.Errorf("list candidates: %w", err)
	}
	defer rows.Close()
	records := make([]domain.QuestionCandidateRecord, 0)
	for rows.Next() {
		record, err := scanCandidate(rows.Scan)
		if err != nil {
			return nil, err
		}
		records = append(records, *record)
	}
	return records, rows.Err()
}

func (db *DB) CandidateByID(ctx context.Context, id string) (*domain.QuestionCandidateRecord, error) {
	record, err := scanCandidate(db.QueryRowContext(ctx, `SELECT `+candidateColumns+` FROM question_candidates WHERE id=$1`, id).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("candidate by id: %w", err)
	}
	return record, nil
}

// SetCandidateStatus records a review decision. Rejecting an already approved
// candidate is refused so historical activation cannot be silently revoked
// through this path; the guard mirrors the import rule.
func (db *DB) SetCandidateStatus(ctx context.Context, id, status, note string) error {
	var current string
	err := db.QueryRowContext(ctx, `SELECT status FROM question_candidates WHERE id=$1`, id).Scan(&current)
	if errors.Is(err, sql.ErrNoRows) {
		return sql.ErrNoRows
	}
	if err != nil {
		return err
	}
	if current == "approved" && status != "approved" {
		return ErrCandidateStatusConflict
	}
	_, err = db.ExecContext(ctx, `UPDATE question_candidates SET status=$1, review_note=$2 WHERE id=$3`, status, note, id)
	return err
}

// SkillReadyForActivation mirrors the readiness rule used by the import path:
// only skills that already have a lesson and belong to grade 4 or lower can
// receive activated questions.
func (db *DB) SkillReadyForActivation(ctx context.Context, skillID string) (bool, error) {
	var ready bool
	err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM skills s JOIN lessons l ON l.skill_id=s.id WHERE s.id=$1 AND s.grade_level<=4)`, skillID).Scan(&ready)
	if err != nil {
		return false, fmt.Errorf("check skill readiness: %w", err)
	}
	return ready, nil
}

// QuestionUsedInAssessment reports whether a question id already appears in a
// stored attempt, so review can warn before touching anything a student saw.
func (db *DB) QuestionUsedInAssessment(ctx context.Context, questionID string) (bool, error) {
	var used bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM assessment_items WHERE question_id=$1
			UNION ALL
			SELECT 1 FROM learning_items WHERE question_id=$1
		)`, questionID).Scan(&used)
	if err != nil {
		return false, fmt.Errorf("check question usage: %w", err)
	}
	return used, nil
}
