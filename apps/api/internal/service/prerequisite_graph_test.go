package service

import (
	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
	"reflect"
	"testing"
)

func TestPrerequisiteOrder(t *testing.T) {
	skills := []domain.Skill{{ID: "comparison", Prereqs: []string{"representation", "equivalence"}}, {ID: "equivalence", Prereqs: []string{"multiplication", "representation"}}, {ID: "representation"}, {ID: "multiplication"}}
	want := []string{"multiplication", "representation", "equivalence", "comparison"}
	for i := 0; i < 2; i++ {
		order, err := prerequisiteOrder(skills)
		if err != nil || !reflect.DeepEqual(order, want) {
			t.Fatalf("order %v error %v", order, err)
		}
		for j, k := 0, len(skills)-1; j < k; j, k = j+1, k-1 {
			skills[j], skills[k] = skills[k], skills[j]
		}
	}
}

func TestPrerequisiteGraphRejectsInvalidData(t *testing.T) {
	cases := map[string][]domain.Skill{
		"cycle":                {{ID: "a", Prereqs: []string{"b"}}, {ID: "b", Prereqs: []string{"a"}}},
		"self cycle":           {{ID: "a", Prereqs: []string{"a"}}},
		"missing prerequisite": {{ID: "a", Prereqs: []string{"missing"}}},
		"duplicate skill":      {{ID: "a"}, {ID: "a"}},
		"duplicate edge":       {{ID: "a"}, {ID: "b", Prereqs: []string{"a", "a"}}},
		"empty ID":             {{ID: ""}},
	}
	for name, skills := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := prerequisiteOrder(skills); err == nil {
				t.Fatal("invalid graph accepted")
			}
		})
	}
	order, err := prerequisiteOrder(nil)
	if err != nil || len(order) != 0 {
		t.Fatalf("empty curriculum: %v %v", order, err)
	}
}

func TestInvalidCurriculumDoesNotCreateAttempt(t *testing.T) {
	db := diagnosticTestDB(t)
	ctx := t.Context()
	s := NewDiagnosticService(repository.NewCurriculumRepository(db), repository.NewAssessmentRepository(db))
	if _, err := db.ExecContext(ctx, "INSERT INTO skill_prerequisites(skill_id, prereq_id) VALUES ('mul_basic','frac_equiv')"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.StartDiagnostic(ctx, "u-student-new", 4); err == nil {
		t.Fatal("cyclic curriculum started diagnostic")
	}
	history, err := s.GetHistory(ctx, "u-student-new")
	if err != nil || len(history.Items) != 0 {
		t.Fatalf("failed start persisted assessment: %+v %v", history, err)
	}
}

func TestEmptyQuestionBankDoesNotCreateAttempt(t *testing.T) {
	db := diagnosticTestDB(t)
	ctx := t.Context()
	s := NewDiagnosticService(repository.NewCurriculumRepository(db), repository.NewAssessmentRepository(db))
	if _, err := db.ExecContext(ctx, "DELETE FROM questions"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.StartDiagnostic(ctx, "u-student-new", 4); err == nil {
		t.Fatal("empty bank started diagnostic")
	}
	history, err := s.GetHistory(ctx, "u-student-new")
	if err != nil || len(history.Items) != 0 {
		t.Fatalf("empty bank persisted assessment: %+v %v", history, err)
	}
}
