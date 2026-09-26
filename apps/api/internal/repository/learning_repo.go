package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/sevima/tilik-api/internal/domain"
)

var ErrLearningChanged = errors.New("learning state changed")
var ErrLearningBankEmpty = errors.New("learning question bank exhausted")

type LearningRepository struct{ db *DB }

func NewLearningRepository(db *DB) *LearningRepository { return &LearningRepository{db: db} }

func (r *LearningRepository) GetLesson(ctx context.Context, skillID string) (*domain.Lesson, error) {
	var lesson domain.Lesson
	var steps []byte
	err := r.db.QueryRowContext(ctx, `SELECT id,skill_id,title,estimated_minutes,steps,hint FROM lessons WHERE skill_id=$1`, skillID).Scan(&lesson.ID, &lesson.SkillID, &lesson.Title, &lesson.EstimatedMinutes, &steps, &lesson.Hint)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(steps, &lesson.Steps); err != nil {
		return nil, err
	}
	return &lesson, nil
}

func (r *LearningRepository) GetSession(ctx context.Context, id string) (*domain.LearningSession, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var s domain.LearningSession
	err = tx.QueryRowContext(ctx, `SELECT l.id,l.student_id,l.source_assessment_id,l.skill_id,k.name,l.request_id,l.rule_version,l.stage,l.revision,l.started_at,l.lesson_completed_at,l.completed_at,l.before_score,l.score,l.outcome,l.stop_reason,l.review_skill_id FROM learning_sessions l JOIN skills k ON k.id=l.skill_id WHERE l.id=$1`, id).Scan(&s.ID, &s.StudentID, &s.SourceAssessmentID, &s.SkillID, &s.SkillName, &s.RequestID, &s.RuleVersion, &s.Stage, &s.Revision, &s.StartedAt, &s.LessonCompletedAt, &s.CompletedAt, &s.BeforeScore, &s.Score, &s.Outcome, &s.StopReason, &s.ReviewSkillID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.Items = make([]domain.LearningItem, 0)
	rows, err := tx.QueryContext(ctx, `SELECT i.id,i.question_id,i.stage,i.order_index,i.student_answer,i.is_correct,i.answered_at,q.skill_id,q.difficulty,q.prompt,q.options,q.explanation FROM learning_items i JOIN questions q ON q.id=i.question_id WHERE i.session_id=$1 ORDER BY i.order_index`, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var item domain.LearningItem
		var q domain.Question
		var options []byte
		var answer sql.NullString
		if err = rows.Scan(&item.ID, &item.QuestionID, &item.Stage, &item.OrderIndex, &answer, &item.IsCorrect, &item.AnsweredAt, &q.SkillID, &q.Difficulty, &q.Prompt, &options, &q.Explanation); err != nil {
			rows.Close()
			return nil, err
		}
		if err = json.Unmarshal(options, &q.Options); err != nil {
			rows.Close()
			return nil, err
		}
		item.StudentAnswer = answer.String
		q.ID = item.QuestionID
		item.Question = &q
		s.Items = append(s.Items, item)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	s.Lesson, err = r.GetLesson(ctx, s.SkillID)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *LearningRepository) AvailableQuestions(ctx context.Context, studentID, skillID, purpose string) ([]domain.Question, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT q.id,q.skill_id,q.difficulty,q.prompt,q.options,q.answer_key,q.explanation FROM questions q WHERE q.skill_id=$1 AND q.purpose=$2 AND NOT EXISTS(SELECT 1 FROM learning_items i JOIN learning_sessions s ON s.id=i.session_id WHERE i.question_id=q.id AND s.student_id=$3) ORDER BY q.difficulty,q.id`, skillID, purpose, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bank := make([]domain.Question, 0)
	for rows.Next() {
		var q domain.Question
		var options []byte
		if err = rows.Scan(&q.ID, &q.SkillID, &q.Difficulty, &q.Prompt, &options, &q.AnswerKey, &q.Explanation); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(options, &q.Options); err != nil {
			return nil, err
		}
		bank = append(bank, q)
	}
	return bank, rows.Err()
}

func (r *LearningRepository) StartSession(ctx context.Context, s *domain.LearningSession) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, s.StudentID+":learning"); err != nil {
		return "", err
	}
	var id, skill, source string
	err = tx.QueryRowContext(ctx, `SELECT session_id,skill_id,source_assessment_id FROM learning_start_requests WHERE student_id=$1 AND request_id=$2`, s.StudentID, s.RequestID).Scan(&id, &skill, &source)
	if err == nil {
		if skill != s.SkillID || source != s.SourceAssessmentID {
			return "", ErrLearningChanged
		}
		return id, tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	err = tx.QueryRowContext(ctx, `SELECT id FROM learning_sessions WHERE student_id=$1 AND skill_id=$2 AND stage IN('lesson','practice','reassessment') ORDER BY started_at DESC LIMIT 1`, s.StudentID, s.SkillID).Scan(&id)
	if err == nil {
		if _, err = tx.ExecContext(ctx, `INSERT INTO learning_start_requests(student_id,request_id,source_assessment_id,skill_id,session_id) VALUES($1,$2,$3,$4,$5)`, s.StudentID, s.RequestID, s.SourceAssessmentID, s.SkillID, id); err != nil {
			return "", err
		}
		return id, tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	// Check quotas under the same start lock; offered questions are never reused.
	for _, purpose := range []string{"practice", "reassessment"} {
		var count int
		if err = tx.QueryRowContext(ctx, `SELECT count(*) FROM questions q WHERE q.skill_id=$1 AND q.purpose=$2 AND NOT EXISTS(SELECT 1 FROM learning_items i JOIN learning_sessions l ON l.id=i.session_id WHERE i.question_id=q.id AND l.student_id=$3)`, s.SkillID, purpose, s.StudentID).Scan(&count); err != nil {
			return "", err
		}
		if count < 3 {
			return "", ErrLearningBankEmpty
		}
	}
	var latestScore int
	err = tx.QueryRowContext(ctx, `SELECT score FROM skill_progress WHERE student_id=$1 AND skill_id=$2 AND updated_at > COALESCE((SELECT MAX(completed_at) FROM assessments WHERE student_id=$1 AND rule_version='progressive-demo-v1'),'-infinity'::timestamptz)`, s.StudentID, s.SkillID).Scan(&latestScore)
	if err == nil {
		s.BeforeScore = &latestScore
	} else if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO learning_sessions(id,student_id,source_assessment_id,skill_id,request_id,rule_version,stage,started_at,before_score) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`, s.ID, s.StudentID, s.SourceAssessmentID, s.SkillID, s.RequestID, s.RuleVersion, s.Stage, s.StartedAt, s.BeforeScore)
	if err != nil {
		return "", err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO learning_start_requests(student_id,request_id,source_assessment_id,skill_id,session_id) VALUES($1,$2,$3,$4,$5)`, s.StudentID, s.RequestID, s.SourceAssessmentID, s.SkillID, s.ID); err != nil {
		return "", err
	}
	return s.ID, tx.Commit()
}

func (r *LearningRepository) SaveStep(ctx context.Context, s *domain.LearningSession, answered, next *domain.LearningItem) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var revision int
	var stage string
	if err = tx.QueryRowContext(ctx, `SELECT revision,stage FROM learning_sessions WHERE id=$1 AND student_id=$2 FOR UPDATE`, s.ID, s.StudentID).Scan(&revision, &stage); err != nil {
		return err
	}
	if revision != s.Revision || stage == "completed" || stage == "exhausted" {
		return ErrLearningChanged
	}
	if answered != nil {
		saved, err := tx.ExecContext(ctx, `UPDATE learning_items SET student_answer=$1,is_correct=$2,answered_at=$3 WHERE id=$4 AND session_id=$5 AND answered_at IS NULL`, answered.StudentAnswer, *answered.IsCorrect, answered.AnsweredAt, answered.ID, s.ID)
		if err != nil {
			return err
		}
		count, err := saved.RowsAffected()
		if err != nil {
			return err
		}
		if count != 1 {
			return ErrLearningChanged
		}
	}
	if next != nil {
		if _, err = tx.ExecContext(ctx, `INSERT INTO learning_items(id,session_id,question_id,stage,order_index) VALUES($1,$2,$3,$4,$5)`, next.ID, s.ID, next.QuestionID, next.Stage, next.OrderIndex); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE learning_sessions SET stage=$1,revision=revision+1,lesson_completed_at=$2,completed_at=$3,score=$4,outcome=$5,stop_reason=$6,review_skill_id=$7 WHERE id=$8`, s.Stage, s.LessonCompletedAt, s.CompletedAt, s.Score, string(s.Outcome), s.StopReason, s.ReviewSkillID, s.ID)
	if err != nil {
		return err
	}
	if s.Stage == "completed" {
		correct := 0
		for _, item := range s.Items {
			if item.Stage == "reassessment" && item.IsCorrect != nil && *item.IsCorrect {
				correct++
			}
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO skill_progress(student_id,skill_id,source_session_id,score,status,evidence_count,correct_count,updated_at) VALUES($1,$2,$3,$4,$5,3,$6,$7) ON CONFLICT(student_id,skill_id) DO UPDATE SET source_session_id=EXCLUDED.source_session_id,score=EXCLUDED.score,status=EXCLUDED.status,evidence_count=EXCLUDED.evidence_count,correct_count=EXCLUDED.correct_count,updated_at=EXCLUDED.updated_at`, s.StudentID, s.SkillID, s.ID, *s.Score, string(s.Outcome), correct, s.CompletedAt)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *LearningRepository) LatestDiagnosticID(ctx context.Context, studentID string) (string, error) {
	var id string
	err := r.db.QueryRowContext(ctx, `SELECT id FROM assessments WHERE student_id=$1 AND status='completed' AND rule_version='progressive-demo-v1' ORDER BY completed_at DESC,id DESC LIMIT 1`, studentID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return id, err
}

func (r *LearningRepository) StoredProgress(ctx context.Context, studentID string) ([]domain.SkillProgress, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT p.skill_id,k.name,p.score,p.status,p.evidence_count,p.correct_count,p.updated_at FROM skill_progress p JOIN skills k ON k.id=p.skill_id WHERE p.student_id=$1 ORDER BY p.skill_id`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]domain.SkillProgress, 0)
	for rows.Next() {
		var p domain.SkillProgress
		if err = rows.Scan(&p.SkillID, &p.SkillName, &p.Score, &p.Status, &p.EvidenceCount, &p.CorrectCount, &p.UpdatedAt); err != nil {
			return nil, err
		}
		p.Source = "reassessment"
		result = append(result, p)
	}
	return result, rows.Err()
}

func (r *LearningRepository) SessionSummaries(ctx context.Context, studentID string) ([]domain.LearningSummary, []domain.LearningSummary, int, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT l.id,l.skill_id,k.name,l.stage,l.started_at,l.completed_at,l.score FROM learning_sessions l JOIN skills k ON k.id=l.skill_id WHERE l.student_id=$1 ORDER BY l.started_at DESC,l.id DESC`, studentID)
	if err != nil {
		return nil, nil, 0, err
	}
	defer rows.Close()
	active := make([]domain.LearningSummary, 0)
	recent := make([]domain.LearningSummary, 0)
	completed := 0
	for rows.Next() {
		var summary domain.LearningSummary
		if err = rows.Scan(&summary.ID, &summary.SkillID, &summary.SkillName, &summary.Stage, &summary.StartedAt, &summary.CompletedAt, &summary.Score); err != nil {
			return nil, nil, 0, err
		}
		if len(recent) < 5 {
			recent = append(recent, summary)
		}
		if summary.Stage == "completed" {
			completed++
		}
		if summary.Stage == "lesson" || summary.Stage == "practice" || summary.Stage == "reassessment" {
			active = append(active, summary)
		}
	}
	return active, recent, completed, rows.Err()
}
