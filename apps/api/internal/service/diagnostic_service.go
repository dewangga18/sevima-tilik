package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
)

var (
	ErrAssessmentNotFound  = errors.New("assessment not found")
	ErrForbidden           = errors.New("forbidden")
	ErrAssessmentCompleted = repository.ErrAssessmentCompleted
)

type DiagnosticService struct {
	curriculumRepo *repository.CurriculumRepository
	assessmentRepo *repository.AssessmentRepository
}

func NewDiagnosticService(curriculumRepo *repository.CurriculumRepository, assessmentRepo *repository.AssessmentRepository) *DiagnosticService {
	return &DiagnosticService{
		curriculumRepo: curriculumRepo,
		assessmentRepo: assessmentRepo,
	}
}

func (s *DiagnosticService) StartDiagnostic(ctx context.Context, studentID string, gradeLevel int) (*domain.Assessment, error) {
	// Check if student already has an in-progress assessment to resume
	latest, err := s.assessmentRepo.GetLatestByStudent(ctx, studentID)
	if err != nil {
		return nil, fmt.Errorf("check latest assessment: %w", err)
	}
	if latest != nil && latest.Status == domain.AssessmentInProgress {
		prepareAssessmentForClient(latest)
		return latest, nil
	}

	// Fetch diagnostic questions
	questions, err := s.curriculumRepo.GetInitialDiagnosticQuestions(ctx, gradeLevel)
	if err != nil {
		return nil, fmt.Errorf("get diagnostic questions: %w", err)
	}

	randBytes := make([]byte, 8)
	rand.Read(randBytes)
	assessmentID := fmt.Sprintf("asm-%s", hex.EncodeToString(randBytes))

	items := make([]domain.AssessmentItem, len(questions))
	for i, q := range questions {
		randItemBytes := make([]byte, 6)
		rand.Read(randItemBytes)
		items[i] = domain.AssessmentItem{
			ID:           fmt.Sprintf("item-%s", hex.EncodeToString(randItemBytes)),
			AssessmentID: assessmentID,
			QuestionID:   q.ID,
			OrderIndex:   i + 1,
			Question: &domain.Question{
				ID:         q.ID,
				SkillID:    q.SkillID,
				Difficulty: q.Difficulty,
				Prompt:     q.Prompt,
				Options:    q.Options,
			},
		}
	}

	assessment := &domain.Assessment{
		ID:         assessmentID,
		StudentID:  studentID,
		GradeLevel: gradeLevel,
		Status:     domain.AssessmentInProgress,
		StartedAt:  time.Now(),
		Items:      items,
	}

	if err := s.assessmentRepo.CreateAssessment(ctx, assessment); err != nil {
		return nil, fmt.Errorf("create assessment: %w", err)
	}

	return assessment, nil
}

type AnswerSubmission struct {
	QuestionID    string `json:"question_id"`
	StudentAnswer string `json:"student_answer"`
}

