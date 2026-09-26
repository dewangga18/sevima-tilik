package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sevima/tilik-api/internal/domain"
)

func (db *DB) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, email, name, role, grade_level, is_active, created_at
		FROM users ORDER BY role, name`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	users := []domain.User{}
	for rows.Next() {
		var u domain.User
		var roleStr string
		if err := rows.Scan(&u.ID, &u.Email, &u.Name, &roleStr, &u.GradeLevel, &u.IsActive, &u.CreatedAt); err != nil {
			return nil, err
		}
		u.Role = domain.Role(roleStr)
		users = append(users, u)
	}
	return users, rows.Err()
}

var ErrEmailTaken = errors.New("email already registered")

func (db *DB) CreateUser(ctx context.Context, user *domain.User) error {
	var exists bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, user.Email).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return ErrEmailTaken
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO users (id, email, password_hash, name, role, grade_level, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, TRUE)`,
		user.ID, user.Email, user.PasswordHash, user.Name, string(user.Role), user.GradeLevel)
	return err
}

func (db *DB) SetUserActive(ctx context.Context, userID string, active bool) error {
	result, err := db.ExecContext(ctx, `UPDATE users SET is_active = $1 WHERE id = $2`, active, userID)
	if err != nil {
		return fmt.Errorf("set user active: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	if !active {
		// Deactivation revokes existing sessions immediately.
		if _, err := db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID); err != nil {
			return fmt.Errorf("revoke sessions: %w", err)
		}
	}
	return nil
}

func (db *DB) ListClassrooms(ctx context.Context) ([]domain.AdminClass, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT c.id, c.name, c.grade_level,
			(SELECT COUNT(*) FROM enrollments e WHERE e.classroom_id = c.id),
			(SELECT COUNT(*) FROM teacher_assignments ta WHERE ta.classroom_id = c.id)
		FROM classrooms c ORDER BY c.name`)
	if err != nil {
		return nil, fmt.Errorf("list classrooms: %w", err)
	}
	defer rows.Close()
	classrooms := []domain.AdminClass{}
	for rows.Next() {
		var c domain.AdminClass
		if err := rows.Scan(&c.ID, &c.Name, &c.GradeLevel, &c.StudentCount, &c.TeacherCount); err != nil {
			return nil, err
		}
		classrooms = append(classrooms, c)
	}
	return classrooms, rows.Err()
}

func (db *DB) ClassroomExists(ctx context.Context, classroomID string) (bool, error) {
	var exists bool
	err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM classrooms WHERE id = $1)`, classroomID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check classroom: %w", err)
	}
	return exists, nil
}

// ClassRoster returns enrolled students and assigned teachers for one classroom.
// Ordering is stable so repeated admin reads do not reshuffle the placement UI.
func (db *DB) ClassRoster(ctx context.Context, classroomID string) (*domain.ClassRoster, error) {
	roster := &domain.ClassRoster{ClassroomID: classroomID, StudentIDs: []string{}, TeacherIDs: []string{}}

	studentRows, err := db.QueryContext(ctx, `SELECT student_id FROM enrollments WHERE classroom_id = $1 ORDER BY student_id`, classroomID)
	if err != nil {
		return nil, fmt.Errorf("roster students: %w", err)
	}
	defer studentRows.Close()
	for studentRows.Next() {
		var id string
		if err := studentRows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan roster student: %w", err)
		}
		roster.StudentIDs = append(roster.StudentIDs, id)
	}
	if err := studentRows.Err(); err != nil {
		return nil, fmt.Errorf("read roster students: %w", err)
	}

	teacherRows, err := db.QueryContext(ctx, `SELECT teacher_id FROM teacher_assignments WHERE classroom_id = $1 ORDER BY teacher_id`, classroomID)
	if err != nil {
		return nil, fmt.Errorf("roster teachers: %w", err)
	}
	defer teacherRows.Close()
	for teacherRows.Next() {
		var id string
		if err := teacherRows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan roster teacher: %w", err)
		}
		roster.TeacherIDs = append(roster.TeacherIDs, id)
	}
	if err := teacherRows.Err(); err != nil {
		return nil, fmt.Errorf("read roster teachers: %w", err)
	}
	return roster, nil
}

