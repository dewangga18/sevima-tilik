package repository

import (
	"regexp"
	"strconv"
	"testing"
)

func TestLearningBankMathematicalConsistency(t *testing.T) {
	digits := regexp.MustCompile(`\d+`)
	prompts := make(map[string]bool)
	for _, lesson := range learningLessons() {
		if len(lesson.Steps) != 3 || lesson.Hint == "" {
			t.Fatal("lesson missing focused steps or hint")
		}
		for _, purpose := range []string{"practice", "reassessment"} {
			for index := 0; index < 6; index++ {
				q := learningQuestion(lesson.SkillID, purpose, index)
				if prompts[q.Prompt] {
					t.Fatalf("repeated learning prompt: %s", q.Prompt)
				}
				prompts[q.Prompt] = true
				options := make(map[string]bool)
				for _, option := range q.Options {
					if options[option] {
						t.Fatalf("duplicate option %s", q.ID)
					}
					options[option] = true
				}
				if !options[q.AnswerKey] || q.Explanation == "" {
					t.Fatalf("invalid key/explanation %s", q.ID)
				}
				tokens := digits.FindAllString(q.Prompt, -1)
				numbers := make([]int, len(tokens))
				for i, token := range tokens {
					numbers[i], _ = strconv.Atoi(token)
				}
				expected := ""
				switch q.SkillID {
				case "mul_basic":
					expected = strconv.Itoa(numbers[0] * numbers[1])
				case "div_basic":
					if numbers[0]%numbers[1] != 0 {
						t.Fatal("division not exact")
					}
					expected = strconv.Itoa(numbers[0] / numbers[1])
				case "frac_rep":
					expected = strconv.Itoa(numbers[1]) + "/" + strconv.Itoa(numbers[0])
				case "frac_equiv":
					expected = strconv.Itoa(numbers[0]*numbers[2]) + "/" + strconv.Itoa(numbers[1]*numbers[2])
				default:
					expected = "="
					if numbers[0]*numbers[3] < numbers[2]*numbers[1] {
						expected = "<"
					}
					if numbers[0]*numbers[3] > numbers[2]*numbers[1] {
						expected = ">"
					}
				}
				if q.AnswerKey != expected {
					t.Fatalf("mathematics mismatch %s: got %s expected %s", q.ID, q.AnswerKey, expected)
				}
			}
		}
	}
	if len(prompts) != 72 {
		t.Fatalf("wrong bank size %d", len(prompts))
	}
}
