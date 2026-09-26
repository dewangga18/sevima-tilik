package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/sevima/tilik-api/internal/domain"
)

type UserRepository struct {
	db *DB
}

func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, name, role, grade_level, created_at
		FROM users
		WHERE email = $1
	`, email)

	var u domain.User
	var roleStr string
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &roleStr, &u.GradeLevel, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	u.Role = domain.Role(roleStr)
	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, email, password_hash, name, role, grade_level, created_at
		FROM users
		WHERE id = $1
	`, id)

	var u domain.User
	var roleStr string
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &roleStr, &u.GradeLevel, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	u.Role = domain.Role(roleStr)
	return &u, nil
}

func (r *UserRepository) CreateSession(ctx context.Context, session *domain.Session) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sessions (token, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
	`, session.Token, session.UserID, session.ExpiresAt, session.CreatedAt)
	return err
}

func (r *UserRepository) FindSession(ctx context.Context, token string) (*domain.Session, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT token, user_id, expires_at, created_at
		FROM sessions
		WHERE token = $1 AND expires_at > $2
	`, token, time.Now())

	var s domain.Session
	err := row.Scan(&s.Token, &s.UserID, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find session: %w", err)
	}
	return &s, nil
}

func (r *UserRepository) DeleteSession(ctx context.Context, token string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token = $1`, token)
	return err
}
