package service

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
)

func learningTestSetup(t *testing.T, strongSource bool) (*LearningService, *DiagnosticService, *repository.DB, *domain.Assessment) {
	t.Helper()
	db := diagnosticTestDB(t)
	ctx := context.Background()
	if err := db.SeedLearning(ctx); err != nil {
		t.Fatal(err)
	}
	curriculum := repository.NewCurriculumRepository(db)
	diagnostic := NewDiagnosticService(curriculum, repository.NewAssessmentRepository(db))
	learning := NewLearningService(repository.NewLearningRepository(db), curriculum, diagnostic)
	bank, err := curriculum.GetQuestionBank(ctx)
	if err != nil || len(bank) != 18 {
		t.Fatal("learning seed polluted diagnostic bank")
	}
	source, err := diagnostic.StartProgressiveDiagnostic(ctx, "u-student-new", 4)
	if err != nil {
		t.Fatal(err)
	}
	for source.Status == domain.AssessmentInProgress {
		item := source.Items[len(source.Items)-1]
		q, err := curriculum.GetQuestionByID(ctx, item.QuestionID)
		if err != nil {
			t.Fatal(err)
		}
		answer := q.AnswerKey
		if !strongSource && q.SkillID == "frac_cmp_diff_den" {
			for _, option := range q.Options {
				if option != q.AnswerKey {
					answer = option
					break
				}
			}
		}
		source, err = diagnostic.AnswerDiagnostic(ctx, source.StudentID, source.ID, q.ID, answer)
		if err != nil {
			t.Fatal(err)
		}
	}
	return learning, diagnostic, db, source
}

func answerLearning(t *testing.T, s *LearningService, session *domain.LearningSession, correct bool) *domain.LearningSession {
	t.Helper()
	pending := session.Items[len(session.Items)-1]
	question, err := s.curriculum.GetQuestionByID(context.Background(), pending.QuestionID)
	if err != nil {
		t.Fatal(err)
	}
	answer := question.AnswerKey
	if !correct {
		for _, option := range question.Options {
			if option != answer {
				answer = option
				break
			}
		}
	}
	next, err := s.Answer(context.Background(), session.StudentID, session.ID, pending.QuestionID, answer)
	if err != nil {
		t.Fatal(err)
	}
	retry, err := s.Answer(context.Background(), session.StudentID, session.ID, pending.QuestionID, answer)
	if err != nil || !reflect.DeepEqual(next, retry) {
		t.Fatalf("retry changed learning result: %v", err)
	}
	return next
}

