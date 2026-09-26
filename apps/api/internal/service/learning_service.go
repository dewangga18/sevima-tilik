package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
)

const LearningRuleVersion = "learning-demo-v1"

var ErrLearningNotFound = errors.New("learning session or lesson not found")
var ErrInvalidLearningRequest = errors.New("invalid learning request")

type LearningService struct {
	learning   *repository.LearningRepository
	curriculum *repository.CurriculumRepository
	diagnostic *DiagnosticService
}

func NewLearningService(learning *repository.LearningRepository, curriculum *repository.CurriculumRepository, diagnostic *DiagnosticService) *LearningService {
	return &LearningService{learning: learning, curriculum: curriculum, diagnostic: diagnostic}
}

func learningStatus(correct int) domain.SkillStatus {
	if correct == 3 {
		return domain.StatusStrongEvidence
	}
	if correct == 2 {
		return domain.StatusInconclusive
	}
	return domain.StatusNeedsPractice
}
func learningScore(correct, total int) int { return (correct*100 + total/2) / total }

func (s *LearningService) Lesson(ctx context.Context, skillID string) (*domain.Lesson, error) {
	lesson, err := s.learning.GetLesson(ctx, skillID)
	if err != nil {
		return nil, err
	}
	if lesson == nil {
		return nil, ErrLearningNotFound
	}
	return lesson, nil
}

func prepareLearningForClient(session *domain.LearningSession) {
	for i := range session.Items {
		item := &session.Items[i]
		if item.Question == nil {
			continue
		}
		item.Question.AnswerKey = ""
		if item.AnsweredAt == nil || (item.Stage == "reassessment" && session.Stage != "completed" && session.Stage != "exhausted") {
			item.IsCorrect = nil
			item.Question.Explanation = ""
			item.Question.Misconception = ""
		}
	}
}

func (s *LearningService) ownedSession(ctx context.Context, studentID, id string) (*domain.LearningSession, error) {
	session, err := s.learning.GetSession(ctx, id)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrLearningNotFound
	}
	if session.StudentID != studentID {
		return nil, ErrForbidden
	}
	return session, nil
}
func (s *LearningService) Session(ctx context.Context, studentID, id string) (*domain.LearningSession, error) {
	session, err := s.ownedSession(ctx, studentID, id)
	if err != nil {
		return nil, err
	}
	prepareLearningForClient(session)
	return session, nil
}

func (s *LearningService) Progress(ctx context.Context, studentID string) (*domain.LearningProgress, error) {
	skills, err := s.diagnostic.GetSkills(ctx, 4)
	if err != nil {
		return nil, err
	}
	progress := &domain.LearningProgress{Skills: make([]domain.SkillProgress, 0, len(skills)), LearningPath: make([]domain.LearningPathItem, 0)}
	byID := make(map[string]domain.SkillProgress)
	for _, skill := range skills {
		byID[skill.ID] = domain.SkillProgress{SkillID: skill.ID, SkillName: skill.Name, Status: domain.StatusUnassessed, Source: "unassessed"}
	}
	id, err := s.learning.LatestDiagnosticID(ctx, studentID)
	if err != nil {
		return nil, err
	}
	var diagnosticAt *time.Time
	if id != "" {
		assessment, err := s.diagnostic.GetAssessment(ctx, studentID, id)
		if err != nil {
			return nil, err
		}
		progress.SourceAssessmentID = id
		diagnosticAt = assessment.CompletedAt
		for _, result := range assessment.Results {
			p := domain.SkillProgress{SkillID: result.SkillID, SkillName: result.SkillName, Status: result.Status, EvidenceCount: result.EvidenceCount, CorrectCount: result.TotalCorrect, Source: "diagnostic", UpdatedAt: diagnosticAt}
			if result.TotalAnswered > 0 {
				score := learningScore(result.TotalCorrect, result.TotalAnswered)
				p.Score = &score
			}
			if result.TotalAnswered == 0 {
				p.Source = "unassessed"
				p.UpdatedAt = nil
			}
			byID[p.SkillID] = p
		}
	}
	stored, err := s.learning.StoredProgress(ctx, studentID)
	if err != nil {
		return nil, err
	}
	for _, p := range stored {
		if diagnosticAt == nil || p.UpdatedAt.After(*diagnosticAt) {
			byID[p.SkillID] = p
		}
	}
	results := make([]domain.SkillResult, 0, len(skills))
	for _, skill := range skills {
		p := byID[skill.ID]
		progress.Skills = append(progress.Skills, p)
		results = append(results, domain.SkillResult{SkillID: p.SkillID, SkillName: p.SkillName, Status: p.Status})
	}
	progress.LearningPath, err = progressivePath(skills, results)
	if err != nil {
		return nil, err
	}
	for i := range progress.LearningPath {
		item := &progress.LearningPath[i]
		if byID[item.SkillID].Source == "reassessment" {
			item.Reason = "Perlu ditinjau berdasarkan reassessment terakhir."
		}
	}
	progress.ActiveSessions, progress.RecentSessions, progress.CompletedCount, err = s.learning.SessionSummaries(ctx, studentID)
	if err != nil {
		return nil, err
	}
	return progress, nil
}