func (s *DiagnosticService) SubmitDiagnostic(ctx context.Context, studentID, assessmentID string, answers []AnswerSubmission) (*domain.Assessment, error) {
	assessment, err := s.assessmentRepo.GetAssessmentByID(ctx, assessmentID)
	if err != nil {
		return nil, err
	}
	if assessment == nil {
		return nil, ErrAssessmentNotFound
	}
	if assessment.StudentID != studentID {
		return nil, ErrForbidden
	}
	if assessment.Status != domain.AssessmentInProgress {
		return nil, ErrAssessmentCompleted
	}

	// Fetch all skills to ensure unassessed skills are also listed
	skills, err := s.curriculumRepo.GetSkills(ctx, assessment.GradeLevel)
	if err != nil {
		return nil, fmt.Errorf("get skills: %w", err)
	}
	skillMap := make(map[string]domain.Skill)
	for _, sk := range skills {
		skillMap[sk.ID] = sk
	}

	// Create answer lookup
	ansMap := make(map[string]string)
	for _, a := range answers {
		ansMap[a.QuestionID] = strings.TrimSpace(a.StudentAnswer)
	}

	type skillStat struct {
		totalAnswered int
		totalCorrect  int
	}
	stats := make(map[string]*skillStat)

	evaluatedItems := make([]domain.AssessmentItem, len(assessment.Items))
	for i, item := range assessment.Items {
		q, err := s.curriculumRepo.GetQuestionByID(ctx, item.QuestionID)
		if err != nil {
			return nil, fmt.Errorf("get question %s: %w", item.QuestionID, err)
		}

		studentAns, hasAns := ansMap[item.QuestionID]
		isCorrect := false
		if hasAns && studentAns != "" {
			isCorrect = (strings.EqualFold(studentAns, strings.TrimSpace(q.AnswerKey)))
			if stats[q.SkillID] == nil {
				stats[q.SkillID] = &skillStat{}
			}
			stats[q.SkillID].totalAnswered++
			if isCorrect {
				stats[q.SkillID].totalCorrect++
			}
		}

		evaluatedItems[i] = domain.AssessmentItem{
			ID:            item.ID,
			AssessmentID:  assessmentID,
			QuestionID:    item.QuestionID,
			StudentAnswer: studentAns,
			IsCorrect:     &isCorrect,
			Question:      q,
		}
	}

	// Compute results per skill
	var results []domain.SkillResult
	for _, sk := range skills {
		stat := stats[sk.ID]
		status := domain.StatusUnassessed
		totalAns := 0
		totalCorr := 0
		evidenceCount := 0
		confidence := "low"

		if stat != nil && stat.totalAnswered > 0 {
			totalAns = stat.totalAnswered
			totalCorr = stat.totalCorrect
			evidenceCount = totalAns

			if totalCorr == totalAns {
				status = domain.StatusMastered
			} else {
				status = domain.StatusNeedsPractice
			}

			if evidenceCount >= 3 {
				confidence = "high"
			} else if evidenceCount >= 2 {
				confidence = "medium"
			} else {
				confidence = "low"
			}
		}

		results = append(results, domain.SkillResult{
			SkillID:       sk.ID,
			SkillName:     sk.Name,
			Status:        status,
			TotalAnswered: totalAns,
			TotalCorrect:  totalCorr,
			EvidenceCount: evidenceCount,
			Confidence:    confidence,
			IsRootGap:     false, // Will be computed with DAG backtracking in Phase 2
		})
	}

	if err := s.assessmentRepo.SaveEvaluation(ctx, assessmentID, evaluatedItems, results, studentID); err != nil {
		return nil, fmt.Errorf("save evaluation: %w", err)
	}

	return s.GetAssessment(ctx, studentID, assessmentID)
}

func (s *DiagnosticService) GetAssessment(ctx context.Context, requestingUserID string, assessmentID string) (*domain.Assessment, error) {
	assessment, err := s.assessmentRepo.GetAssessmentByID(ctx, assessmentID)
	if err != nil {
		return nil, err
	}
	if assessment == nil {
		return nil, ErrAssessmentNotFound
	}

	// Student can only view their own assessment
	if assessment.StudentID != requestingUserID {
		return nil, ErrForbidden
	}

	prepareAssessmentForClient(assessment)
	return assessment, nil
}

func (s *DiagnosticService) GetLatestAssessment(ctx context.Context, studentID string) (*domain.Assessment, error) {
	assessment, err := s.assessmentRepo.GetLatestByStudent(ctx, studentID)
	if err != nil {
		return nil, err
	}
	prepareAssessmentForClient(assessment)
	return assessment, nil
}

func prepareAssessmentForClient(assessment *domain.Assessment) {
	if assessment == nil {
		return
	}
	for i := range assessment.Items {
		question := assessment.Items[i].Question
		if question == nil {
			continue
		}
		question.AnswerKey = ""
		if assessment.Status != domain.AssessmentCompleted {
			question.Explanation = ""
			question.Misconception = ""
		}
	}
}

func (s *DiagnosticService) GetSkills(ctx context.Context, gradeLevel int) ([]domain.Skill, error) {
	return s.curriculumRepo.GetSkills(ctx, gradeLevel)
}

func (s *DiagnosticService) GetHistory(ctx context.Context, studentID string) (*domain.AssessmentHistory, error) {
	return s.assessmentRepo.GetHistoryByStudent(ctx, studentID)
}
