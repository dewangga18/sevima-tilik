package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/sevima/tilik-api/internal/domain"
)

type AssessmentRepository struct {
	db *DB
}

var ErrAssessmentCompleted = errors.New("assessment already completed")

func NewAssessmentRepository(db *DB) *AssessmentRepository {
	return &AssessmentRepository{db: db}
}

func (r *AssessmentRepository) CreateAssessment(ctx context.Context, a *domain.Assessment) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO assessments (id, student_id, grade_level, status, started_at)
		VALUES ($1, $2, $3, $4, $5)
	`, a.ID, a.StudentID, a.GradeLevel, string(a.Status), a.StartedAt)
	if err != nil {
		return fmt.Errorf("insert assessment: %w", err)
	}

	for _, item := range a.Items {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO assessment_items (id, assessment_id, question_id, order_index)
			VALUES ($1, $2, $3, $4)
		`, item.ID, item.AssessmentID, item.QuestionID, item.OrderIndex)
		if err != nil {
			return fmt.Errorf("insert assessment item: %w", err)
		}
	}

	return tx.Commit()
}

func (r *AssessmentRepository) GetAssessmentByID(ctx context.Context, id string) (*domain.Assessment, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, student_id, grade_level, status, started_at, completed_at
		FROM assessments
		WHERE id = $1
	`, id)

	var a domain.Assessment
	var statusStr string
	var completedAt sql.NullTime
	if err := row.Scan(&a.ID, &a.StudentID, &a.GradeLevel, &statusStr, &a.StartedAt, &completedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan assessment: %w", err)
	}
	a.Status = domain.AssessmentStatus(statusStr)
	if completedAt.Valid {
		a.CompletedAt = &completedAt.Time
	}

	// Fetch items
	rows, err := r.db.QueryContext(ctx, `
		SELECT ai.id, ai.assessment_id, ai.question_id, ai.order_index, ai.student_answer, ai.is_correct, ai.answered_at,
		       q.skill_id, q.difficulty, q.prompt, q.options, q.explanation
		FROM assessment_items ai
		JOIN questions q ON q.id = ai.question_id
		WHERE ai.assessment_id = $1
		ORDER BY ai.order_index ASC
	`, a.ID)
	if err != nil {
		return nil, fmt.Errorf("query assessment items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.AssessmentItem
		var q domain.Question
		var optsJSON []byte
		var studentAns sql.NullString
		var isCorrect sql.NullBool
		var answeredAt sql.NullTime

		err := rows.Scan(
			&item.ID, &item.AssessmentID, &item.QuestionID, &item.OrderIndex,
			&studentAns, &isCorrect, &answeredAt,
			&q.SkillID, &q.Difficulty, &q.Prompt, &optsJSON, &q.Explanation,
		)
		if err != nil {
			return nil, err
		}

		q.ID = item.QuestionID
		_ = json.Unmarshal(optsJSON, &q.Options)
		item.Question = &q

		if studentAns.Valid {
			item.StudentAnswer = studentAns.String
		}
		if isCorrect.Valid {
			item.IsCorrect = &isCorrect.Bool
		}
		if answeredAt.Valid {
			item.AnsweredAt = &answeredAt.Time
		}
		a.Items = append(a.Items, item)
	}

	// Fetch results/evidence
	eRows, err := r.db.QueryContext(ctx, `
		SELECT se.skill_id, s.name, se.status, se.total_answered, se.total_correct, se.evidence_count, se.confidence, se.is_root_gap
		FROM skill_evidence se
		JOIN skills s ON s.id = se.skill_id
		WHERE se.assessment_id = $1
		ORDER BY se.skill_id ASC
	`, a.ID)
	if err != nil {
		return nil, fmt.Errorf("query evidence: %w", err)
	}
	defer eRows.Close()

	for eRows.Next() {
		var res domain.SkillResult
		var statusStr string
		err := eRows.Scan(&res.SkillID, &res.SkillName, &statusStr, &res.TotalAnswered, &res.TotalCorrect, &res.EvidenceCount, &res.Confidence, &res.IsRootGap)
		if err != nil {
			return nil, err
		}
		res.Status = domain.SkillStatus(statusStr)
		a.Results = append(a.Results, res)
	}

	return &a, nil
}

func (r *AssessmentRepository) GetLatestByStudent(ctx context.Context, studentID string) (*domain.Assessment, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id
		FROM assessments
		WHERE student_id = $1
		ORDER BY started_at DESC
		LIMIT 1
	`, studentID)

	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return r.GetAssessmentByID(ctx, id)
}

func (r *AssessmentRepository) SaveEvaluation(ctx context.Context, assessmentID string, items []domain.AssessmentItem, results []domain.SkillResult, studentID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	now := time.Now()

	// Claim completion before changing evidence. Concurrent submissions wait on
	// this row and recheck the status after the winning transaction commits.
	completion, err := tx.ExecContext(ctx, `
		UPDATE assessments
		SET status = $1, completed_at = $2
		WHERE id = $3 AND student_id = $4 AND status = $5
	`, string(domain.AssessmentCompleted), now, assessmentID, studentID, string(domain.AssessmentInProgress))
	if err != nil {
		return fmt.Errorf("complete assessment: %w", err)
	}
	count, err := completion.RowsAffected()
	if err != nil {
		return fmt.Errorf("check assessment completion: %w", err)
	}
	if count != 1 {
		return ErrAssessmentCompleted
	}

	// Update items
	for _, item := range items {
		_, err = tx.ExecContext(ctx, `
			UPDATE assessment_items
			SET student_answer = $1, is_correct = $2, answered_at = $3
			WHERE assessment_id = $4 AND question_id = $5
		`, item.StudentAnswer, *item.IsCorrect, now, assessmentID, item.QuestionID)
		if err != nil {
			return fmt.Errorf("update assessment item: %w", err)
		}
	}

	// Save skill evidence
	for _, res := range results {
		evidenceID := fmt.Sprintf("ev-%s-%s", assessmentID, res.SkillID)
		_, err = tx.ExecContext(ctx, `
			INSERT INTO skill_evidence (id, student_id, skill_id, assessment_id, status, total_answered, total_correct, evidence_count, confidence, is_root_gap, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
			ON CONFLICT (id) DO UPDATE SET
				status = EXCLUDED.status,
				total_answered = EXCLUDED.total_answered,
				total_correct = EXCLUDED.total_correct,
				evidence_count = EXCLUDED.evidence_count,
				confidence = EXCLUDED.confidence,
				is_root_gap = EXCLUDED.is_root_gap,
				updated_at = EXCLUDED.updated_at
		`, evidenceID, studentID, res.SkillID, assessmentID, string(res.Status), res.TotalAnswered, res.TotalCorrect, res.EvidenceCount, res.Confidence, res.IsRootGap, now)
		if err != nil {
			return fmt.Errorf("insert skill evidence: %w", err)
		}
	}

	return tx.Commit()
}