func (s *LearningService) Start(ctx context.Context, studentID, sourceID, skillID, requestID string) (*domain.LearningSession, error) {
	if sourceID == "" || skillID == "" || strings.TrimSpace(requestID) == "" || len(requestID) > 100 {
		return nil, ErrInvalidLearningRequest
	}
	source, err := s.diagnostic.GetAssessment(ctx, studentID, sourceID)
	if err != nil {
		return nil, err
	}
	if source.Status != domain.AssessmentCompleted || source.RuleVersion != ProgressiveRuleVersion || source.GradeLevel != 4 {
		return nil, ErrInvalidLearningRequest
	}
	allowed := false
	for _, result := range source.Results {
		if result.SkillID == skillID {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, ErrInvalidLearningRequest
	}
	if _, err = s.Lesson(ctx, skillID); err != nil {
		return nil, err
	}
	progress, err := s.Progress(ctx, studentID)
	if err != nil {
		return nil, err
	}
	var before *int
	for _, p := range progress.Skills {
		if p.SkillID == skillID {
			before = p.Score
			break
		}
	}
	id, err := diagnosticID("learn-")
	if err != nil {
		return nil, err
	}
	session := &domain.LearningSession{ID: id, StudentID: studentID, SourceAssessmentID: sourceID, SkillID: skillID, RequestID: requestID, RuleVersion: LearningRuleVersion, Stage: "lesson", StartedAt: time.Now(), BeforeScore: before}
	created, err := s.learning.StartSession(ctx, session)
	if err != nil {
		return nil, err
	}
	return s.Session(ctx, studentID, created)
}

func selectLearningQuestion(bank []domain.Question, desired int) *domain.Question {
	if len(bank) == 0 {
		return nil
	}
	distance := func(n int) int {
		if n < 0 {
			return -n
		}
		return n
	}
	sort.Slice(bank, func(i, j int) bool {
		di, dj := distance(bank[i].Difficulty-desired), distance(bank[j].Difficulty-desired)
		if di != dj {
			return di < dj
		}
		if bank[i].Difficulty != bank[j].Difficulty {
			return bank[i].Difficulty < bank[j].Difficulty
		}
		return bank[i].ID < bank[j].ID
	})
	return &bank[0]
}
func learningItem(q *domain.Question, stage string, order int) (*domain.LearningItem, error) {
	id, err := diagnosticID("li-")
	return &domain.LearningItem{ID: id, QuestionID: q.ID, Stage: stage, OrderIndex: order, Question: q}, err
}

func (s *LearningService) CompleteLesson(ctx context.Context, studentID, id string) (*domain.LearningSession, error) {
	if id == "" {
		return nil, ErrInvalidLearningRequest
	}
	for retry := 0; retry < 3; retry++ {
		session, err := s.ownedSession(ctx, studentID, id)
		if err != nil {
			return nil, err
		}
		if session.Stage != "lesson" {
			prepareLearningForClient(session)
			return session, nil
		}
		bank, err := s.learning.AvailableQuestions(ctx, studentID, session.SkillID, "practice")
		if err != nil {
			return nil, err
		}
		desired := 1
		if session.BeforeScore != nil && *session.BeforeScore == 100 {
			desired = 2
		}
		q := selectLearningQuestion(bank, desired)
		var next *domain.LearningItem
		now := time.Now()
		session.LessonCompletedAt = &now
		if q == nil {
			session.Stage = "exhausted"
			session.StopReason = "insufficient_questions"
		} else {
			session.Stage = "practice"
			next, err = learningItem(q, "practice", 1)
			if err != nil {
				return nil, err
			}
		}
		err = s.learning.SaveStep(ctx, session, nil, next)
		if errors.Is(err, repository.ErrLearningChanged) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return s.Session(ctx, studentID, id)
	}
	return nil, repository.ErrLearningChanged
}

func (s *LearningService) Answer(ctx context.Context, studentID, id, questionID, answer string) (*domain.LearningSession, error) {
	answer = strings.TrimSpace(answer)
	if id == "" || questionID == "" || answer == "" {
		return nil, ErrInvalidLearningRequest
	}
	for retry := 0; retry < 3; retry++ {
		session, err := s.ownedSession(ctx, studentID, id)
		if err != nil {
			return nil, err
		}
		var item *domain.LearningItem
		for i := range session.Items {
			if session.Items[i].QuestionID == questionID {
				item = &session.Items[i]
				break
			}
		}
		if item == nil {
			return nil, ErrInvalidLearningRequest
		}
		if item.AnsweredAt != nil {
			if item.StudentAnswer != answer {
				return nil, repository.ErrLearningChanged
			}
			prepareLearningForClient(session)
			return session, nil
		}
		if item.Stage != session.Stage || (session.Stage != "practice" && session.Stage != "reassessment") {
			return nil, repository.ErrLearningChanged
		}
		question, err := s.curriculum.GetQuestionByID(ctx, questionID)
		if err != nil {
			return nil, err
		}
		valid := false
		for _, option := range question.Options {
			if option == answer {
				valid = true
				break
			}
		}
		if !valid {
			return nil, ErrInvalidLearningRequest
		}
		correct := answer == question.AnswerKey
		now := time.Now()
		item.StudentAnswer = answer
		item.IsCorrect = &correct
		item.AnsweredAt = &now
		count, right := 0, 0
		for _, record := range session.Items {
			if record.Stage == session.Stage && record.AnsweredAt != nil {
				count++
				if record.IsCorrect != nil && *record.IsCorrect {
					right++
				}
			}
		}
		desired := question.Difficulty - 1
		if correct {
			desired = question.Difficulty + 1
		}
		if session.Stage == "practice" {
			// A review prompt is not a diagnosis; no score changes during practice.
			wrongRun := 0
			for i := len(session.Items) - 1; i >= 0; i-- {
				record := session.Items[i]
				if record.Stage != "practice" || record.IsCorrect == nil || *record.IsCorrect {
					break
				}
				wrongRun++
			}
			if wrongRun >= 2 {
				skills, err := s.diagnostic.GetSkills(ctx, 4)
				if err != nil {
					return nil, err
				}
				for _, skill := range skills {
					if skill.ID == session.SkillID && len(skill.Prereqs) > 0 {
						prereqs := append([]string(nil), skill.Prereqs...)
						sort.Strings(prereqs)
						session.ReviewSkillID = prereqs[0]
					}
				}
			}
			if count == 3 {
				session.Stage = "reassessment"
				desired = 1
				if correct {
					desired = 2
				}
			}
		} else if count == 3 {
			session.Stage = "completed"
			session.CompletedAt = &now
			score := learningScore(right, 3)
			session.Score = &score
			session.Outcome = learningStatus(right)
		}
		var next *domain.LearningItem
		if session.Stage != "completed" {
			bank, err := s.learning.AvailableQuestions(ctx, studentID, session.SkillID, session.Stage)
			if err != nil {
				return nil, err
			}
			q := selectLearningQuestion(bank, desired)
			if q == nil {
				session.Stage = "exhausted"
				session.StopReason = "insufficient_questions"
			} else {
				next, err = learningItem(q, session.Stage, len(session.Items)+1)
				if err != nil {
					return nil, err
				}
			}
		}
		err = s.learning.SaveStep(ctx, session, item, next)
		if errors.Is(err, repository.ErrLearningChanged) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return s.Session(ctx, studentID, id)
	}
	return nil, repository.ErrLearningChanged
}
