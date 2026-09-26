package service

import (
	"context"
	"errors"
	"testing"

	"github.com/sevima/tilik-api/internal/domain"
	"github.com/sevima/tilik-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

func TestDemoLoginEnvironmentBoundary(t *testing.T) {
	db := diagnosticTestDB(t)
	ctx := context.Background()
	repo := repository.NewUserRepository(db)
	development := NewAuthService(repo, true)
	production := NewAuthService(repo, false)
	legacyHash, err := bcrypt.GenerateFromPassword([]byte("legacy-demo-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	for _, role := range []domain.Role{domain.RoleStudent, domain.RoleTeacher, domain.RoleAdmin} {
		user, session, err := development.DemoLogin(ctx, role)
		if err != nil || user.Role != role || session.Token == "" {
			t.Fatalf("demo login %s failed: %v", role, err)
		}
		validated, err := development.ValidateSession(ctx, session.Token)
		if err != nil || validated.ID != user.ID {
			t.Fatalf("demo session invalid: %v", err)
		}
		if _, _, err := production.DemoLogin(ctx, role); !errors.Is(err, ErrDemoDisabled) {
			t.Fatalf("production permits demo login: %v", err)
		}
		if _, err := production.ValidateSession(ctx, session.Token); !errors.Is(err, ErrUnauthorized) {
			t.Fatalf("production accepts a development demo session: %v", err)
		}
		if _, err := db.ExecContext(ctx, "UPDATE users SET password_hash = $1 WHERE id = $2", string(legacyHash), user.ID); err != nil {
			t.Fatal(err)
		}
		if _, _, err := production.Login(ctx, user.Email, "legacy-demo-password"); !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("production permits legacy demo password login: %v", err)
		}
	}
	if _, err := db.ExecContext(ctx, `INSERT INTO users (id, email, password_hash, name, role) VALUES ($1, $2, $3, $4, $5)`, "ordinary-user", "ordinary@example.test", string(legacyHash), "Ordinary", "student"); err != nil {
		t.Fatal(err)
	}
	user, session, err := production.Login(ctx, "ordinary@example.test", "legacy-demo-password")
	if err != nil || user.ID != "ordinary-user" {
		t.Fatalf("production login for ordinary user failed: %v", err)
	}
	if _, err := production.ValidateSession(ctx, session.Token); err != nil {
		t.Fatalf("production session for ordinary user failed: %v", err)
	}
	if _, _, err := development.DemoLogin(ctx, domain.Role("superuser")); !errors.Is(err, ErrInvalidDemoRole) {
		t.Fatalf("invalid role accepted: %v", err)
	}
}

func TestCurriculumSeedDoesNotCreateDemoUsers(t *testing.T) {
	db := testDB(t, false)
	var count int
	if err := db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM users").Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("curriculum seed created demo accounts")
	}
}
