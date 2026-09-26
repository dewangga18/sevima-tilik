package service

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
)

const fixtureQuestion = `{"question":"Berapa hasil 3 × 4?","options":[{"id":"A","text":"12"},{"id":"B","text":"7"},{"id":"C","text":"16"},{"id":"D","text":"9"}],"correct_answer":"A","explanation":"3 kelompok masing-masing 4 berarti 4 + 4 + 4 = 12."}`

func fixtureInput() AIGenerateInput {
	return AIGenerateInput{RequestID: "test-request-123456789", SkillID: "multiplication", GradeLevel: 4, Difficulty: 1, Purpose: "practice"}
}
func TestAIStorageEncryption(t *testing.T) {
	key := base64.StdEncoding.EncodeToString(make([]byte, 32))
	s := NewAISettingsService(nil, key)
	first, err := s.seal("test-secret-key-value")
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.seal("test-secret-key-value")
	if err != nil {
		t.Fatal(err)
	}
	if string(first) == string(second) || strings.Contains(string(first), "test-secret") {
		t.Fatal("encryption must use fresh nonce")
	}
	plain, err := s.open(first)
	if err != nil || plain != "test-secret-key-value" {
		t.Fatal("roundtrip failed")
	}
	first[len(first)-1] ^= 1
	if _, err = s.open(first); err == nil {
		t.Fatal("tampered ciphertext accepted")
	}
	if _, err = NewAISettingsService(nil, "").open(second); err == nil {
		t.Fatal("missing encryption key accepted")
	}
}
func TestGeneratedOutputGuards(t *testing.T) {
	input := fixtureInput()
	for _, test := range []struct {
		name, body string
		valid      bool
	}{
		{"valid", fixtureQuestion, true},
		{"unknown field", strings.TrimSuffix(fixtureQuestion, "}") + `,"publish":true}`, false},
		{"missing key", strings.Replace(fixtureQuestion, `"correct_answer":"A"`, `"correct_answer":"Z"`, 1), false},
		{"numeric equivalence", strings.Replace(fixtureQuestion, `"7"`, `"12,0"`, 1), false},
		{"duplicate IDs", strings.Replace(fixtureQuestion, `"id":"B"`, `"id":"A"`, 1), false},
		{"trailing JSON", fixtureQuestion + ` {}`, false},
		{"missing explanation", strings.Replace(fixtureQuestion, `"3 kelompok masing-masing 4 berarti 4 + 4 + 4 = 12."`, `""`, 1), false},
	} {
		t.Run(test.name, func(t *testing.T) {
			entry, err := generatedEntry([]byte(test.body), input, "gemini-2.5-flash", "Perkalian dasar", "numerasi")
			if (err == nil) != test.valid {
				t.Fatalf("unexpected validation: %v", err)
			}
			if err == nil && entry.Status != "draft" {
				t.Fatal("must remain draft")
			}
		})
	}
}
func TestGeminiTransport(t *testing.T) {
	for _, test := range []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"success", 200, `{"candidates":[{"finishReason":"STOP","content":{"parts":[{"text":` + quoted(fixtureQuestion) + `}]}}]}`, nil},
		{"invalid key", 400, `{"error":{"message":"test-secret-key-value"}}`, ErrAIProvider},
		{"quota", 429, `{}`, ErrAIQuota},
		{"truncated", 200, `{"candidates":[{"finishReason":"MAX_TOKENS","content":{"parts":[{"text":"{}"}]}}]}`, ErrAIOutput},
		{"blocked", 200, `{"promptFeedback":{"blockReason":"SAFETY"}}`, ErrAIOutput},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("x-goog-api-key") != "test-secret-key-value" || strings.Contains(r.URL.String(), "test-secret") {
					t.Error("key must be header only")
				}
				var body map[string]any
				if json.NewDecoder(r.Body).Decode(&body) != nil {
					t.Error("invalid request")
				}
				config, ok := body["generationConfig"].(map[string]any)
				if !ok || config["responseJsonSchema"] == nil {
					t.Error("schema missing")
				}
				w.WriteHeader(test.status)
				w.Write([]byte(test.body))
			}))
			defer server.Close()
			client := NewGeminiClient()
			client.baseURL = server.URL + "/"
			client.client = server.Client()
			_, err := client.Generate(context.Background(), "test-secret-key-value", "gemini-2.5-flash", "fixture", questionSchema())
			if !errors.Is(err, test.want) {
				t.Fatalf("want %v got %v", test.want, err)
			}
			if err != nil && strings.Contains(err.Error(), "test-secret") {
				t.Fatal("secret in error")
			}
		})
	}
}
func quoted(value string) string { b, _ := json.Marshal(value); return string(b) }

