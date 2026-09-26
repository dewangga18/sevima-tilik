package service

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/sevima/tilik-api/internal/repository"
)

var (
	ErrAIUnavailable = errors.New("AI is not configured")
	ErrAIInput       = errors.New("invalid AI input")
	ErrAIProvider    = errors.New("AI provider rejected request")
	ErrAIOutput      = errors.New("invalid AI output")
	ErrAIQuota       = errors.New("AI provider unavailable or quota exceeded")
	ErrAIRate        = errors.New("AI rate limit")
)
var modelPattern = regexp.MustCompile(`^gemini-[a-z0-9.-]{1,93}$`)

type AISettingsView struct {
	Configured   bool       `json:"configured"`
	StorageReady bool       `json:"storage_ready"`
	Model        string     `json:"model"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}
type AISettingsService struct {
	db         *repository.DB
	encryption cipher.AEAD
}

func NewAISettingsService(db *repository.DB, storageKey string) *AISettingsService {
	s := &AISettingsService{db: db}
	key, err := base64.StdEncoding.DecodeString(storageKey)
	if err == nil && len(key) == 32 {
		block, err := aes.NewCipher(key)
		if err == nil {
			s.encryption, _ = cipher.NewGCM(block)
		}
	}
	return s
}
func (s *AISettingsService) View(ctx context.Context) (AISettingsView, error) {
	record, err := s.db.AISettings(ctx)
	return AISettingsView{Configured: len(record.Ciphertext) > 0, StorageReady: s.encryption != nil, Model: record.Model, UpdatedAt: record.UpdatedAt}, err
}
func (s *AISettingsService) Save(ctx context.Context, key, model string) (AISettingsView, error) {
	if !modelPattern.MatchString(model) {
		return AISettingsView{}, ErrAIInput
	}
	if s.encryption == nil {
		return AISettingsView{}, ErrAIUnavailable
	}
	old, err := s.db.AISettings(ctx)
	if err != nil {
		return AISettingsView{}, err
	}
	encrypted := old.Ciphertext
	if key != "" {
		if len(key) < 16 || len(key) > 512 {
			return AISettingsView{}, ErrAIInput
		}
		for _, c := range key {
			if c < 33 || c > 126 {
				return AISettingsView{}, ErrAIInput
			}
		}
		encrypted, err = s.seal(key)
		if err != nil {
			return AISettingsView{}, err
		}
	}
	if len(encrypted) == 0 {
		return AISettingsView{}, ErrAIInput
	}
	if err = s.db.SaveAISettings(ctx, encrypted, model); err != nil {
		return AISettingsView{}, err
	}
	return s.View(ctx)
}
func (s *AISettingsService) Delete(ctx context.Context) (AISettingsView, error) {
	if err := s.db.DeleteAISettings(ctx); err != nil {
		return AISettingsView{}, err
	}
	return s.View(ctx)
}
func (s *AISettingsService) credentials(ctx context.Context) (string, string, error) {
	if s.encryption == nil {
		return "", "", ErrAIUnavailable
	}
	record, err := s.db.AISettings(ctx)
	if err != nil {
		return "", "", err
	}
	key, err := s.open(record.Ciphertext)
	if err != nil {
		return "", "", ErrAIUnavailable
	}
	return key, record.Model, nil
}
func (s *AISettingsService) seal(key string) ([]byte, error) {
	nonce := make([]byte, s.encryption.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return s.encryption.Seal(nonce, nonce, []byte(key), []byte("tilik:gemini-key:v1")), nil
}
func (s *AISettingsService) open(value []byte) (string, error) {
	if s.encryption == nil || len(value) < s.encryption.NonceSize()+s.encryption.Overhead() {
		return "", ErrAIUnavailable
	}
	n := s.encryption.NonceSize()
	plain, err := s.encryption.Open(nil, value[:n], value[n:], []byte("tilik:gemini-key:v1"))
	if err != nil || strings.TrimSpace(string(plain)) == "" {
		return "", ErrAIUnavailable
	}
	return string(plain), nil
}
