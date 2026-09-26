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

var ErrAssessmentChanged = errors.New("assessment revision changed")

// Serialize starts for one student, including starts that race before an attempt exists.
func (r *AssessmentRepository) CreateProgressiveAssessment(ctx context.Context, a *domain.Assessment) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, a.StudentID); err != nil {
		return "", err
	}
	var active string
	err = tx.QueryRowContext(ctx, `SELECT id FROM assessments WHERE student_id=$1 AND status='in_progress' ORDER BY started_at DESC,id DESC LIMIT 1`, a.StudentID).Scan(&active)
	if err == nil {
		return active, tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO assessments(id,student_id,grade_level,status,started_at,rule_version,target_skill_id) VALUES($1,$2,$3,$4,$5,$6,$7)`, a.ID, a.StudentID, a.GradeLevel, string(a.Status), a.StartedAt, a.RuleVersion, a.TargetSkillID)
	if err != nil {
		return "", err
	}
	item := a.Items[0]
	if _, err = tx.ExecContext(ctx, `INSERT INTO assessment_items(id,assessment_id,question_id,order_index,probe_for_skill_id) VALUES($1,$2,$3,$4,$5)`, item.ID, a.ID, item.QuestionID, item.OrderIndex, item.ProbeForSkillID); err != nil {
		return "", err
	}
	return a.ID, tx.Commit()
}

// The revision claim, answer, next question, evidence, and completion commit together.
func (r *AssessmentRepository) SaveProgressiveStep(ctx context.Context, a *domain.Assessment, answered domain.AssessmentItem, next *domain.AssessmentItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var revision int
	var status string
	if err = tx.QueryRowContext(ctx, `SELECT revision,status FROM assessments WHERE id=$1 AND student_id=$2 FOR UPDATE`, a.ID, a.StudentID).Scan(&revision, &status); err != nil {
		return err
	}
	if revision != a.Revision || status != string(domain.AssessmentInProgress) {
		return ErrAssessmentChanged
	}
	saved, err := tx.ExecContext(ctx, `UPDATE assessment_items SET student_answer=$1,is_correct=$2,answered_at=$3 WHERE id=$4 AND assessment_id=$5 AND answered_at IS NULL`, answered.StudentAnswer, *answered.IsCorrect, answered.AnsweredAt, answered.ID, a.ID)
	if err != nil {
		return err
	}
	count, err := saved.RowsAffected()
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrAssessmentChanged
	}
	if next != nil {
		if _, err = tx.ExecContext(ctx, `INSERT INTO assessment_items(id,assessment_id,question_id,order_index,probe_for_skill_id) VALUES($1,$2,$3,$4,$5)`, next.ID, a.ID, next.QuestionID, next.OrderIndex, next.ProbeForSkillID); err != nil {
			return err
		}
	}
	path, err := json.Marshal(a.LearningPath)
	if err != nil {
		return err
	}
	if a.LearningPath == nil {
		path = []byte("[]")
	}
	var completedAt *time.Time
	if a.Status == domain.AssessmentCompleted {
		now := time.Now()
		completedAt = &now
	}
	if _, err = tx.ExecContext(ctx, `UPDATE assessments SET revision=revision+1,status=$1,completed_at=$2,stop_reason=$3,learning_path=$4 WHERE id=$5`, string(a.Status), completedAt, a.StopReason, path, a.ID); err != nil {
		return err
	}
	if a.Status == domain.AssessmentCompleted {
		for _, result := range a.Results {
			_, err = tx.ExecContext(ctx, `INSERT INTO skill_evidence(id,student_id,skill_id,assessment_id,status,total_answered,total_correct,evidence_count,confidence,is_root_gap,related_target_skill_id) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, fmt.Sprintf("ev-%s-%s", a.ID, result.SkillID), a.StudentID, result.SkillID, a.ID, string(result.Status), result.TotalAnswered, result.TotalCorrect, result.EvidenceCount, result.Confidence, result.IsRootGap, result.RelatedTargetSkillID)
			if err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}
