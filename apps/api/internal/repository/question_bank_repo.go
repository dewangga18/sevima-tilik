package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/sevima/tilik-api/internal/domain"
)

func (db *DB) RegisterQuestionBank(ctx context.Context, entries []domain.BankEntry, apply bool) (*domain.BankImportReport, error) {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: !apply})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended('question-bank-import',0))`); err != nil {
		return nil, err
	}
	report := &domain.BankImportReport{Items: make([]domain.BankImportItem, 0, len(entries))}
	seen := map[string]bool{}
	plannedPrompts := map[string]string{}
	for _, e := range entries {
		c := e.Candidate
		if seen[c.ID] {
			return nil, fmt.Errorf("duplicate ID across files: %s", c.ID)
		}
		seen[c.ID] = true
		var oldHash, oldStatus string
		err = tx.QueryRowContext(ctx, `SELECT content_hash,status FROM question_candidates WHERE id=$1`, c.ID).Scan(&oldHash, &oldStatus)
		exists := err == nil
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		if exists && oldHash != e.Hash {
			return nil, fmt.Errorf("content conflict for %s; keep registered content immutable and use a new ID", c.ID)
		}
		if exists && oldStatus == "approved" && e.Status != "approved" {
			return nil, fmt.Errorf("%s is already approved; import cannot revoke historical content", c.ID)
		}
		item := domain.BankImportItem{ID: c.ID, Hash: e.Hash, Status: e.Status, Reason: e.Note, Action: "register"}
		if exists && oldStatus == e.Status {
			item.Action = "unchanged"
		}
		if e.Status == "approved" {
			if c.Grade != 4 || c.Difficulty > 2 {
				return nil, fmt.Errorf("%s: activation limited to grade 4, levels 1–2", c.ID)
			}
			var ready bool
			if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM skills s JOIN lessons l ON l.skill_id=s.id WHERE s.id=$1 AND s.grade_level<=4)`, c.Skill).Scan(&ready); err != nil {
				return nil, err
			}
			if !ready {
				return nil, fmt.Errorf("%s: skill/lesson not ready for activation", c.ID)
			}
			var runtimeID string
			err = tx.QueryRowContext(ctx, `SELECT id FROM questions WHERE id=$1`, c.ID).Scan(&runtimeID)
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}
			if err == nil && (!exists || oldStatus != "approved") {
				return nil, fmt.Errorf("runtime question ID conflict: %s", c.ID)
			}
			if errors.Is(err, sql.ErrNoRows) {
				// Compare prompts across purposes so fresh reassessment cannot repeat practice verbatim.
				var duplicate string
				err = tx.QueryRowContext(ctx, `SELECT id FROM questions WHERE lower(regexp_replace(trim(prompt),'\s+',' ','g'))=$1 ORDER BY id LIMIT 1`, strings.ToLower(strings.Join(strings.Fields(c.Prompt), " "))).Scan(&duplicate)
				if err != nil && !errors.Is(err, sql.ErrNoRows) {
					return nil, err
				}
				promptKey := strings.ToLower(strings.Join(strings.Fields(c.Prompt), " "))
				if err != nil && plannedPrompts[promptKey] != "" {
					duplicate = plannedPrompts[promptKey]
					err = nil
				}
				if err == nil {
					item.Status = "draft"
					item.Reason = "Prompt duplikat bank aktif: " + duplicate
					e.Status = "draft"
					e.Note = item.Reason
				} else {
					options, err := json.Marshal(e.Question.Options)
					if err != nil {
						return nil, err
					}
					if apply {
						if _, err = tx.ExecContext(ctx, `INSERT INTO questions(id,skill_id,difficulty,prompt,options,answer_key,explanation,purpose) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, c.ID, c.Skill, c.Difficulty, c.Prompt, options, e.Question.AnswerKey, c.Explanation, c.Purpose); err != nil {
							return nil, err
						}
					}
					plannedPrompts[promptKey] = c.ID
					item.Action = "activate"
				}
			}
		}
		if exists && oldStatus == e.Status && item.Action != "activate" {
			item.Action = "unchanged"
		}
		if apply && item.Action != "unchanged" {
			if _, err = tx.ExecContext(ctx, `INSERT INTO question_candidates(id,content_hash,source_file,payload,status,review_note) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(id) DO UPDATE SET status=EXCLUDED.status,review_note=EXCLUDED.review_note`, c.ID, e.Hash, e.SourceFile, e.Payload, e.Status, e.Note); err != nil {
				return nil, err
			}
		}
		report.Items = append(report.Items, item)
	}
	if apply {
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		report.Applied = true
	}
	return report, nil
}
