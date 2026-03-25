package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"ridex/backend/internal/models"
)

var ErrEmailExists = errors.New("email already exists")

func encodeInterests(interests []string) (string, error) {
	if interests == nil {
		interests = []string{}
	}
	data, err := json.Marshal(interests)
	if err != nil {
		return "", fmt.Errorf("marshal interests: %w", err)
	}
	return string(data), nil
}

func decodeInterests(raw string) ([]string, error) {
	if strings.TrimSpace(raw) == "" {
		return []string{}, nil
	}
	var interests []string
	if err := json.Unmarshal([]byte(raw), &interests); err != nil {
		return nil, fmt.Errorf("unmarshal interests: %w", err)
	}
	return interests, nil
}

func (s *Store) CreateUser(ctx context.Context, name, email, passwordHash, verificationCode string) (*models.User, error) {
	now := time.Now().UTC()
	user := &models.User{
		ID:                uuid.NewString(),
		Name:              name,
		Email:             strings.ToLower(strings.TrimSpace(email)),
		AvatarData:        "",
		Interests:         []string{},
		PasswordHash:      passwordHash,
		Rating:            0,
		RatingCount:       0,
		TripsCompleted:    0,
		EmailVerified:     false,
		VerificationCode:  verificationCode,
		PasswordResetCode: "",
		AuthProvider:      "password",
		CreatedAt:         now,
	}

	interestsJSON, err := encodeInterests(user.Interests)
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO users (id, name, email, avatar_data, interests, password_hash, rating, rating_count, trips_completed, email_verified, verification_code, password_reset_code, password_reset_sent_at, auth_provider, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = s.DB.ExecContext(ctx, query, user.ID, user.Name, user.Email, user.AvatarData, interestsJSON, user.PasswordHash, user.Rating, user.RatingCount, user.TripsCompleted, user.EmailVerified, user.VerificationCode, user.PasswordResetCode, nil, user.AuthProvider, user.CreatedAt)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, ErrEmailExists
		}
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return user, nil
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, name, email, avatar_data, interests, password_hash, rating, rating_count, trips_completed, email_verified, verification_code, password_reset_code, password_reset_sent_at, auth_provider, created_at FROM users WHERE email = ?`
	row := s.DB.QueryRowContext(ctx, query, strings.ToLower(strings.TrimSpace(email)))

	var u models.User
	var interestsRaw string
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.AvatarData, &interestsRaw, &u.PasswordHash, &u.Rating, &u.RatingCount, &u.TripsCompleted, &u.EmailVerified, &u.VerificationCode, &u.PasswordResetCode, &u.PasswordResetSentAt, &u.AuthProvider, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("select user by email: %w", err)
	}
	decodedInterests, err := decodeInterests(interestsRaw)
	if err != nil {
		return nil, err
	}
	u.Interests = decodedInterests
	return &u, nil
}

func (s *Store) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `SELECT id, name, email, avatar_data, interests, password_hash, rating, rating_count, trips_completed, email_verified, verification_code, password_reset_code, password_reset_sent_at, auth_provider, created_at FROM users WHERE id = ?`
	row := s.DB.QueryRowContext(ctx, query, id)

	var u models.User
	var interestsRaw string
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.AvatarData, &interestsRaw, &u.PasswordHash, &u.Rating, &u.RatingCount, &u.TripsCompleted, &u.EmailVerified, &u.VerificationCode, &u.PasswordResetCode, &u.PasswordResetSentAt, &u.AuthProvider, &u.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("select user by id: %w", err)
	}
	decodedInterests, err := decodeInterests(interestsRaw)
	if err != nil {
		return nil, err
	}
	u.Interests = decodedInterests
	return &u, nil
}

func (s *Store) UpdateUserProfile(ctx context.Context, id, name, avatarData string, interests []string) (*models.User, error) {
	interestsJSON, err := encodeInterests(interests)
	if err != nil {
		return nil, err
	}

	query := `UPDATE users SET name = ?, avatar_data = ?, interests = ? WHERE id = ?`
	result, err := s.DB.ExecContext(ctx, query, strings.TrimSpace(name), avatarData, interestsJSON, id)
	if err != nil {
		return nil, fmt.Errorf("update user profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("update user profile affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

	return s.GetUserByID(ctx, id)
}

func (s *Store) CreateOAuthUser(ctx context.Context, name, email, provider string) (*models.User, error) {
	now := time.Now().UTC()
	user := &models.User{
		ID:                uuid.NewString(),
		Name:              strings.TrimSpace(name),
		Email:             strings.ToLower(strings.TrimSpace(email)),
		AvatarData:        "",
		Interests:         []string{},
		PasswordHash:      "",
		Rating:            0,
		RatingCount:       0,
		TripsCompleted:    0,
		EmailVerified:     true,
		VerificationCode:  "",
		PasswordResetCode: "",
		AuthProvider:      provider,
		CreatedAt:         now,
	}

	interestsJSON, err := encodeInterests(user.Interests)
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO users (id, name, email, avatar_data, interests, password_hash, rating, rating_count, trips_completed, email_verified, verification_code, password_reset_code, password_reset_sent_at, auth_provider, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = s.DB.ExecContext(ctx, query, user.ID, user.Name, user.Email, user.AvatarData, interestsJSON, user.PasswordHash, user.Rating, user.RatingCount, user.TripsCompleted, user.EmailVerified, user.VerificationCode, user.PasswordResetCode, nil, user.AuthProvider, user.CreatedAt)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, ErrEmailExists
		}
		return nil, fmt.Errorf("insert oauth user: %w", err)
	}

	return user, nil
}

func (s *Store) VerifyUserByEmail(ctx context.Context, email string) error {
	query := `UPDATE users SET email_verified = 1, verification_code = '' WHERE email = ?`
	result, err := s.DB.ExecContext(ctx, query, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return fmt.Errorf("verify user by email: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("verify user by email affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) UpdateVerificationCode(ctx context.Context, email, verificationCode string) error {
	query := `UPDATE users SET verification_code = ? WHERE email = ?`
	result, err := s.DB.ExecContext(ctx, query, verificationCode, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return fmt.Errorf("update verification code: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("verification code affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *Store) VerifyUserEmail(ctx context.Context, email, verificationCode string) error {
	query := `UPDATE users SET email_verified = 1, verification_code = '' WHERE email = ? AND verification_code = ?`
	result, err := s.DB.ExecContext(ctx, query, strings.ToLower(strings.TrimSpace(email)), strings.TrimSpace(verificationCode))
	if err != nil {
		return fmt.Errorf("verify user email: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("verify user rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *Store) SetPasswordResetCode(ctx context.Context, email, resetCode string, resetSentAt time.Time) error {
	query := `UPDATE users SET password_reset_code = ?, password_reset_sent_at = ? WHERE email = ?`
	result, err := s.DB.ExecContext(ctx, query, resetCode, resetSentAt, strings.ToLower(strings.TrimSpace(email)))
	if err != nil {
		return fmt.Errorf("set password reset code: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("password reset affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *Store) ResetPassword(ctx context.Context, email, resetCode, passwordHash string, now time.Time, expiryWindow time.Duration) error {
	query := `UPDATE users
		SET password_hash = ?, password_reset_code = '', password_reset_sent_at = NULL
		WHERE email = ? AND password_reset_code = ? AND password_reset_sent_at IS NOT NULL AND password_reset_sent_at >= ?`
	result, err := s.DB.ExecContext(ctx, query, passwordHash, strings.ToLower(strings.TrimSpace(email)), strings.TrimSpace(resetCode), now.Add(-expiryWindow))
	if err != nil {
		return fmt.Errorf("reset password: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("reset password affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
