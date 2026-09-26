package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
)

const ProgressiveRuleVersion = "progressive-demo-v1"
const diagnosticQuestionLimit = 18
const evidenceMinimum = 3

var ErrInvalidDiagnosticAnswer = errors.New("invalid progressive diagnostic answer")
var ErrAnswerImmutable = errors.New("committed answer cannot be changed")
var ErrUnsupportedGrade = errors.New("unsupported diagnostic grade")

type diagnosticEvidence struct{ answered, correct int }

func progressiveResults(skills []domain.Skill, items []domain.AssessmentItem, target string) []domain.SkillResult {
	stats := make(map[string]diagnosticEvidence)
	for _, item := range items {
		if item.AnsweredAt == nil || item.Question == nil || item.IsCorrect == nil {
			continue
		}
		stat := stats[item.Question.SkillID]
		stat.answered++
		if *item.IsCorrect {
			stat.correct++
		}
		stats[item.Question.SkillID] = stat
	}
	results := make([]domain.SkillResult, 0, len(skills))
	for _, skill := range skills {
		stat := stats[skill.ID]
		status := domain.StatusUnassessed
		confidence := "low"
		if stat.answered > 0 {
			status = domain.StatusInconclusive
		}
		if stat.answered >= evidenceMinimum {
			confidence = "medium"
			if stat.correct == stat.answered {
				status = domain.StatusStrongEvidence
			} else if stat.correct <= 1 {
				status = domain.StatusNeedsPractice
			}
		}
		results = append(results, domain.SkillResult{SkillID: skill.ID, SkillName: skill.Name, Status: status, TotalAnswered: stat.answered, TotalCorrect: stat.correct, EvidenceCount: stat.answered, Confidence: confidence})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].SkillID < results[j].SkillID })
	byID := make(map[string]int)
	graph := make(map[string]domain.Skill)
	for i, result := range results {
		byID[result.SkillID] = i
	}
	for _, skill := range skills {
		graph[skill.ID] = skill
	}
	visited := make(map[string]bool)
	var trace func(string)
	trace = func(id string) {
		if visited[id] {
			return
		}
		visited[id] = true
		i, exists := byID[id]
		if !exists || results[i].Status != domain.StatusNeedsPractice {
			return
		}
		results[i].RelatedTargetSkillID = target
		allStrong := true
		for _, prerequisite := range graph[id].Prereqs {
			j, exists := byID[prerequisite]
			if !exists || results[j].Status != domain.StatusStrongEvidence {
				allStrong = false
			}
			trace(prerequisite)
		}
		// The visible difficulty itself is not labeled a root prerequisite gap.
		results[i].IsRootGap = id != target && allStrong
	}
	trace(target)
	return results
}

