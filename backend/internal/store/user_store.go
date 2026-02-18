package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"ridex/backend/internal/models"
)

var ErrEmailExists = errors.New("email already exists")

func (s *Store) CreateUser(ctx context.Context, name, email, passwordHash string) (*models.User, error) {
	now := time.Now().UTC()
	user := &models.User{
		ID:           uuid.NewString(),
		Name:         name,
		Email:        strings.ToLower(strings.TrimSpace(email)),
		PasswordHash: passwordHash,
		Rating:       5.0,
		CreatedAt:    now,
	}

	query := `INSERT INTO users (id, name, email, password_hash, rating, created_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := s.DB.ExecContext(ctx, query, user.ID, user.Name, user.Email, user.PasswordHash, user.Rating, user.CreatedAt)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, ErrEmailExists
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, rating, created_at FROM users WHERE email = ?`
	row := s.DB.QueryRowContext(ctx, query, strings.ToLower(strings.TrimSpace(email)))

	var u models.User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Rating, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("select user by email: %w", err)
	}
	return &u, nil
}

func (s *Store) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, rating, created_at FROM users WHERE id = ?`
	row := s.DB.QueryRowContext(ctx, query, id)

	var u models.User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Rating, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("select user by id: %w", err)
	}
	return &u, nil
}
