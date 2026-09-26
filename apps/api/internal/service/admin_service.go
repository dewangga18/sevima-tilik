package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrAdminInputInvalid  = errors.New("invalid admin input")
	ErrAdminUserNotFound  = errors.New("user not found")
	ErrAdminClassNotFound = errors.New("classroom not found")
	ErrAdminEmailTaken    = repository.ErrEmailTaken
	ErrAdminSelfChange    = errors.New("admin cannot change own role or active state")
)

const (
	maxEmailLength    = 254
	maxNameLength     = 100
	minPasswordLength = 8
	// bcrypt rejects input beyond 72 bytes, so the limit is enforced before hashing.
	maxPasswordLength = 72
	// Cost 12 matches the demo user seed and keeps login latency acceptable at this scale.
	adminPasswordCost = 12
)

// adminGradeRange is the only grade with seeded content and working diagnostic,
// lesson, practice, and reassessment flows. Creating other grades would produce
// accounts that cannot complete a learning loop, so placement is restricted until
// the grade 5-9 slices ship.
const (
	adminMinGrade = 4
	adminMaxGrade = 4
)

type AdminService struct {
	db    *repository.DB
	users *repository.UserRepository
}

func NewAdminService(db *repository.DB, users *repository.UserRepository) *AdminService {
	return &AdminService{db: db, users: users}
}

type NewUserInput struct {
	Email      string
	Name       string
	Role       domain.Role
	GradeLevel int
	Password   string
}

type UpdateUserInput struct {
	Role     *domain.Role
	IsActive *bool
}

type NewClassInput struct {
	Name       string
	GradeLevel int
}

func (s *AdminService) ListUsers(ctx context.Context) ([]domain.User, error) {
	users, err := s.db.ListUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (s *AdminService) CreateUser(ctx context.Context, in NewUserInput) (*domain.User, error) {
	email, err := validateEmail(in.Email)
	if err != nil {
		return nil, err
	}
	name, err := validateName(in.Name)
	if err != nil {
		return nil, err
	}
	if err := validateRole(in.Role); err != nil {
		return nil, err
	}
	if err := validatePassword(in.Password); err != nil {
		return nil, err
	}

	grade := 0
	if in.Role == domain.RoleStudent {
		if err := validateGrade(in.GradeLevel); err != nil {
			return nil, err
		}
		grade = in.GradeLevel
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), adminPasswordCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	id, err := newAdminID("usr-")
	if err != nil {
		return nil, err
	}

	user := &domain.User{ID: id, Email: email, PasswordHash: string(hash), Name: name, Role: in.Role, GradeLevel: grade, IsActive: true}
	if err := s.db.CreateUser(ctx, user); err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			return nil, ErrAdminEmailTaken
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	// The insert relies on database defaults for created_at, so the stored row is
	// re-read to keep the response consistent with what was persisted.
	created, err := s.users.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("reload created user: %w", err)
	}
	if created == nil {
		return nil, fmt.Errorf("created user %s not found after insert", id)
	}
	return created, nil
}

// UpdateUser applies role and active-state changes. Guarding the actor's own
// record keeps a single admin from removing their own access and locking the
// school out of management.
func (s *AdminService) UpdateUser(ctx context.Context, actorID, userID string, in UpdateUserInput) (*domain.User, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrAdminInputInvalid
	}
	if in.Role == nil && in.IsActive == nil {
		return nil, ErrAdminInputInvalid
	}
	if userID == actorID {
		return nil, ErrAdminSelfChange
	}

	target, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load user: %w", err)
	}
	if target == nil {
		return nil, ErrAdminUserNotFound
	}

	if in.Role != nil {
		if err := validateRole(*in.Role); err != nil {
			return nil, err
		}
		if err := s.db.UpdateUserRole(ctx, userID, *in.Role); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrAdminUserNotFound
			}
			return nil, fmt.Errorf("update role: %w", err)
		}
	}

	if in.IsActive != nil {
		if err := s.db.SetUserActive(ctx, userID, *in.IsActive); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, ErrAdminUserNotFound
			}
			return nil, fmt.Errorf("set user active: %w", err)
		}
	}

	updated, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("reload user: %w", err)
	}
	if updated == nil {
		return nil, ErrAdminUserNotFound
	}
	return updated, nil
}