// A weak branch probes prerequisites; strong or mixed evidence does not invent a gap.
func chooseProgressiveQuestion(skills []domain.Skill, bank []domain.Question, items []domain.AssessmentItem, target string) (*domain.Question, string, string, error) {
	if _, err := prerequisiteOrder(skills); err != nil {
		return nil, "", "", err
	}
	results := progressiveResults(skills, items, target)
	byID := make(map[string]domain.SkillResult)
	graph := make(map[string]domain.Skill)
	for _, result := range results {
		byID[result.SkillID] = result
	}
	for _, skill := range skills {
		graph[skill.ID] = skill
	}
	if _, exists := graph[target]; !exists {
		return nil, "", "", fmt.Errorf("target skill missing")
	}
	used := make(map[string]bool)
	answered := 0
	for _, item := range items {
		used[item.QuestionID] = true
		if item.AnsweredAt != nil {
			answered++
		}
	}
	if answered >= diagnosticQuestionLimit {
		return nil, "", "question_limit", nil
	}
	visited := make(map[string]bool)
	exhausted := false
	var choose func(string, string) (*domain.Question, string)
	choose = func(id, parent string) (*domain.Question, string) {
		if visited[id] {
			return nil, ""
		}
		visited[id] = true
		result := byID[id]
		if result.EvidenceCount < evidenceMinimum {
			desired := 1
			for _, item := range items {
				if item.Question != nil && item.Question.SkillID == id && item.AnsweredAt != nil && item.IsCorrect != nil {
					desired = item.Question.Difficulty - 1
					if *item.IsCorrect {
						desired = item.Question.Difficulty + 1
					}
				}
			}
			candidates := make([]domain.Question, 0)
			for _, q := range bank {
				if q.SkillID == id && !used[q.ID] {
					candidates = append(candidates, q)
				}
			}
			distance := func(n int) int {
				if n < 0 {
					return -n
				}
				return n
			}
			sort.Slice(candidates, func(i, j int) bool {
				di, dj := distance(candidates[i].Difficulty-desired), distance(candidates[j].Difficulty-desired)
				if di != dj {
					return di < dj
				}
				if candidates[i].Difficulty != candidates[j].Difficulty {
					return candidates[i].Difficulty < candidates[j].Difficulty
				}
				return candidates[i].ID < candidates[j].ID
			})
			if len(candidates) > 0 {
				return &candidates[0], parent
			}
			exhausted = true
			return nil, ""
		}
		if result.Status == domain.StatusNeedsPractice {
			prerequisites := append([]string(nil), graph[id].Prereqs...)
			sort.Strings(prerequisites)
			for _, prerequisite := range prerequisites {
				if q, probe := choose(prerequisite, id); q != nil {
					return q, probe
				}
			}
		}
		return nil, ""
	}
	if q, probe := choose(target, ""); q != nil {
		return q, probe, "", nil
	}
	reason := "evidence_complete"
	if exhausted {
		reason = "insufficient_evidence"
	}
	for _, result := range results {
		if result.Status == domain.StatusInconclusive {
			reason = "insufficient_evidence"
		}
	}
	return nil, "", reason, nil
}

func progressivePath(skills []domain.Skill, results []domain.SkillResult) ([]domain.LearningPathItem, error) {
	order, err := prerequisiteOrder(skills)
	if err != nil {
		return nil, err
	}
	byID := make(map[string]domain.SkillResult)
	for _, result := range results {
		byID[result.SkillID] = result
	}
	path := make([]domain.LearningPathItem, 0)
	for _, id := range order {
		result := byID[id]
		if result.Status != domain.StatusNeedsPractice && result.Status != domain.StatusInconclusive {
			continue
		}
		reason := "Perlu penguatan berdasarkan jawaban diagnostic."
		if result.IsRootGap {
			reason = "Kandidat gap prasyarat; tinjau fondasi ini sebelum materi berikutnya."
		}
		if result.Status == domain.StatusInconclusive {
			reason = "Bukti belum cukup untuk menyimpulkan pemahaman; perlu review tambahan."
		}
		path = append(path, domain.LearningPathItem{SkillID: id, SkillName: result.SkillName, Reason: reason})
	}
	return path, nil
}

func diagnosticID(prefix string) (string, error) {
	value := make([]byte, 12)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return prefix + hex.EncodeToString(value), nil
}

func newProgressiveItem(assessmentID string, question *domain.Question, order int, probe string) (domain.AssessmentItem, error) {
	id, err := diagnosticID("item-")
	return domain.AssessmentItem{ID: id, AssessmentID: assessmentID, QuestionID: question.ID, OrderIndex: order, Question: question, ProbeForSkillID: probe}, err
}

