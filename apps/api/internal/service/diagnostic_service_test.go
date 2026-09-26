package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
)

func TestAssessmentDisclosure(t *testing.T) {
	for _, status := range []domain.AssessmentStatus{domain.AssessmentInProgress, domain.AssessmentCompleted} {
		t.Run(string(status), func(t *testing.T) {
			a := &domain.Assessment{Status: status, Items: []domain.AssessmentItem{{Question: &domain.Question{
				AnswerKey: "secret answer", Explanation: "solution", Misconception: "answer clue",
			}}, {}}}
			prepareAssessmentForClient(a)
			body, err := json.Marshal(a)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(body), "secret answer") {
				t.Fatal("answer key exposed")
			}
			if strings.Contains(string(body), "solution") != (status == domain.AssessmentCompleted) {
				t.Fatalf("wrong explanation visibility: %s", body)
			}
			if status == domain.AssessmentInProgress && strings.Contains(string(body), "answer clue") {
				t.Fatal("misconception exposes a clue")
			}
		})
	}
	prepareAssessmentForClient(nil)
}

// Each integration test uses its own schema so existing demo data is untouched.
func diagnosticTestDB(t *testing.T) *repository.DB {
	return testDB(t, true)
}

func testDB(t *testing.T, demoUsers bool) *repository.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is required for PostgreSQL integration tests")
	}
	admin, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("diagnostic_test_%d", time.Now().UnixNano())
	if _, err := admin.ExecContext(context.Background(), "CREATE SCHEMA "+schema); err != nil {
		admin.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if _, err := admin.ExecContext(context.Background(), "DROP SCHEMA "+schema+" CASCADE"); err != nil {
			t.Errorf("clean test schema: %v", err)
		}
		admin.Close()
	})
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	sqlDB, err := sql.Open("pgx", parsed.String())
	if err != nil {
		t.Fatal(err)
	}
	db := &repository.DB{DB: sqlDB}
	t.Cleanup(func() { db.Close() })
	if err := db.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	if demoUsers {
		if err := db.SeedDemoUsers(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.Seed(context.Background()); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestDiagnosticLifecycle(t *testing.T) {
	db := diagnosticTestDB(t)
	ctx := context.Background()
	s := NewDiagnosticService(repository.NewCurriculumRepository(db), repository.NewAssessmentRepository(db))
	studentID := "u-student-1"
	started, err := s.StartDiagnostic(ctx, studentID, 4)
	if err != nil {
		t.Fatal(err)
	}
	assertHidden := func(a *domain.Assessment, err error) {
		t.Helper()
		if err != nil || a == nil || len(a.Items) == 0 {
			t.Fatalf("assessment missing: %v", err)
		}
		for _, item := range a.Items {
			if item.Question.Explanation != "" || item.Question.AnswerKey != "" || item.Question.Misconception != "" {
				t.Fatal("in-progress assessment exposes solution")
			}
		}
	}
	assertHidden(started, nil)
	assertHidden(s.StartDiagnostic(ctx, studentID, 4))
	assertHidden(s.GetLatestAssessment(ctx, studentID))
	assertHidden(s.GetAssessment(ctx, studentID, started.ID))
	if _, err := s.SubmitDiagnostic(ctx, "u-teacher-1", started.ID, nil); !errors.Is(err, ErrForbidden) {
		t.Fatalf("wrong owner was not denied: %v", err)
	}
	completed, err := s.SubmitDiagnostic(ctx, studentID, started.ID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != domain.AssessmentCompleted || completed.CompletedAt == nil || completed.Items[0].Question.Explanation == "" {
		t.Fatal("completed assessment lacks result/explanations")
	}
	if _, err := s.SubmitDiagnostic(ctx, studentID, started.ID, []AnswerSubmission{{QuestionID: started.Items[0].QuestionID, StudentAnswer: "12"}}); !errors.Is(err, ErrAssessmentCompleted) {
		t.Fatalf("completed assessment accepted a new submission: %v", err)
	}
	latest, err := s.GetLatestAssessment(ctx, studentID)
	if err != nil || !reflect.DeepEqual(completed, latest) {
		t.Fatalf("completed record changed after resubmission: %v", err)
	}
}

func TestConcurrentCompletionAndRollback(t *testing.T) {
	db := diagnosticTestDB(t)
	ctx := context.Background()
	repo := repository.NewAssessmentRepository(db)
	s := NewDiagnosticService(repository.NewCurriculumRepository(db), repo)
	a, err := s.StartDiagnostic(ctx, "u-student-1", 4)
	if err != nil {
		t.Fatal(err)
	}
	correct := true
	item := a.Items[0]
	item.IsCorrect = &correct
	// Evidence failure must roll back the completion claim and answer update.
	err = repo.SaveEvaluation(ctx, a.ID, []domain.AssessmentItem{item}, []domain.SkillResult{{SkillID: "missing-skill"}}, a.StudentID)
	if err == nil {
		t.Fatal("invalid evidence should fail")
	}
	unchanged, err := repo.GetAssessmentByID(ctx, a.ID)
	if err != nil || unchanged.Status != domain.AssessmentInProgress || unchanged.CompletedAt != nil || unchanged.Items[0].IsCorrect != nil {
		t.Fatalf("failed transaction persisted changes: %v", err)
	}
	gate := make(chan struct{})
	results := make(chan error, 2)
	for _, answer := range []string{"first", "second"} {
		go func(answer string) {
			<-gate
			attempt := item
			attempt.StudentAnswer = answer
			results <- repo.SaveEvaluation(ctx, a.ID, []domain.AssessmentItem{attempt}, nil, a.StudentID)
		}(answer)
	}
	close(gate)
	winners, conflicts := 0, 0
	for i := 0; i < 2; i++ {
		err := <-results
		if err == nil {
			winners++
		} else if errors.Is(err, repository.ErrAssessmentCompleted) {
			conflicts++
		} else {
			t.Fatal(err)
		}
	}
	if winners != 1 || conflicts != 1 {
		t.Fatalf("want one winner and one conflict, got %d/%d", winners, conflicts)
	}
	completed, err := repo.GetAssessmentByID(ctx, a.ID)
	if err != nil || completed.Status != domain.AssessmentCompleted || completed.Items[0].StudentAnswer == "" {
		t.Fatalf("winning submission missing: %v", err)
	}
	if err := repo.SaveEvaluation(ctx, a.ID, nil, nil, a.StudentID); !errors.Is(err, repository.ErrAssessmentCompleted) {
		t.Fatalf("repository permits overwrite: %v", err)
	}
	after, err := repo.GetAssessmentByID(ctx, a.ID)
	if err != nil || !reflect.DeepEqual(completed, after) {
		t.Fatal("retry changed winning submission")
	}
}

func TestStudentHistoryAndOwnership(t *testing.T) {
	db := diagnosticTestDB(t)
	ctx := context.Background()
	s := NewDiagnosticService(repository.NewCurriculumRepository(db), repository.NewAssessmentRepository(db))
	history, err := s.GetHistory(ctx, "u-student-new")
	if err != nil || len(history.Items) != 0 || history.CompletedCount != 0 {
		t.Fatalf("fresh history: %+v %v", history, err)
	}
	for i := 0; i < 7; i++ {
		assessment, err := s.StartDiagnostic(ctx, "u-student-new", 4)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = s.GetAssessment(ctx, "u-student-1", assessment.ID); !errors.Is(err, ErrForbidden) {
			t.Fatalf("ownership error: %v", err)
		}
		answers := make([]AnswerSubmission, len(assessment.Items))
		for j, item := range assessment.Items {
			answers[j] = AnswerSubmission{QuestionID: item.QuestionID, StudentAnswer: item.Question.Options[0]}
		}
		if _, err = s.SubmitDiagnostic(ctx, "u-student-new", assessment.ID, answers); err != nil {
			t.Fatal(err)
		}
	}
	history, err = s.GetHistory(ctx, "u-student-new")
	if err != nil {
		t.Fatal(err)
	}
	if history.CompletedCount != 7 || len(history.Items) != 5 {
		t.Fatalf("history limit and total: %+v", history)
	}
	for _, item := range history.Items {
		if item.AnsweredCount != item.QuestionCount || item.AssessedSkillCount == 0 {
			t.Fatalf("missing evidence: %+v", item)
		}
	}
	other, err := s.GetHistory(ctx, "u-student-1")
	if err != nil || len(other.Items) != 0 {
		t.Fatalf("history leaked across owners: %+v %v", other, err)
	}
}