// Uses only an explicitly supplied local test database. All fixture rows are
// removed, including activation, so the real bank and student progress stay intact.
func TestAIDraftLocalIntegration(t *testing.T) {
	url := os.Getenv("AI_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set AI_TEST_DATABASE_URL for local PostgreSQL checkpoint")
	}
	rawDB, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	defer rawDB.Close()
	db := &repository.DB{DB: rawDB}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	input := fixtureInput()
	input.RequestID = "test-" + time.Now().Format("20060102150405.000000000")
	input.SkillID = ""
	if err = db.QueryRowContext(ctx, `SELECT s.id FROM skills s JOIN lessons l ON l.skill_id=s.id WHERE s.name ILIKE '%perkalian%' LIMIT 1`).Scan(&input.SkillID); err != nil {
		t.Fatal(err)
	}
	var draftID string
	defer func() {
		if draftID != "" {
			db.ExecContext(context.Background(), `DELETE FROM questions WHERE id=$1`, draftID)
			db.ExecContext(context.Background(), `DELETE FROM question_candidates WHERE id=$1`, draftID)
		}
		db.ExecContext(context.Background(), `DELETE FROM ai_generation_requests WHERE user_id='ai-test-user' AND request_id=$1`, input.RequestID)
	}()
	calls := 0
	run := func() (domain.BankEntry, error) {
		calls++
		return generatedEntry([]byte(fixtureQuestion), input, "fixture-test", "Perkalian dasar", "numerasi")
	}
	response, err := db.GenerateAIDraft(ctx, "ai-test-user", input.RequestID, "hash-1", run)
	if err != nil {
		t.Fatal(err)
	}
	var result struct {
		Candidate domain.QuestionCandidateRecord `json:"candidate"`
	}
	if json.Unmarshal(response, &result) != nil {
		t.Fatal("invalid result")
	}
	draftID = result.Candidate.ID
	if result.Candidate.Status != "draft" {
		t.Fatal("not draft")
	}
	var active int
	if err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM questions WHERE id=$1`, draftID).Scan(&active); err != nil || active != 0 {
		t.Fatal("draft auto-published")
	}
	replay, err := db.GenerateAIDraft(ctx, "ai-test-user", input.RequestID, "hash-1", run)
	if err != nil || calls != 1 {
		t.Fatal("replay generated again")
	}
	var a, b any
	json.Unmarshal(response, &a)
	json.Unmarshal(replay, &b)
	if quoted(mustJSON(a)) != quoted(mustJSON(b)) {
		t.Fatal("replay changed")
	}
	if _, err = db.GenerateAIDraft(ctx, "ai-test-user", input.RequestID, "hash-2", run); !errors.Is(err, repository.ErrAIRequestConflict) {
		t.Fatal("conflicting input accepted")
	}
	bank := NewQuestionBankService(db)
	report, err := bank.Approve(ctx, draftID, "Fixture tes: 3 × 4 = 12 diperiksa manual; dibersihkan sesudah tes.")
	if err != nil || report.Items[0].Status != "approved" {
		t.Fatalf("approval failed %v", err)
	}
	if err = db.QueryRowContext(ctx, `SELECT COUNT(*) FROM questions WHERE id=$1`, draftID).Scan(&active); err != nil || active != 1 {
		t.Fatal("review failed activation")
	}
}
func mustJSON(value any) string { b, _ := json.Marshal(value); return string(b) }
