package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrDemoDisabled       = errors.New("demo login disabled")
	ErrInvalidDemoRole    = errors.New("invalid demo role")
)

type AuthService struct {
	userRepo    *repository.UserRepository
	demoEnabled bool
}

func NewAuthService(userRepo *repository.UserRepository, demoEnabled bool) *AuthService {
	return &AuthService{userRepo: userRepo, demoEnabled: demoEnabled}
}

var demoUserIDs = map[domain.Role]string{
	domain.RoleStudent: "u-student-new",
	domain.RoleTeacher: "u-teacher-1",
	domain.RoleAdmin:   "u-admin-1",
}

func isDemoUser(user *domain.User) bool {
	if user.ID == "u-student-1" {
		return true
	}
	for _, id := range demoUserIDs {
		if user.ID == id {
			return true
		}
	}
	return false
}

func (s *AuthService) DemoLogin(ctx context.Context, role domain.Role) (*domain.User, *domain.Session, error) {
	if !s.demoEnabled {
		return nil, nil, ErrDemoDisabled
	}
	id, ok := demoUserIDs[role]
	if !ok {
		return nil, nil, ErrInvalidDemoRole
	}
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if user == nil || user.Role != role {
		return nil, nil, ErrUnauthorized
	}
	return s.createSession(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*domain.User, *domain.Session, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, nil, err
	}
	if user == nil || (!s.demoEnabled && isDemoUser(user)) {
		return nil, nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}
	return s.createSession(ctx, user)
}

func (s *AuthService) createSession(ctx context.Context, user *domain.User) (*domain.User, *domain.Session, error) {
	// Generate secure session token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, nil, fmt.Errorf("generate session token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	session := &domain.Session{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(24 * 7 * time.Hour), // 7 days
		CreatedAt: time.Now(),
	}

	if err := s.userRepo.CreateSession(ctx, session); err != nil {
		return nil, nil, fmt.Errorf("create session: %w", err)
	}

	return user, session, nil
}

func (s *AuthService) ValidateSession(ctx context.Context, token string) (*domain.User, error) {
	if token == "" {
		return nil, ErrUnauthorized
	}

	session, err := s.userRepo.FindSession(ctx, token)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return nil, ErrUnauthorized
	}

	user, err := s.userRepo.FindByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil || (!s.demoEnabled && isDemoUser(user)) {
		return nil, ErrUnauthorized
	}

	return user, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.userRepo.DeleteSession(ctx, token)
}