func TestLearningLoopUpdatesProgressWithoutChangingDiagnostic(t *testing.T) {
	learning, diagnostic, _, source := learningTestSetup(t, false)
	ctx := context.Background()
	session, err := learning.Start(ctx, source.StudentID, source.ID, "frac_cmp_diff_den", "request-1")
	if err != nil {
		t.Fatal(err)
	}
	if session.Stage != "lesson" || session.BeforeScore == nil || *session.BeforeScore != 0 || session.Score != nil {
		t.Fatal("invalid baseline lesson")
	}
	alias, err := learning.Start(ctx, source.StudentID, source.ID, session.SkillID, "alias-request")
	if err != nil || alias.ID != session.ID {
		t.Fatal("active start did not resume")
	}
	session, err = learning.CompleteLesson(ctx, session.StudentID, session.ID)
	if err != nil {
		t.Fatal(err)
	}
	duplicate, err := learning.CompleteLesson(ctx, session.StudentID, session.ID)
	if err != nil || !reflect.DeepEqual(session, duplicate) {
		t.Fatal("lesson retry added work")
	}
	before, err := learning.Progress(ctx, session.StudentID)
	if err != nil {
		t.Fatal(err)
	}
	for _, skill := range before.Skills {
		if skill.SkillID == session.SkillID && (skill.Score == nil || *skill.Score != 0) {
			t.Fatal("lesson completion raised score")
		}
	}
	seen := make(map[string]bool)
	for _, item := range source.Items {
		seen[item.QuestionID] = true
	}
	for session.Stage == "practice" || session.Stage == "reassessment" {
		pending := session.Items[len(session.Items)-1]
		if seen[pending.QuestionID] {
			t.Fatal("practice/reassessment repeated a question")
		}
		seen[pending.QuestionID] = true
		if pending.Question.Explanation != "" || pending.IsCorrect != nil {
			t.Fatal("pending learning solution leaked")
		}
		if session.Stage == "reassessment" {
			for _, item := range session.Items {
				if item.Stage == "reassessment" && (item.IsCorrect != nil || item.Question.Explanation != "") {
					t.Fatal("reassessment solution leaked before completion")
				}
			}
		}
		session = answerLearning(t, learning, session, true)
		if len(session.Items) == 2 && (pending.Question.Difficulty != 1 || session.Items[1].Question.Difficulty != 2) {
			t.Fatal("correct practice did not increase difficulty")
		}
	}
	if session.Stage != "completed" || session.Score == nil || *session.Score != 100 || session.Outcome != domain.StatusStrongEvidence {
		t.Fatal("completion score not based on reassessment")
	}
	progress, err := learning.Progress(ctx, session.StudentID)
	if err != nil {
		t.Fatal(err)
	}
	if progress.CompletedCount != 1 || len(progress.LearningPath) != 0 {
		t.Fatalf("improvement did not update path: %+v", progress)
	}
	original, err := diagnostic.GetAssessment(ctx, source.StudentID, source.ID)
	if err != nil || !reflect.DeepEqual(source, original) {
		t.Fatal("learning rewrote diagnostic snapshot")
	}
	retryStart, err := learning.Start(ctx, source.StudentID, source.ID, session.SkillID, "alias-request")
	if err != nil || retryStart.ID != session.ID || retryStart.Stage != "completed" {
		t.Fatal("start retry after completion created new session")
	}
	if _, err := learning.Session(ctx, "u-student-1", session.ID); !errors.Is(err, ErrForbidden) {
		t.Fatal("other student read learning")
	}
	if _, err := learning.Start(ctx, "u-student-1", source.ID, session.SkillID, "other"); !errors.Is(err, ErrForbidden) {
		t.Fatal("other student used diagnostic source")
	}
}

func TestLearningNonImprovementAndQuestionExhaustion(t *testing.T) {
	learning, _, _, source := learningTestSetup(t, false)
	ctx := context.Background()
	for attempt := 0; attempt < 2; attempt++ {
		request := []string{"first", "second"}[attempt]
		session, err := learning.Start(ctx, source.StudentID, source.ID, "frac_cmp_diff_den", request)
		if err != nil {
			t.Fatal(err)
		}
		session, err = learning.CompleteLesson(ctx, session.StudentID, session.ID)
		if err != nil {
			t.Fatal(err)
		}
		for session.Stage == "practice" || session.Stage == "reassessment" {
			session = answerLearning(t, learning, session, false)
		}
		if session.Score == nil || *session.Score != 0 || session.Outcome != domain.StatusNeedsPractice || session.ReviewSkillID == "" {
			t.Fatal("wrong answers lost gap/prerequisite review")
		}
		progress, err := learning.Progress(ctx, source.StudentID)
		if err != nil || len(progress.LearningPath) == 0 {
			t.Fatal("nonimprovement removed review path")
		}
	}
	if _, err := learning.Start(ctx, source.StudentID, source.ID, "frac_cmp_diff_den", "third"); !errors.Is(err, repository.ErrLearningBankEmpty) {
		t.Fatalf("bank reuse allowed: %v", err)
	}
	progress, err := learning.Progress(ctx, source.StudentID)
	if err != nil || progress.CompletedCount != 2 {
		t.Fatal("exhaustion created completion")
	}
}

