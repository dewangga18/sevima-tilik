package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/url"
	"strings"

	"github.com/sevima/tilik-api/internal/domain"
)

// Hash semantic JSON rather than whitespace, so formatting changes do not alter approval.
func CandidateHash(raw json.RawMessage) (string, error) {
	var value any
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := d.Decode(&value); err != nil {
		return "", err
	}
	canonical, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}

func ParseQuestionCandidates(body []byte, file string, reviews map[string]domain.CandidateReview) ([]domain.BankEntry, error) {
	if len(body) > 8<<20 {
		return nil, fmt.Errorf("%s: file exceeds 8 MiB", file)
	}
	var raws []json.RawMessage
	d := json.NewDecoder(bytes.NewReader(body))
	if err := d.Decode(&raws); err != nil {
		return nil, fmt.Errorf("%s: invalid JSON: %w", file, err)
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("%s: trailing JSON", file)
	}
	if len(raws) == 0 {
		return nil, fmt.Errorf("%s: empty bank", file)
	}
	entries := make([]domain.BankEntry, 0, len(raws))
	seen := map[string]bool{}
	for _, raw := range raws {
		var c domain.QuestionCandidate
		d := json.NewDecoder(bytes.NewReader(raw))
		d.DisallowUnknownFields()
		if err := d.Decode(&c); err != nil {
			return nil, fmt.Errorf("%s: invalid question fields: %w", file, err)
		}
		if c.ID == "" || seen[c.ID] {
			return nil, fmt.Errorf("%s: missing/duplicate ID %q", file, c.ID)
		}
		seen[c.ID] = true
		hash, err := CandidateHash(raw)
		if err != nil {
			return nil, err
		}
		e := domain.BankEntry{Candidate: c, Hash: hash, Payload: raw, SourceFile: file, Status: "draft", Note: "Belum direview untuk aktivasi."}
		if err := validateCandidate(c); err != nil {
			e.Status = "rejected"
			e.Note = err.Error()
		}
		if review, ok := reviews[c.ID]; ok {
			if review.Hash != hash {
				return nil, fmt.Errorf("%s: approval hash changed for %s; review content again", file, c.ID)
			}
			if review.Note == "" || (review.Status != "approved" && review.Status != "rejected" && review.Status != "draft") {
				return nil, fmt.Errorf("invalid review for %s", c.ID)
			}
			if e.Status == "rejected" && review.Status == "approved" {
				return nil, fmt.Errorf("%s: structural validation rejected approval: %s", c.ID, e.Note)
			}
			if e.Status != "rejected" {
				e.Status = review.Status
				e.Note = review.Note
			}
			if e.Status == "approved" && (review.ExpectedAnswer == "" || review.ExpectedAnswer != candidateAnswer(c)) {
				return nil, fmt.Errorf("%s: reviewed answer differs from key", c.ID)
			}
		}
		e.Question = domain.Question{ID: c.ID, SkillID: c.Skill, Difficulty: c.Difficulty, Prompt: c.Prompt, AnswerKey: candidateAnswer(c), Explanation: c.Explanation}
		for _, o := range c.Options {
			e.Question.Options = append(e.Question.Options, o.Text)
		}
		// Fixed rotation by content ID balances positions and stays stable across retry/reload.
		if e.Status == "approved" {
			sum := sha256.Sum256([]byte(c.ID))
			rotation := int(sum[0]) % len(c.Options)
			e.Question.Options = append(e.Question.Options[rotation:], e.Question.Options[:rotation]...)
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func candidateAnswer(c domain.QuestionCandidate) string {
	for _, o := range c.Options {
		if o.ID == c.CorrectAnswer {
			return o.Text
		}
	}
	return ""
}

func validateCandidate(c domain.QuestionCandidate) error {
	if strings.TrimSpace(c.Skill) == "" || strings.TrimSpace(c.Prompt) == "" || strings.TrimSpace(c.Explanation) == "" || c.Grade < 4 || c.Grade > 9 || c.Difficulty < 1 || c.Difficulty > 5 || c.QuestionType != "multiple_choice" {
		return fmt.Errorf("Skill/prompt/explanation/grade/difficulty/type tidak valid.")
	}
	if c.Purpose != "diagnostic" && c.Purpose != "practice" && c.Purpose != "reassessment" {
		return fmt.Errorf("Purpose tidak valid.")
	}
	if len(c.Options) != 4 {
		return fmt.Errorf("Harus ada empat opsi.")
	}
	ids, texts := map[string]bool{}, map[string]bool{}
	key := candidateAnswer(c)
	if key == "" {
		return fmt.Errorf("Key tidak menunjuk opsi yang tersedia.")
	}
	keyNumber, keyNumeric := new(big.Rat).SetString(key)
	for _, o := range c.Options {
		text := strings.TrimSpace(o.Text)
		if o.ID == "" || text == "" || text != o.Text || ids[o.ID] || texts[text] {
			return fmt.Errorf("ID/teks opsi kosong atau duplikat.")
		}
		ids[o.ID] = true
		texts[text] = true
		if number, numeric := new(big.Rat).SetString(text); keyNumeric && numeric && o.ID != c.CorrectAnswer && number.Cmp(keyNumber) == 0 {
			return fmt.Errorf("Dua opsi bernilai sama dengan jawaban benar; perlu review wording/opsi.")
		}
	}
	for _, m := range c.Misconceptions {
		if !ids[m.WrongAnswer] || m.WrongAnswer == c.CorrectAnswer || strings.TrimSpace(m.Reason) == "" {
			return fmt.Errorf("Mapping misconception tidak valid.")
		}
	}
	for _, p := range c.Prerequisites {
		if strings.TrimSpace(p) == "" {
			return fmt.Errorf("Prerequisite kosong.")
		}
	}
	u, err := url.Parse(c.SourceURL)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || c.SourceTitle == "" || c.SourceType == "" || c.LicenseNote == "" {
		return fmt.Errorf("Metadata sumber tidak lengkap.")
	}
	return nil
}
