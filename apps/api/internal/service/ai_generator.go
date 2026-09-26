package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
)

type AIGenerateInput struct {
	RequestID  string `json:"request_id"`
	SkillID    string `json:"skill_id"`
	GradeLevel int    `json:"grade_level"`
	Difficulty int    `json:"difficulty"`
	Purpose    string `json:"purpose"`
}
type generatedQuestion struct {
	Question      string                   `json:"question"`
	Options       []domain.CandidateOption `json:"options"`
	CorrectAnswer string                   `json:"correct_answer"`
	Explanation   string                   `json:"explanation"`
}
type aiWindow struct {
	start time.Time
	count int
}
type AIGeneratorService struct {
	db       *repository.DB
	settings *AISettingsService
	gemini   *GeminiClient
	slots    chan struct{}
	mu       sync.Mutex
	windows  map[string]aiWindow
}

func NewAIGeneratorService(db *repository.DB, settings *AISettingsService) *AIGeneratorService {
	return &AIGeneratorService{db: db, settings: settings, gemini: NewGeminiClient(), slots: make(chan struct{}, 2), windows: map[string]aiWindow{}}
}

var requestIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{16,80}$`)

func (s *AIGeneratorService) Generate(ctx context.Context, userID string, input AIGenerateInput) (json.RawMessage, error) {
	if !requestIDPattern.MatchString(input.RequestID) || len(input.SkillID) > 100 || input.GradeLevel != 4 || (input.Difficulty != 1 && input.Difficulty != 2) || (input.Purpose != "diagnostic" && input.Purpose != "practice" && input.Purpose != "reassessment") {
		return nil, ErrAIInput
	}
	ready, err := s.db.SkillReadyForActivation(ctx, input.SkillID)
	if err != nil {
		return nil, err
	}
	if !ready {
		return nil, ErrAIInput
	}
	body, _ := json.Marshal(input)
	hash := sha256.Sum256(body)
	return s.db.GenerateAIDraft(ctx, userID, input.RequestID, hex.EncodeToString(hash[:]), func() (domain.BankEntry, error) {
		select {
		case s.slots <- struct{}{}:
			defer func() { <-s.slots }()
		default:
			return domain.BankEntry{}, ErrAIRate
		}
		if !s.allow(userID) {
			return domain.BankEntry{}, ErrAIRate
		}
		key, model, err := s.settings.credentials(ctx)
		if err != nil {
			return domain.BankEntry{}, err
		}
		var name, description, skillDomain string
		if err = s.db.QueryRowContext(ctx, `SELECT name,description,domain FROM skills WHERE id=$1`, input.SkillID).Scan(&name, &description, &skillDomain); err != nil {
			return domain.BankEntry{}, err
		}
		spec, _ := json.Marshal(map[string]any{"skill_id": input.SkillID, "skill_name": name, "description": description, "grade": 4, "difficulty": input.Difficulty, "purpose": input.Purpose, "instructions": "Empat opsi A,B,C,D yang berbeda nilainya; satu benar. Explanation menjelaskan hitungan. Tidak ada data siswa. Jangan sertakan URL/sumber."})
		providerCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
		defer cancel()
		raw, err := s.gemini.Generate(providerCtx, key, model, string(spec), questionSchema())
		if err != nil {
			return domain.BankEntry{}, err
		}
		return generatedEntry(raw, input, model, name, skillDomain)
	})
}
func (s *AIGeneratorService) allow(userID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for id, w := range s.windows {
		if now.Sub(w.start) >= time.Minute {
			delete(s.windows, id)
		}
	}
	w := s.windows[userID]
	if w.start.IsZero() {
		w.start = now
	}
	if w.count >= 6 {
		return false
	}
	w.count++
	s.windows[userID] = w
	return true
}
func questionSchema() map[string]any {
	str := map[string]any{"type": "string"}
	return map[string]any{"type": "object", "properties": map[string]any{
		"question": str, "explanation": str, "correct_answer": map[string]any{"type": "string", "enum": []string{"A", "B", "C", "D"}},
		"options": map[string]any{"type": "array", "minItems": 4, "maxItems": 4, "items": map[string]any{"type": "object", "properties": map[string]any{"id": map[string]any{"type": "string", "enum": []string{"A", "B", "C", "D"}}, "text": str}, "required": []string{"id", "text"}, "additionalProperties": false}},
	}, "required": []string{"question", "explanation", "correct_answer", "options"}, "additionalProperties": false}
}
func generatedEntry(raw []byte, input AIGenerateInput, model, name, skillDomain string) (domain.BankEntry, error) {
	var q generatedQuestion
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&q) != nil {
		return domain.BankEntry{}, ErrAIOutput
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return domain.BankEntry{}, ErrAIOutput
	}
	if strings.TrimSpace(q.Question) == "" || len(q.Question) > 2000 || strings.TrimSpace(q.Explanation) == "" || len(q.Explanation) > 4000 || len(q.Options) != 4 {
		return domain.BankEntry{}, ErrAIOutput
	}
	numbers := []*big.Rat{}
	for _, o := range q.Options {
		if !strings.Contains("ABCD", o.ID) || len(o.ID) != 1 || len(o.Text) > 500 {
			return domain.BankEntry{}, ErrAIOutput
		}
		if n, ok := new(big.Rat).SetString(strings.ReplaceAll(o.Text, ",", ".")); ok {
			for _, prev := range numbers {
				if n.Cmp(prev) == 0 {
					return domain.BankEntry{}, ErrAIOutput
				}
			}
			numbers = append(numbers, n)
		}
	}
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return domain.BankEntry{}, err
	}
	c := domain.QuestionCandidate{ID: "ai-" + hex.EncodeToString(random), Purpose: input.Purpose, Grade: 4, Phase: "B", Domain: skillDomain, Topic: name, Skill: input.SkillID, Difficulty: input.Difficulty, QuestionType: "multiple_choice", CognitiveLevel: "generated_unreviewed", Prompt: q.Question, Options: q.Options, CorrectAnswer: q.CorrectAnswer, Explanation: q.Explanation, Prerequisites: []string{}, Misconceptions: []domain.CandidateMisconception{}, MetadataInferred: true, SourceURL: "https://ai.google.dev/", SourceTitle: "Draft Gemini / " + model, SourceType: "ai_generated", LicenseNote: "Konten hasil generasi; wajib review manusia sebelum digunakan. URL menandai provider, bukan sumber soal terverifikasi.", RawSourceReference: "Gemini " + model}
	if err := validateCandidate(c); err != nil {
		return domain.BankEntry{}, ErrAIOutput
	}
	payload, err := json.Marshal(c)
	if err != nil {
		return domain.BankEntry{}, err
	}
	hash, err := CandidateHash(payload)
	if err != nil {
		return domain.BankEntry{}, err
	}
	return domain.BankEntry{Candidate: c, Hash: hash, SourceFile: "gemini:" + model, Payload: payload, Status: "draft", Note: fmt.Sprintf("Draft AI (%s). Verifikasi kunci, penjelasan, dan kesesuaian skill sebelum menyetujui.", model)}, nil
}