func TestLearningDecliningAndUnassessedBaselines(t *testing.T) {
	for _, skill := range []string{"frac_cmp_diff_den", "mul_basic"} {
		t.Run(skill, func(t *testing.T) {
			learning, _, _, source := learningTestSetup(t, true)
			ctx := context.Background()
			session, err := learning.Start(ctx, source.StudentID, source.ID, skill, "baseline")
			if err != nil {
				t.Fatal(err)
			}
			if skill == "mul_basic" && session.BeforeScore != nil {
				t.Fatal("unassessed baseline invented zero")
			}
			if skill == "frac_cmp_diff_den" && (session.BeforeScore == nil || *session.BeforeScore != 100) {
				t.Fatal("strong baseline lost")
			}
			session, err = learning.CompleteLesson(ctx, session.StudentID, session.ID)
			if err != nil {
				t.Fatal(err)
			}
			for session.Stage == "practice" || session.Stage == "reassessment" {
				session = answerLearning(t, learning, session, false)
			}
			if session.Score == nil || *session.Score != 0 || session.Outcome != domain.StatusNeedsPractice {
				t.Fatal("decline not reflected")
			}
		})
	}
}

func TestLearningConcurrentAnswersAndCompletionRollback(t *testing.T) {
	learning, _, db, source := learningTestSetup(t, false)
	ctx := context.Background()
	session, err := learning.Start(ctx, source.StudentID, source.ID, "frac_cmp_diff_den", "concurrent")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	failures := make(chan error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := learning.CompleteLesson(ctx, source.StudentID, session.ID)
			failures <- err
		}()
	}
	wg.Wait()
	for i := 0; i < 2; i++ {
		if err := <-failures; err != nil {
			t.Fatal(err)
		}
	}
	session, err = learning.Session(ctx, source.StudentID, session.ID)
	if err != nil || session.Revision != 1 || len(session.Items) != 1 {
		t.Fatal("concurrent lesson completion duplicated work")
	}
	pending := session.Items[0]
	q, err := learning.curriculum.GetQuestionByID(ctx, pending.QuestionID)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := learning.Answer(ctx, source.StudentID, session.ID, q.ID, q.AnswerKey)
			failures <- err
		}()
	}
	wg.Wait()
	for i := 0; i < 2; i++ {
		if err := <-failures; err != nil {
			t.Fatal(err)
		}
	}
	session, err = learning.Session(ctx, source.StudentID, session.ID)
	if err != nil || session.Revision != 2 || len(session.Items) != 2 {
		t.Fatal("concurrent answer duplicated work")
	}
	for len(session.Items) < 6 {
		session = answerLearning(t, learning, session, true)
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE skill_progress ADD CONSTRAINT test_no_perfect_score CHECK(score<100)`); err != nil {
		t.Fatal(err)
	}
	pending = session.Items[5]
	q, err = learning.curriculum.GetQuestionByID(ctx, pending.QuestionID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := learning.Answer(ctx, source.StudentID, session.ID, q.ID, q.AnswerKey); err == nil {
		t.Fatal("expected progress storage failure")
	}
	saved, err := learning.Session(ctx, source.StudentID, session.ID)
	if err != nil || saved.Revision != session.Revision || saved.Items[5].AnsweredAt != nil || saved.Stage != "reassessment" {
		t.Fatal("completion partially persisted")
	}
	if _, err := db.ExecContext(ctx, `ALTER TABLE skill_progress DROP CONSTRAINT test_no_perfect_score`); err != nil {
		t.Fatal(err)
	}
	session = answerLearning(t, learning, saved, true)
	if session.Stage != "completed" || *session.Score != 100 {
		t.Fatal("failed completion could not retry")
	}
}

func TestLearningConcurrentStartsAndImmutableAnswers(t *testing.T) {
	learning, _, _, source := learningTestSetup(t, false)
	ctx := context.Background()
	results := make(chan *domain.LearningSession, 2)
	errorsChannel := make(chan error, 2)
	var wg sync.WaitGroup
	for _, key := range []string{"start-a", "start-b"} {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			session, err := learning.Start(ctx, source.StudentID, source.ID, "frac_cmp_diff_den", key)
			results <- session
			errorsChannel <- err
		}(key)
	}
	wg.Wait()
	first, second := <-results, <-results
	for i := 0; i < 2; i++ {
		if err := <-errorsChannel; err != nil {
			t.Fatal(err)
		}
	}
	if first.ID != second.ID {
		t.Fatal("concurrent starts created two active sessions")
	}
	if _, err := learning.Start(ctx, source.StudentID, source.ID, "mul_basic", "start-a"); !errors.Is(err, repository.ErrLearningChanged) {
		t.Fatal("request key reused with different payload")
	}
	session, err := learning.CompleteLesson(ctx, source.StudentID, first.ID)
	if err != nil {
		t.Fatal(err)
	}
	pending := session.Items[0]
	session = answerLearning(t, learning, session, true)
	q, err := learning.curriculum.GetQuestionByID(ctx, pending.QuestionID)
	if err != nil {
		t.Fatal(err)
	}
	for _, option := range q.Options {
		if option != q.AnswerKey {
			if _, err := learning.Answer(ctx, source.StudentID, session.ID, q.ID, option); !errors.Is(err, repository.ErrLearningChanged) {
				t.Fatal("stored answer was overwritten")
			}
			break
		}
	}
}

func TestLearningMixedEvidenceAndNewDiagnostic(t *testing.T) {
	learning, diagnostic, _, source := learningTestSetup(t, false)
	ctx := context.Background()
	session, err := learning.Start(ctx, source.StudentID, source.ID, "frac_cmp_diff_den", "mixed")
	if err != nil {
		t.Fatal(err)
	}
	session, err = learning.CompleteLesson(ctx, source.StudentID, session.ID)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 6; i++ {
		session = answerLearning(t, learning, session, i != 5)
	}
	if *session.Score != 67 || session.Outcome != domain.StatusInconclusive {
		t.Fatal("mixed evidence incorrectly classified")
	}
	progress, err := learning.Progress(ctx, source.StudentID)
	if err != nil || len(progress.LearningPath) == 0 {
		t.Fatal("mixed evidence removed review path")
	}
	next, err := diagnostic.StartProgressiveDiagnostic(ctx, source.StudentID, 4)
	if err != nil {
		t.Fatal(err)
	}
	for next.Status == domain.AssessmentInProgress {
		item := next.Items[len(next.Items)-1]
		q, err := learning.curriculum.GetQuestionByID(ctx, item.QuestionID)
		if err != nil {
			t.Fatal(err)
		}
		next, err = diagnostic.AnswerDiagnostic(ctx, source.StudentID, next.ID, q.ID, q.AnswerKey)
		if err != nil {
			t.Fatal(err)
		}
	}
	progress, err = learning.Progress(ctx, source.StudentID)
	if err != nil {
		t.Fatal(err)
	}
	for _, skill := range progress.Skills {
		if skill.SkillID == "frac_cmp_diff_den" && (skill.Score == nil || *skill.Score != 100 || skill.Source != "diagnostic") {
			t.Fatal("old reassessment overrode newer diagnostic")
		}
	}
}

func TestLearningExhaustionDuringReassessmentKeepsProgress(t *testing.T) {
	learning, _, db, source := learningTestSetup(t, false)
	ctx := context.Background()
	session, err := learning.Start(ctx, source.StudentID, source.ID, "frac_cmp_diff_den", "exhaust")
	if err != nil {
		t.Fatal(err)
	}
	session, err = learning.CompleteLesson(ctx, source.StudentID, session.ID)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		session = answerLearning(t, learning, session, true)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM questions q WHERE q.skill_id=$1 AND q.purpose='reassessment' AND NOT EXISTS(SELECT 1 FROM learning_items i WHERE i.question_id=q.id)`, session.SkillID); err != nil {
		t.Fatal(err)
	}
	session = answerLearning(t, learning, session, true)
	if session.Stage != "exhausted" || session.Score != nil || session.CompletedAt != nil {
		t.Fatal("insufficient reassessment created mastery")
	}
	progress, err := learning.Progress(ctx, source.StudentID)
	if err != nil {
		t.Fatal(err)
	}
	if progress.CompletedCount != 0 {
		t.Fatal("exhaustion counted as completion")
	}
	for _, skill := range progress.Skills {
		if skill.SkillID == session.SkillID && (skill.Score == nil || *skill.Score != 0 || skill.Source != "diagnostic") {
			t.Fatal("partial evidence changed baseline")
		}
	}
}
