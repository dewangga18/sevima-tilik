package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sevima/tilik-api/internal/domain"
)

var ErrNotAssigned = errors.New("classroom not assigned to teacher")

func (db *DB) GetTeacherClasses(ctx context.Context, teacherID string) ([]domain.TeacherClass, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT c.id, c.name, c.grade_level,
			(SELECT COUNT(*) FROM enrollments e WHERE e.classroom_id = c.id),
			(SELECT COUNT(*) FROM enrollments e JOIN assessments a ON a.student_id = e.student_id AND a.status = 'completed'
				WHERE e.classroom_id = c.id)
		FROM teacher_assignments ta
		JOIN classrooms c ON c.id = ta.classroom_id
		WHERE ta.teacher_id = $1
		ORDER BY c.name`, teacherID)
	if err != nil {
		return nil, fmt.Errorf("query teacher classes: %w", err)
	}
	defer rows.Close()

	classes := []domain.TeacherClass{}
	for rows.Next() {
		var c domain.TeacherClass
		if err := rows.Scan(&c.ID, &c.Name, &c.GradeLevel, &c.StudentCount, &c.AssessedCount); err != nil {
			return nil, err
		}
		classes = append(classes, c)
	}
	return classes, rows.Err()
}

func (db *DB) IsClassAssigned(ctx context.Context, teacherID, classroomID string) (bool, error) {
	var assigned bool
	err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM teacher_assignments WHERE teacher_id=$1 AND classroom_id=$2)`, teacherID, classroomID).Scan(&assigned)
	return assigned, err
}

func (db *DB) GetClassStudents(ctx context.Context, classroomID string) ([]domain.StudentOverview, error) {
	rows, err := db.QueryContext(ctx, `
		WITH latest AS (
			SELECT DISTINCT ON (student_id) id, student_id, status, completed_at
			FROM assessments WHERE status = 'completed'
			ORDER BY student_id, completed_at DESC NULLS LAST
		)
		SELECT u.id, u.name,
			COALESCE(l.status, CASE WHEN inprog.student_id IS NOT NULL THEN 'in_progress' ELSE 'unassessed' END),
			COALESCE((SELECT COUNT(*) FROM skill_evidence se WHERE se.student_id = u.id AND se.assessment_id = l.id), 0),
			COALESCE((SELECT COUNT(*) FROM skill_evidence se WHERE se.student_id = u.id AND se.assessment_id = l.id AND se.status IN ('needs_practice','inconclusive')), 0),
			COALESCE((SELECT COUNT(*) FROM skill_evidence se WHERE se.student_id = u.id AND se.assessment_id = l.id AND se.is_root_gap), 0),
			l.completed_at
		FROM enrollments e
		JOIN users u ON u.id = e.student_id AND u.role = 'student'
		LEFT JOIN latest l ON l.student_id = u.id
		LEFT JOIN (SELECT DISTINCT student_id FROM assessments WHERE status = 'in_progress') inprog ON inprog.student_id = u.id
		WHERE e.classroom_id = $1
		ORDER BY u.name`, classroomID)
	if err != nil {
		return nil, fmt.Errorf("query class students: %w", err)
	}
	defer rows.Close()

	students := []domain.StudentOverview{}
	for rows.Next() {
		var s domain.StudentOverview
		var completedAt sql.NullTime
		if err := rows.Scan(&s.ID, &s.Name, &s.LatestStatus, &s.AssessedSkills, &s.VisibleGapCount, &s.RootGapCount, &completedAt); err != nil {
			return nil, err
		}
		if completedAt.Valid {
			t := completedAt.Time
			s.UpdatedAt = &t
		}
		students = append(students, s)
	}
	return students, rows.Err()
}

func (db *DB) GetStudentInsight(ctx context.Context, studentID string) (*domain.StudentInsight, error) {
	insight := &domain.StudentInsight{Results: []domain.SkillResult{}, Recommendations: []domain.LearningPathItem{}, Progress: []domain.ProgressEntry{}}

	var classroomID, classroomName string
	err := db.QueryRowContext(ctx, `
		SELECT u.name, c.id, c.name FROM users u
		JOIN enrollments e ON e.student_id = u.id
		JOIN classrooms c ON c.id = e.classroom_id
		WHERE u.id = $1 AND u.role = 'student'`, studentID).Scan(&insight.StudentName, &classroomID, &classroomName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("query student classroom: %w", err)
	}
	insight.StudentID = studentID
	insight.ClassroomID = classroomID
	insight.ClassroomName = classroomName

	var assessmentID, targetSkill, learningPath sql.NullString
	var completedAt sql.NullTime
	err = db.QueryRowContext(ctx, `
		SELECT id, completed_at, target_skill_id, learning_path FROM assessments
		WHERE student_id = $1 AND status = 'completed'
		ORDER BY completed_at DESC NULLS LAST LIMIT 1`, studentID).Scan(&assessmentID, &completedAt, &targetSkill, &learningPath)
	if errors.Is(err, sql.ErrNoRows) {
		return insight, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query latest assessment: %w", err)
	}
	if assessmentID.Valid {
		insight.AssessmentID = assessmentID.String
	}
	if targetSkill.Valid {
		insight.TargetSkillID = targetSkill.String
	}
	if completedAt.Valid {
		t := completedAt.Time
		insight.AssessedAt = &t
	}
	if learningPath.Valid && learningPath.String != "" {
		_ = json.Unmarshal([]byte(learningPath.String), &insight.Recommendations)
	}

	rows, err := db.QueryContext(ctx, `
		SELECT se.skill_id, s.name, se.status, se.total_answered, se.total_correct, se.evidence_count, se.confidence, se.is_root_gap
		FROM skill_evidence se JOIN skills s ON s.id = se.skill_id
		WHERE se.student_id = $1 AND se.assessment_id = $2
		ORDER BY se.is_root_gap DESC, s.name`, studentID, insight.AssessmentID)
	if err != nil {
		return nil, fmt.Errorf("query skill evidence: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var r domain.SkillResult
		if err := rows.Scan(&r.SkillID, &r.SkillName, &r.Status, &r.TotalAnswered, &r.TotalCorrect, &r.EvidenceCount, &r.Confidence, &r.IsRootGap); err != nil {
			return nil, err
		}
		insight.Results = append(insight.Results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	progressRows, err := db.QueryContext(ctx, `
		SELECT sp.skill_id, s.name, sp.score, sp.status, sp.evidence_count, sp.updated_at
		FROM skill_progress sp JOIN skills s ON s.id = sp.skill_id
		WHERE sp.student_id = $1 ORDER BY sp.updated_at DESC`, studentID)
	if err != nil {
		return nil, fmt.Errorf("query skill progress: %w", err)
	}
	defer progressRows.Close()
	for progressRows.Next() {
		var p domain.ProgressEntry
		if err := progressRows.Scan(&p.SkillID, &p.SkillName, &p.Score, &p.Status, &p.EvidenceCount, &p.UpdatedAt); err != nil {
			return nil, err
		}
		insight.Progress = append(insight.Progress, p)
	}
	return insight, progressRows.Err()
}

func (db *DB) IsStudentTaughtBy(ctx context.Context, teacherID, studentID string) (bool, error) {
	var taught bool
	err := db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM teacher_assignments ta
			JOIN enrollments e ON e.classroom_id = ta.classroom_id
			WHERE ta.teacher_id = $1 AND e.student_id = $2)`, teacherID, studentID).Scan(&taught)
	return taught, err
}