func (s *AdminService) ListClasses(ctx context.Context) ([]domain.AdminClass, error) {
	classes, err := s.db.ListClassrooms(ctx)
	if err != nil {
		return nil, fmt.Errorf("list classrooms: %w", err)
	}
	return classes, nil
}

func (s *AdminService) CreateClass(ctx context.Context, in NewClassInput) (*domain.AdminClass, error) {
	name, err := validateName(in.Name)
	if err != nil {
		return nil, err
	}
	if err := validateGrade(in.GradeLevel); err != nil {
		return nil, err
	}
	id, err := newAdminID("cls-")
	if err != nil {
		return nil, err
	}
	if err := s.db.CreateClassroom(ctx, id, name, in.GradeLevel); err != nil {
		return nil, fmt.Errorf("create classroom: %w", err)
	}
	return &domain.AdminClass{ID: id, Name: name, GradeLevel: in.GradeLevel}, nil
}

func (s *AdminService) Roster(ctx context.Context, classroomID string) (*domain.ClassRoster, error) {
	if strings.TrimSpace(classroomID) == "" {
		return nil, ErrAdminInputInvalid
	}
	exists, err := s.db.ClassroomExists(ctx, classroomID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrAdminClassNotFound
	}
	roster, err := s.db.ClassRoster(ctx, classroomID)
	if err != nil {
		return nil, fmt.Errorf("class roster: %w", err)
	}
	return roster, nil
}

func (s *AdminService) SetEnrollment(ctx context.Context, classroomID, studentID string, enrolled bool) error {
	if strings.TrimSpace(classroomID) == "" || strings.TrimSpace(studentID) == "" {
		return ErrAdminInputInvalid
	}
	exists, err := s.db.ClassroomExists(ctx, classroomID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrAdminClassNotFound
	}
	if err := s.db.SetEnrollment(ctx, classroomID, studentID, enrolled); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAdminUserNotFound
		}
		return fmt.Errorf("set enrollment: %w", err)
	}
	return nil
}

func (s *AdminService) SetAssignment(ctx context.Context, classroomID, teacherID string, assigned bool) error {
	if strings.TrimSpace(classroomID) == "" || strings.TrimSpace(teacherID) == "" {
		return ErrAdminInputInvalid
	}
	exists, err := s.db.ClassroomExists(ctx, classroomID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrAdminClassNotFound
	}
	if err := s.db.SetTeacherAssignment(ctx, teacherID, classroomID, assigned); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAdminUserNotFound
		}
		return fmt.Errorf("set assignment: %w", err)
	}
	return nil
}

func validateEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if normalized == "" || len(normalized) > maxEmailLength {
		return "", ErrAdminInputInvalid
	}
	if !strings.Contains(normalized, "@") || !strings.Contains(normalized, ".") {
		return "", ErrAdminInputInvalid
	}
	if strings.ContainsAny(normalized, " \t\r\n") {
		return "", ErrAdminInputInvalid
	}
	return normalized, nil
}

func validateName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || len(trimmed) > maxNameLength {
		return "", ErrAdminInputInvalid
	}
	return trimmed, nil
}

func validateRole(role domain.Role) error {
	switch role {
	case domain.RoleStudent, domain.RoleTeacher, domain.RoleAdmin:
		return nil
	default:
		return ErrAdminInputInvalid
	}
}

func validatePassword(password string) error {
	if len(password) < minPasswordLength || len(password) > maxPasswordLength {
		return ErrAdminInputInvalid
	}
	return nil
}

// validateGrade keeps student accounts and classrooms inside the grade range
// that actually has content and a working learning loop.
func validateGrade(grade int) error {
	if grade < adminMinGrade || grade > adminMaxGrade {
		return ErrAdminInputInvalid
	}
	return nil
}

func newAdminID(prefix string) (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return prefix + hex.EncodeToString(buf), nil
}
