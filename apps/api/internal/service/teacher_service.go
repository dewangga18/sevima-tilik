package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
)

var ErrClassForbidden = errors.New("classroom is not assigned to this teacher")

type TeacherService struct {
	repo *repository.DB
}

func NewTeacherService(repo *repository.DB) *TeacherService {
	return &TeacherService{repo: repo}
}

func (s *TeacherService) GetClasses(ctx context.Context, teacherID string) ([]domain.TeacherClass, error) {
	classes, err := s.repo.GetTeacherClasses(ctx, teacherID)
	if err != nil {
		return nil, fmt.Errorf("teacher classes: %w", err)
	}
	return classes, nil
}

func (s *TeacherService) GetClassStudents(ctx context.Context, teacherID, classroomID string) ([]domain.StudentOverview, error) {
	assigned, err := s.repo.IsClassAssigned(ctx, teacherID, classroomID)
	if err != nil {
		return nil, fmt.Errorf("check assignment: %w", err)
	}
	if !assigned {
		return nil, ErrClassForbidden
	}
	students, err := s.repo.GetClassStudents(ctx, classroomID)
	if err != nil {
		return nil, fmt.Errorf("class students: %w", err)
	}
	return students, nil
}

func (s *TeacherService) GetStudentInsight(ctx context.Context, teacherID, studentID string) (*domain.StudentInsight, error) {
	taught, err := s.repo.IsStudentTaughtBy(ctx, teacherID, studentID)
	if err != nil {
		return nil, fmt.Errorf("check taught student: %w", err)
	}
	if !taught {
		return nil, ErrClassForbidden
	}
	insight, err := s.repo.GetStudentInsight(ctx, studentID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, sql.ErrNoRows
	}
	if err != nil {
		return nil, fmt.Errorf("student insight: %w", err)
	}
	return insight, nil
}
