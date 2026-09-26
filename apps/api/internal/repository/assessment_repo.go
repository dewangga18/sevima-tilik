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
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	row := tx.QueryRowContext(ctx, `
		SELECT id, student_id, grade_level, status, started_at, completed_at, rule_version, target_skill_id, revision, stop_reason, learning_path
		FROM assessments
		WHERE id = $1
	`, id)

	var a domain.Assessment
	var statusStr string
	var completedAt sql.NullTime
	var pathJSON []byte
	if err := row.Scan(&a.ID, &a.StudentID, &a.GradeLevel, &statusStr, &a.StartedAt, &completedAt, &a.RuleVersion, &a.TargetSkillID, &a.Revision, &a.StopReason, &pathJSON); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("scan assessment: %w", err)
	}
	if err := json.Unmarshal(pathJSON, &a.LearningPath); err != nil {
		return nil, fmt.Errorf("decode learning path: %w", err)
	}
	if a.RuleVersion == "progressive-demo-v1" {
		a.MaxQuestions = 18
	}
	a.Status = domain.AssessmentStatus(statusStr)
	if completedAt.Valid {
		a.CompletedAt = &completedAt.Time
	}

	// Fetch items
	rows, err := tx.QueryContext(ctx, `
		SELECT ai.id, ai.assessment_id, ai.question_id, ai.order_index, ai.student_answer, ai.is_correct, ai.answered_at,
		       q.skill_id, q.difficulty, q.prompt, q.options, q.explanation, ai.probe_for_skill_id
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
			&q.SkillID, &q.Difficulty, &q.Prompt, &optsJSON, &q.Explanation, &item.ProbeForSkillID,
		)
		if err != nil {
			return nil, err
		}

		q.ID = item.QuestionID
		if err := json.Unmarshal(optsJSON, &q.Options); err != nil {
			return nil, fmt.Errorf("decode question options: %w", err)
		}
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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read assessment items: %w", err)
	}
	// Fetch results/evidence
	eRows, err := tx.QueryContext(ctx, `
		SELECT se.skill_id, s.name, se.status, se.total_answered, se.total_correct, se.evidence_count, se.confidence, se.is_root_gap, se.related_target_skill_id
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
		err := eRows.Scan(&res.SkillID, &res.SkillName, &statusStr, &res.TotalAnswered, &res.TotalCorrect, &res.EvidenceCount, &res.Confidence, &res.IsRootGap, &res.RelatedTargetSkillID)
		if err != nil {
			return nil, err
		}
		res.Status = domain.SkillStatus(statusStr)
		a.Results = append(a.Results, res)
	}

	if err := eRows.Err(); err != nil {
		return nil, fmt.Errorf("read skill evidence: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
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

func (r *AssessmentRepository) GetHistoryByStudent(ctx context.Context, studentID string) (*domain.AssessmentHistory, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT a.id, a.grade_level, a.status, a.started_at, a.completed_at,
		       (SELECT COUNT(*) FROM assessment_items WHERE assessment_id = a.id),
		       (SELECT COUNT(*) FROM assessment_items WHERE assessment_id = a.id AND student_answer <> ''),
		       CASE WHEN a.status = 'completed' THEN (SELECT COUNT(*) FROM assessment_items WHERE assessment_id = a.id AND is_correct = TRUE) END,
		       (SELECT COUNT(*) FROM skill_evidence WHERE assessment_id = a.id AND evidence_count > 0),
		       COUNT(*) FILTER (WHERE a.status = $2) OVER ()
		FROM assessments a
		WHERE a.student_id = $1
		ORDER BY a.started_at DESC, a.id DESC
		LIMIT 5
	`, studentID, string(domain.AssessmentCompleted))
	if err != nil {
		return nil, fmt.Errorf("query assessment history: %w", err)
	}
	defer rows.Close()
	history := &domain.AssessmentHistory{Items: []domain.AssessmentSummary{}}
	for rows.Next() {
		var item domain.AssessmentSummary
		var completedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.GradeLevel, &item.Status, &item.StartedAt, &completedAt,
			&item.QuestionCount, &item.AnsweredCount, &item.CorrectCount, &item.AssessedSkillCount, &history.CompletedCount); err != nil {
			return nil, fmt.Errorf("scan assessment history: %w", err)
		}
		if completedAt.Valid {
			item.CompletedAt = &completedAt.Time
		}
		history.Items = append(history.Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read assessment history: %w", err)
	}
	return history, nil
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

// CleanupStaleAssessments marks in-progress assessments older than specified hours as abandoned
func (r *AssessmentRepository) CleanupStaleAssessments(ctx context.Context, hoursOld int) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE assessments 
		SET status = $1, 
		    completed_at = NOW(),
		    stop_reason = $2
		WHERE status = $3 
		  AND started_at < NOW() - INTERVAL '1 hour' * $4
	`, string(domain.AssessmentCompleted), "abandoned_timeout", string(domain.AssessmentInProgress), hoursOld)
	
	if err != nil {
		return 0, fmt.Errorf("cleanup stale assessments: %w", err)
	}
	
	return result.RowsAffected()
}
