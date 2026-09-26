package service

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
)

func progressiveTestService(t *testing.T) (*DiagnosticService, *repository.DB) {
	db := diagnosticTestDB(t)
	return NewDiagnosticService(repository.NewCurriculumRepository(db), repository.NewAssessmentRepository(db)), db
}

func TestProgressiveDiagnosticScenarios(t *testing.T) {
	for _, scenario := range []string{"strong", "root gap", "mixed", "exhaustion"} {
		t.Run(scenario, func(t *testing.T) {
			s, db := progressiveTestService(t)
			ctx := context.Background()
			if scenario == "exhaustion" {
				if _, err := db.ExecContext(ctx, "DELETE FROM questions WHERE id='q-cmp-diff-3'"); err != nil {
					t.Fatal(err)
				}
			}
			assessment, err := s.StartProgressiveDiagnostic(ctx, "u-student-new", 4)
			if err != nil {
				t.Fatal(err)
			}
			if len(assessment.Items) != 1 || assessment.RuleVersion != ProgressiveRuleVersion {
				t.Fatal("wrong initial progressive state")
			}
			counts := make(map[string]int)
			seen := make(map[string]bool)
			for assessment.Status == domain.AssessmentInProgress {
				if len(assessment.Items) > diagnosticQuestionLimit {
					t.Fatal("question limit exceeded")
				}
				pending := assessment.Items[len(assessment.Items)-1]
				for _, item := range assessment.Items {
					if item.IsCorrect != nil || item.Question.Explanation != "" || item.Question.AnswerKey != "" {
						t.Fatal("answer evidence disclosed before completion")
					}
				}
				if seen[pending.QuestionID] {
					t.Fatal("question repeated")
				}
				seen[pending.QuestionID] = true
				question, err := s.curriculumRepo.GetQuestionByID(ctx, pending.QuestionID)
				if err != nil {
					t.Fatal(err)
				}
				counts[question.SkillID]++
				correct := true
				if scenario == "root gap" {
					correct = question.SkillID != "frac_cmp_diff_den" && question.SkillID != "frac_equiv" && question.SkillID != "mul_basic"
				}
				if scenario == "mixed" {
					correct = counts[question.SkillID] < 3
				}
				answer := question.AnswerKey
				if !correct {
					for _, option := range question.Options {
						if option != question.AnswerKey {
							answer = option
							break
						}
					}
				}
				before := assessment.Revision
				assessment, err = s.AnswerDiagnostic(ctx, assessment.StudentID, assessment.ID, question.ID, answer)
				if err != nil {
					t.Fatal(err)
				}
				if assessment.Revision != before+1 {
					t.Fatal("answer revision did not advance")
				}
				retried, err := s.AnswerDiagnostic(ctx, assessment.StudentID, assessment.ID, question.ID, answer)
				if err != nil || !reflect.DeepEqual(assessment, retried) {
					t.Fatalf("retry changed evidence: %v", err)
				}
				for _, option := range question.Options {
					if option != answer {
						if _, err := s.AnswerDiagnostic(ctx, assessment.StudentID, assessment.ID, question.ID, option); !errors.Is(err, ErrAnswerImmutable) {
							t.Fatalf("committed answer changed: %v", err)
						}
						break
					}
				}
			}
			if assessment.CompletedAt == nil || assessment.Items[0].Question.Explanation == "" {
				t.Fatal("completion not persisted")
			}
			if scenario == "strong" {
				if len(assessment.Items) != 3 || len(assessment.LearningPath) != 0 {
					t.Fatal("strong target should stop without invented gaps")
				}
				for _, result := range assessment.Results {
					if result.SkillID != assessment.TargetSkillID && result.Status != domain.StatusUnassessed {
						t.Fatal("untested prerequisite marked assessed")
					}
				}
			}
			if scenario == "root gap" {
				roots := []string{}
				for _, result := range assessment.Results {
					if result.IsRootGap {
						roots = append(roots, result.SkillID)
					}
				}
				if !reflect.DeepEqual(roots, []string{"mul_basic"}) {
					t.Fatalf("wrong root candidates %v", roots)
				}
				path := []string{}
				for _, item := range assessment.LearningPath {
					path = append(path, item.SkillID)
				}
				if !reflect.DeepEqual(path, []string{"mul_basic", "frac_equiv", "frac_cmp_diff_den"}) {
					t.Fatalf("path not prerequisite first: %v", path)
				}
				probed := false
				for _, item := range assessment.Items {
					if item.Question.SkillID == "mul_basic" && item.ProbeForSkillID == "frac_equiv" {
						probed = true
					}
				}
				if !probed {
					t.Fatal("missing prerequisite provenance")
				}
			}
			if scenario == "mixed" || scenario == "exhaustion" {
				if assessment.StopReason != "insufficient_evidence" {
					t.Fatalf("wrong stop: %s", assessment.StopReason)
				}
				for _, result := range assessment.Results {
					if result.IsRootGap || result.Status == domain.StatusStrongEvidence {
						t.Fatal("insufficient evidence became confident result")
					}
				}
			}
			latest, err := s.GetLatestAssessment(ctx, assessment.StudentID)
			if err != nil || !reflect.DeepEqual(latest, assessment) {
				t.Fatal("resume lost completion")
			}
			if _, err := s.GetAssessment(ctx, "u-student-1", assessment.ID); !errors.Is(err, ErrForbidden) {
				t.Fatal("other student read result")
			}
		})
	}
}

