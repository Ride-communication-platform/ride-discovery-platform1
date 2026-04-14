package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"ridex/backend/internal/models"
)

func (s *Store) CreateNotification(ctx context.Context, userID, title, body string) (*models.Notification, error) {
	n := &models.Notification{
		ID:        uuid.NewString(),
		UserID:    strings.TrimSpace(userID),
		Title:     strings.TrimSpace(title),
		Body:      strings.TrimSpace(body),
		Read:      false,
		CreatedAt: time.Now().UTC(),
	}

	_, err := s.DB.ExecContext(
		ctx,
		`INSERT INTO notifications (id, user_id, title, body, read, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		n.ID, n.UserID, n.Title, n.Body, 0, n.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert notification: %w", err)
	}
	return n, nil
}

func (s *Store) ListNotificationsByUser(ctx context.Context, userID string) ([]models.Notification, error) {
	rows, err := s.DB.QueryContext(
		ctx,
		`SELECT id, user_id, title, body, read, created_at
		 FROM notifications
		 WHERE user_id = ?
		 ORDER BY created_at DESC`,
		strings.TrimSpace(userID),
	)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	out := []models.Notification{}
	for rows.Next() {
		var n models.Notification
		var readInt int
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &readInt, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		n.Read = readInt != 0
		out = append(out, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}
	return out, nil
}

