package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/sevima/tilik-api/internal/domain"
)

var ErrAIRequestConflict = errors.New("AI request conflict")

type AISettingsRecord struct {
	Ciphertext []byte
	Model      string
	UpdatedAt  *time.Time
}

func (db *DB) AISettings(ctx context.Context) (AISettingsRecord, error) {
	var s AISettingsRecord
	err := db.QueryRowContext(ctx, `SELECT ciphertext,model,updated_at FROM ai_settings WHERE id=1`).Scan(&s.Ciphertext, &s.Model, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return AISettingsRecord{Model: "gemini-3.8-flash"}, nil
	}
	return s, err
}
func (db *DB) SaveAISettings(ctx context.Context, ciphertext []byte, model string) error {
	_, err := db.ExecContext(ctx, `INSERT INTO ai_settings(id,ciphertext,model) VALUES(1,$1,$2) ON CONFLICT(id) DO UPDATE SET ciphertext=EXCLUDED.ciphertext,model=EXCLUDED.model,updated_at=NOW()`, ciphertext, model)
	return err
}
func (db *DB) DeleteAISettings(ctx context.Context) error {
	_, err := db.ExecContext(ctx, `DELETE FROM ai_settings WHERE id=1`)
	return err
}

// One transaction keeps the successful provider result, candidate, and retry
// ledger together. Failed provider calls are rolled back and never replayed.
func (db *DB) GenerateAIDraft(ctx context.Context, userID, requestID, inputHash string, generate func() (domain.BankEntry, error)) (json.RawMessage, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var locked bool
	if err = tx.QueryRowContext(ctx, `SELECT pg_try_advisory_xact_lock(hashtextextended($1,0))`, "ai-generator:"+userID+":"+requestID).Scan(&locked); err != nil {
		return nil, err
	}
	if !locked {
		return nil, ErrAIRequestConflict
	}
	var previousHash string
	var response []byte
	err = tx.QueryRowContext(ctx, `SELECT input_hash,response FROM ai_generation_requests WHERE user_id=$1 AND request_id=$2`, userID, requestID).Scan(&previousHash, &response)
	if err == nil {
		if previousHash != inputHash {
			return nil, ErrAIRequestConflict
		}
		return json.RawMessage(response), nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	entry, err := generate()
	if err != nil {
		return nil, err
	}
	if _, err = registerQuestionBankTx(ctx, tx, []domain.BankEntry{entry}, true); err != nil {
		return nil, err
	}
	record, err := scanCandidate(tx.QueryRowContext(ctx, `SELECT `+candidateColumns+` FROM question_candidates WHERE id=$1`, entry.Candidate.ID).Scan)
	if err != nil {
		return nil, err
	}
	response, err = json.Marshal(map[string]any{"candidate": record})
	if err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO ai_generation_requests(user_id,request_id,input_hash,response) VALUES($1,$2,$3,$4)`, userID, requestID, inputHash, response); err != nil {
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return json.RawMessage(response), nil
}