func TestProgressiveConcurrentStartsAndAnswers(t *testing.T) {
	s, _ := progressiveTestService(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	starts := make(chan *domain.Assessment, 2)
	failures := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			assessment, err := s.StartProgressiveDiagnostic(ctx, "u-student-new", 4)
			if err != nil {
				failures <- err
			} else {
				starts <- assessment
			}
		}()
	}
	wg.Wait()
	close(starts)
	if len(failures) > 0 {
		t.Fatal(<-failures)
	}
	var first *domain.Assessment
	for assessment := range starts {
		if first == nil {
			first = assessment
		} else if first.ID != assessment.ID {
			t.Fatal("parallel starts created two attempts")
		}
	}
	question := first.Items[0].Question
	answers := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.AnswerDiagnostic(ctx, first.StudentID, first.ID, question.ID, question.Options[0])
			answers <- err
		}()
	}
	wg.Wait()
	close(answers)
	for err := range answers {
		if err != nil {
			t.Fatal(err)
		}
	}
	latest, err := s.GetLatestAssessment(ctx, first.StudentID)
	if err != nil || latest.Revision != 1 || len(latest.Items) != 2 {
		t.Fatalf("duplicate answer/next: %+v %v", latest, err)
	}
	history, err := s.GetHistory(ctx, first.StudentID)
	if err != nil || len(history.Items) != 1 || history.Items[0].CorrectCount != nil {
		t.Fatal("history exposed in-progress correctness")
	}

	if _, err := s.AnswerDiagnostic(ctx, "u-student-1", first.ID, question.ID, question.Options[0]); !errors.Is(err, ErrForbidden) {
		t.Fatal("ownership bypass")
	}
	if _, err := s.AnswerDiagnostic(ctx, first.StudentID, first.ID, "not-offered", question.Options[0]); !errors.Is(err, ErrInvalidDiagnosticAnswer) {
		t.Fatal("future question accepted")
	}
	if _, err := s.SubmitDiagnostic(ctx, first.StudentID, first.ID, nil); !errors.Is(err, ErrInvalidDiagnosticAnswer) {
		t.Fatal("bulk submit bypass")
	}
	if _, err := s.AnswerDiagnostic(ctx, first.StudentID, first.ID, latest.Items[1].QuestionID, "invalid option"); !errors.Is(err, ErrInvalidDiagnosticAnswer) {
		t.Fatal("invalid option accepted")
	}
	if _, err := s.StartProgressiveDiagnostic(ctx, first.StudentID, 7); !errors.Is(err, ErrUnsupportedGrade) {
		t.Fatal("unsupported grade accepted")
	}
}

func TestProgressiveAnswerRollback(t *testing.T) {
	s, db := progressiveTestService(t)
	ctx := context.Background()
	a, err := s.StartProgressiveDiagnostic(ctx, "u-student-new", 4)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE assessment_items ADD CONSTRAINT test_no_answer CHECK (student_answer IS NULL OR student_answer='')`); err != nil {
		t.Fatal(err)
	}
	_, err = s.AnswerDiagnostic(ctx, a.StudentID, a.ID, a.Items[0].QuestionID, a.Items[0].Question.Options[0])
	if err == nil || !strings.Contains(err.Error(), "test_no_answer") {
		t.Fatalf("expected storage failure: %v", err)
	}
	saved, err := s.GetAssessment(ctx, a.StudentID, a.ID)
	if err != nil || saved.Revision != 0 || len(saved.Items) != 1 || saved.Items[0].AnsweredAt != nil {
		t.Fatal("failed answer partially persisted")
	}
}

func TestProgressiveReplayAndConflictingAnswers(t *testing.T) {
	s, _ := progressiveTestService(t)
	ctx := context.Background()
	skills, err := s.GetSkills(ctx, 4)
	if err != nil {
		t.Fatal(err)
	}
	bank, err := s.curriculumRepo.GetQuestionBank(ctx)
	if err != nil {
		t.Fatal(err)
	}
	first, _, _, err := chooseProgressiveQuestion(skills, bank, nil, "frac_cmp_diff_den")
	if err != nil {
		t.Fatal(err)
	}
	for i, j := 0, len(bank)-1; i < j; i, j = i+1, j-1 {
		bank[i], bank[j] = bank[j], bank[i]
	}
	replay, _, _, err := chooseProgressiveQuestion(skills, bank, nil, "frac_cmp_diff_den")
	if err != nil || first.ID != replay.ID {
		t.Fatal("selection depends on query order")
	}
	a, err := s.StartProgressiveDiagnostic(ctx, "u-student-new", 4)
	if err != nil {
		t.Fatal(err)
	}
	q := a.Items[0].Question
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, option := range q.Options[:2] {
		wg.Add(1)
		go func(answer string) {
			defer wg.Done()
			_, err := s.AnswerDiagnostic(ctx, a.StudentID, a.ID, q.ID, answer)
			results <- err
		}(option)
	}
	wg.Wait()
	close(results)
	accepted, rejected := 0, 0
	for err := range results {
		if err == nil {
			accepted++
		} else if errors.Is(err, ErrAnswerImmutable) {
			rejected++
		} else {
			t.Fatal(err)
		}
	}
	if accepted != 1 || rejected != 1 {
		t.Fatalf("concurrent answers %d accepted %d rejected", accepted, rejected)
	}
	saved, err := s.GetAssessment(ctx, a.StudentID, a.ID)
	if err != nil || saved.Revision != 1 || len(saved.Items) != 2 {
		t.Fatal("conflicting submission changed committed state")
	}
}