// UserHasRole reports whether the user exists with exactly the given role, so
// placement rules can be checked before writing enrollments or assignments.
func (db *DB) UserHasRole(ctx context.Context, userID string, role domain.Role) (bool, error) {
	var ok bool
	err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND role = $2)`, userID, string(role)).Scan(&ok)
	if err != nil {
		return false, fmt.Errorf("check user role: %w", err)
	}
	return ok, nil
}

// UpdateUserRole changes the stored role. Sessions re-read the user on every
// request, so the new role applies to the target's next request without
// needing to revoke their token. Grade level belongs to students only, so it is
// cleared when the account becomes a teacher or admin.
func (db *DB) UpdateUserRole(ctx context.Context, userID string, role domain.Role) error {
	result, err := db.ExecContext(ctx, `
		UPDATE users
		SET role = $1, grade_level = CASE WHEN $1 = 'student' THEN grade_level ELSE 0 END
		WHERE id = $2`, string(role), userID)
	if err != nil {
		return fmt.Errorf("update user role: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (db *DB) CreateClassroom(ctx context.Context, id, name string, gradeLevel int) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO classrooms (id, name, grade_level) VALUES ($1, $2, $3)
		ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, grade_level = EXCLUDED.grade_level`,
		id, name, gradeLevel)
	return err
}

func (db *DB) SetEnrollment(ctx context.Context, classroomID, studentID string, enrolled bool) error {
	if enrolled {
		isStudent, err := db.UserHasRole(ctx, studentID, domain.RoleStudent)
		if err != nil {
			return err
		}
		if !isStudent {
			return sql.ErrNoRows
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO enrollments (classroom_id, student_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, classroomID, studentID); err != nil {
			return fmt.Errorf("enroll student: %w", err)
		}
		return nil
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM enrollments WHERE classroom_id = $1 AND student_id = $2`, classroomID, studentID); err != nil {
		return fmt.Errorf("unenroll student: %w", err)
	}
	return nil
}

func (db *DB) SetTeacherAssignment(ctx context.Context, teacherID, classroomID string, assigned bool) error {
	if assigned {
		isTeacher, err := db.UserHasRole(ctx, teacherID, domain.RoleTeacher)
		if err != nil {
			return err
		}
		if !isTeacher {
			return sql.ErrNoRows
		}
		if _, err := db.ExecContext(ctx, `INSERT INTO teacher_assignments (teacher_id, classroom_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, teacherID, classroomID); err != nil {
			return fmt.Errorf("assign teacher: %w", err)
		}
		return nil
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM teacher_assignments WHERE teacher_id = $1 AND classroom_id = $2`, teacherID, classroomID); err != nil {
		return fmt.Errorf("unassign teacher: %w", err)
	}
	return nil
}

func (db *DB) QuestionBankSummary(ctx context.Context) (*domain.QuestionBankSummary, error) {
	summary := &domain.QuestionBankSummary{
		Active:    []domain.QuestionBankSkillRow{},
		Candidate: []domain.QuestionBankStatusRow{},
	}
	rows, err := db.QueryContext(ctx, `
		SELECT s.id, s.name, COUNT(q.id)
		FROM skills s LEFT JOIN questions q ON q.skill_id = s.id
		GROUP BY s.id, s.name ORDER BY s.name`)
	if err != nil {
		return nil, fmt.Errorf("bank active summary: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var r domain.QuestionBankSkillRow
		if err := rows.Scan(&r.SkillID, &r.SkillName, &r.QuestionCount); err != nil {
			return nil, err
		}
		summary.Active = append(summary.Active, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	statusRows, err := db.QueryContext(ctx, `
		SELECT status, COUNT(*) FROM question_candidates GROUP BY status ORDER BY status`)
	if err != nil {
		return nil, fmt.Errorf("bank candidate summary: %w", err)
	}
	defer statusRows.Close()
	for statusRows.Next() {
		var r domain.QuestionBankStatusRow
		if err := statusRows.Scan(&r.Status, &r.Count); err != nil {
			return nil, err
		}
		summary.Candidate = append(summary.Candidate, r)
	}
	return summary, statusRows.Err()
}