func (s *DiagnosticService) StartProgressiveDiagnostic(ctx context.Context, studentID string, grade int) (*domain.Assessment, error) {
	if grade != 4 {
		return nil, ErrUnsupportedGrade
	}
	latest, err := s.assessmentRepo.GetLatestByStudent(ctx, studentID)
	if err != nil {
		return nil, err
	}
	if latest != nil && latest.Status == domain.AssessmentInProgress {
		prepareAssessmentForClient(latest)
		return latest, nil
	}
	skills, err := s.GetSkills(ctx, grade)
	if err != nil {
		return nil, err
	}
	bank, err := s.curriculumRepo.GetQuestionBank(ctx)
	if err != nil {
		return nil, err
	}
	const target = "frac_cmp_diff_den"
	q, probe, _, err := chooseProgressiveQuestion(skills, bank, nil, target)
	if err != nil {
		return nil, err
	}
	if q == nil {
		return nil, errors.New("diagnostic target question bank is empty")
	}
	id, err := diagnosticID("asm-")
	if err != nil {
		return nil, err
	}
	item, err := newProgressiveItem(id, q, 1, probe)
	if err != nil {
		return nil, err
	}
	assessment := &domain.Assessment{ID: id, StudentID: studentID, GradeLevel: grade, Status: domain.AssessmentInProgress, StartedAt: time.Now(), RuleVersion: ProgressiveRuleVersion, TargetSkillID: target, MaxQuestions: diagnosticQuestionLimit, Items: []domain.AssessmentItem{item}}
	createdID, err := s.assessmentRepo.CreateProgressiveAssessment(ctx, assessment)
	if err != nil {
		return nil, err
	}
	return s.GetAssessment(ctx, studentID, createdID)
}

func (s *DiagnosticService) AnswerDiagnostic(ctx context.Context, studentID, assessmentID, questionID, answer string) (*domain.Assessment, error) {
	answer = strings.TrimSpace(answer)
	if assessmentID == "" || questionID == "" || answer == "" {
		return nil, ErrInvalidDiagnosticAnswer
	}
	for retry := 0; retry < 3; retry++ {
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
		if assessment.RuleVersion != ProgressiveRuleVersion {
			return nil, ErrInvalidDiagnosticAnswer
		}
		var item *domain.AssessmentItem
		for i := range assessment.Items {
			if assessment.Items[i].QuestionID == questionID {
				item = &assessment.Items[i]
				break
			}
		}
		if item == nil {
			return nil, ErrInvalidDiagnosticAnswer
		}
		if item.AnsweredAt != nil {
			if item.StudentAnswer != answer {
				return nil, ErrAnswerImmutable
			}
			prepareAssessmentForClient(assessment)
			return assessment, nil
		}
		if assessment.Status != domain.AssessmentInProgress {
			return nil, ErrAnswerImmutable
		}
		q, err := s.curriculumRepo.GetQuestionByID(ctx, questionID)
		if err != nil {
			return nil, err
		}
		valid := false
		for _, option := range q.Options {
			if option == answer {
				valid = true
				break
			}
		}
		if !valid {
			return nil, ErrInvalidDiagnosticAnswer
		}
		correct := strings.EqualFold(answer, strings.TrimSpace(q.AnswerKey))
		now := time.Now()
		item.StudentAnswer = answer
		item.IsCorrect = &correct
		item.AnsweredAt = &now
		skills, err := s.GetSkills(ctx, assessment.GradeLevel)
		if err != nil {
			return nil, err
		}
		bank, err := s.curriculumRepo.GetQuestionBank(ctx)
		if err != nil {
			return nil, err
		}
		nextQuestion, probe, reason, err := chooseProgressiveQuestion(skills, bank, assessment.Items, assessment.TargetSkillID)
		if err != nil {
			return nil, err
		}
		var next *domain.AssessmentItem
		if nextQuestion != nil {
			nextItem, err := newProgressiveItem(assessment.ID, nextQuestion, len(assessment.Items)+1, probe)
			if err != nil {
				return nil, err
			}
			next = &nextItem
		} else {
			assessment.Status = domain.AssessmentCompleted
			assessment.StopReason = reason
			assessment.Results = progressiveResults(skills, assessment.Items, assessment.TargetSkillID)
			assessment.LearningPath, err = progressivePath(skills, assessment.Results)
			if err != nil {
				return nil, err
			}
		}
		err = s.assessmentRepo.SaveProgressiveStep(ctx, assessment, *item, next)
		if errors.Is(err, repository.ErrAssessmentChanged) {
			continue
		}
		if err != nil {
			return nil, err
		}
		return s.GetAssessment(ctx, studentID, assessmentID)
	}
	return nil, repository.ErrAssessmentChanged
}
