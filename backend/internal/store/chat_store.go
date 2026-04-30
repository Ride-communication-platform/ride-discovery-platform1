package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"ridex/backend/internal/models"
)

func (s *Store) CreateOrGetConversation(ctx context.Context, riderUserID, driverUserID, rideRequestID, publishedRideID, tripID string) (*models.ChatConversation, error) {
	riderUserID = strings.TrimSpace(riderUserID)
	driverUserID = strings.TrimSpace(driverUserID)
	rideRequestID = strings.TrimSpace(rideRequestID)
	publishedRideID = strings.TrimSpace(publishedRideID)
	tripID = strings.TrimSpace(tripID)

	if rideRequestID != "" {
		conversation, err := s.GetConversationByRequestAndParticipants(ctx, rideRequestID, riderUserID, driverUserID)
		if err == nil {
			if tripID != "" || publishedRideID != "" {
				if err := s.updateConversationLinks(ctx, conversation.ID, publishedRideID, tripID); err != nil {
					return nil, err
				}
				return s.GetConversationByIDForUser(ctx, conversation.ID, riderUserID)
			}
			return conversation, nil
		}
		if err != sql.ErrNoRows {
			return nil, err
		}
	}

	now := time.Now().UTC()
	conversation := &models.ChatConversation{
		ID:              uuid.NewString(),
		RiderUserID:     riderUserID,
		DriverUserID:    driverUserID,
		RideRequestID:   rideRequestID,
		PublishedRideID: publishedRideID,
		TripID:          tripID,
		Status:          "active",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	_, err := s.DB.ExecContext(
		ctx,
		`INSERT INTO chat_conversations (
			id, rider_user_id, driver_user_id, ride_request_id, published_ride_id, trip_id, status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		conversation.ID,
		conversation.RiderUserID,
		conversation.DriverUserID,
		conversation.RideRequestID,
		conversation.PublishedRideID,
		conversation.TripID,
		conversation.Status,
		conversation.CreatedAt,
		conversation.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert chat conversation: %w", err)
	}

	return conversation, nil
}

func (s *Store) GetConversationByRequestAndParticipants(ctx context.Context, rideRequestID, riderUserID, driverUserID string) (*models.ChatConversation, error) {
	var conversation models.ChatConversation
	err := s.DB.QueryRowContext(
		ctx,
		`SELECT id, rider_user_id, driver_user_id, ride_request_id, published_ride_id, trip_id, status, created_at, updated_at
		 FROM chat_conversations
		 WHERE ride_request_id = ? AND rider_user_id = ? AND driver_user_id = ?
		 LIMIT 1`,
		strings.TrimSpace(rideRequestID),
		strings.TrimSpace(riderUserID),
		strings.TrimSpace(driverUserID),
	).Scan(
		&conversation.ID,
		&conversation.RiderUserID,
		&conversation.DriverUserID,
		&conversation.RideRequestID,
		&conversation.PublishedRideID,
		&conversation.TripID,
		&conversation.Status,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("select chat conversation by request: %w", err)
	}
	return &conversation, nil
}

func (s *Store) GetConversationByIDForUser(ctx context.Context, conversationID, userID string) (*models.ChatConversation, error) {
	var conversation models.ChatConversation
	err := s.DB.QueryRowContext(
		ctx,
		`SELECT id, rider_user_id, driver_user_id, ride_request_id, published_ride_id, trip_id, status, created_at, updated_at
		 FROM chat_conversations
		 WHERE id = ? AND (rider_user_id = ? OR driver_user_id = ?)`,
		strings.TrimSpace(conversationID),
		strings.TrimSpace(userID),
		strings.TrimSpace(userID),
	).Scan(
		&conversation.ID,
		&conversation.RiderUserID,
		&conversation.DriverUserID,
		&conversation.RideRequestID,
		&conversation.PublishedRideID,
		&conversation.TripID,
		&conversation.Status,
		&conversation.CreatedAt,
		&conversation.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("select chat conversation: %w", err)
	}
	return &conversation, nil
}

func (s *Store) updateConversationLinks(ctx context.Context, conversationID, publishedRideID, tripID string) error {
	_, err := s.DB.ExecContext(
		ctx,
		`UPDATE chat_conversations
		 SET published_ride_id = CASE WHEN ? != '' THEN ? ELSE published_ride_id END,
		     trip_id = CASE WHEN ? != '' THEN ? ELSE trip_id END,
		     updated_at = ?
		 WHERE id = ?`,
		publishedRideID,
		publishedRideID,
		tripID,
		tripID,
		time.Now().UTC(),
		strings.TrimSpace(conversationID),
	)
	if err != nil {
		return fmt.Errorf("update chat conversation links: %w", err)
	}
	return nil
}

func (s *Store) ListChatSummariesByUser(ctx context.Context, userID string) ([]models.ChatConversationSummary, error) {
	rows, err := s.DB.QueryContext(
		ctx,
		`SELECT
			c.id, c.rider_user_id, c.driver_user_id, c.ride_request_id, c.published_ride_id, c.trip_id, c.status, c.created_at, c.updated_at,
			CASE WHEN c.rider_user_id = ? THEN c.driver_user_id ELSE c.rider_user_id END AS other_user_id,
			CASE WHEN c.rider_user_id = ? THEN du.name ELSE ru.name END AS other_user_name,
			CASE WHEN c.rider_user_id = ? THEN du.avatar_data ELSE ru.avatar_data END AS other_user_avatar_data,
			COALESCE(rr.from_label, ''), COALESCE(rr.to_label, ''), COALESCE(rr.ride_date, ''), COALESCE(rr.ride_time, ''),
			COALESCE(t.status, ''),
			COALESCE(CASE
				WHEN last_message.message_type = 'image' THEN '[Image]'
				WHEN last_message.message_type = 'location' THEN COALESCE(last_message.location_label, '[Location]')
				ELSE last_message.body
			END, ''),
			COALESCE(last_message.sender_user_id, ''), last_message.created_at
		FROM chat_conversations c
		JOIN users ru ON ru.id = c.rider_user_id
		JOIN users du ON du.id = c.driver_user_id
		LEFT JOIN ride_requests rr ON rr.id = c.ride_request_id
		LEFT JOIN trips t ON t.id = c.trip_id
		LEFT JOIN chat_messages last_message
			ON last_message.id = (
				SELECT cm.id
				FROM chat_messages cm
				WHERE cm.conversation_id = c.id
				ORDER BY cm.created_at DESC
				LIMIT 1
			)
		WHERE c.rider_user_id = ? OR c.driver_user_id = ?
		ORDER BY COALESCE(last_message.created_at, c.updated_at) DESC`,
		strings.TrimSpace(userID),
		strings.TrimSpace(userID),
		strings.TrimSpace(userID),
		strings.TrimSpace(userID),
		strings.TrimSpace(userID),
	)
	if err != nil {
		return nil, fmt.Errorf("list chat conversations: %w", err)
	}
	defer rows.Close()

	out := []models.ChatConversationSummary{}
	for rows.Next() {
		var item models.ChatConversationSummary
		var lastMessageAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.RiderUserID,
			&item.DriverUserID,
			&item.RideRequestID,
			&item.PublishedRideID,
			&item.TripID,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.OtherUserID,
			&item.OtherUserName,
			&item.OtherUserAvatarData,
			&item.FromLabel,
			&item.ToLabel,
			&item.RideDate,
			&item.RideTime,
			&item.TripStatus,
			&item.LastMessage,
			&item.LastMessageSenderID,
			&lastMessageAt,
		); err != nil {
			return nil, fmt.Errorf("scan chat conversation summary: %w", err)
		}
		if lastMessageAt.Valid {
			item.LastMessageAt = lastMessageAt.Time
		} else {
			item.LastMessageAt = item.UpdatedAt
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat conversations: %w", err)
	}
	return out, nil
}

func (s *Store) ListChatMessagesByConversation(ctx context.Context, conversationID, userID string) ([]models.ChatMessage, error) {
	rows, err := s.DB.QueryContext(
		ctx,
		`SELECT cm.id, cm.conversation_id, cm.sender_user_id, cm.body, cm.message_type, cm.image_data, cm.location_label, cm.location_lat, cm.location_lon, cm.created_at
		 FROM chat_messages cm
		 JOIN chat_conversations c ON c.id = cm.conversation_id
		 WHERE cm.conversation_id = ? AND (c.rider_user_id = ? OR c.driver_user_id = ?)
		 ORDER BY cm.created_at ASC`,
		strings.TrimSpace(conversationID),
		strings.TrimSpace(userID),
		strings.TrimSpace(userID),
	)
	if err != nil {
		return nil, fmt.Errorf("list chat messages: %w", err)
	}
	defer rows.Close()

	out := []models.ChatMessage{}
	for rows.Next() {
		var message models.ChatMessage
		if err := rows.Scan(&message.ID, &message.ConversationID, &message.SenderUserID, &message.Body, &message.MessageType, &message.ImageData, &message.LocationLabel, &message.LocationLat, &message.LocationLon, &message.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan chat message: %w", err)
		}
		out = append(out, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat messages: %w", err)
	}
	return out, nil
}

func (s *Store) CreateChatMessage(ctx context.Context, input models.ChatMessage) (*models.ChatMessage, error) {
	conversationID := strings.TrimSpace(input.ConversationID)
	senderUserID := strings.TrimSpace(input.SenderUserID)
	conversation, err := s.GetConversationByIDForUser(ctx, conversationID, senderUserID)
	if err != nil {
		return nil, err
	}

	message := &models.ChatMessage{
		ID:             uuid.NewString(),
		ConversationID: conversation.ID,
		SenderUserID:   senderUserID,
		Body:           strings.TrimSpace(input.Body),
		MessageType:    strings.TrimSpace(input.MessageType),
		ImageData:      strings.TrimSpace(input.ImageData),
		LocationLabel:  strings.TrimSpace(input.LocationLabel),
		LocationLat:    input.LocationLat,
		LocationLon:    input.LocationLon,
		CreatedAt:      time.Now().UTC(),
	}
	if message.MessageType == "" {
		message.MessageType = "text"
	}

	_, err = s.DB.ExecContext(
		ctx,
		`INSERT INTO chat_messages (
			id, conversation_id, sender_user_id, body, message_type, image_data, location_label, location_lat, location_lon, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		message.ID,
		message.ConversationID,
		message.SenderUserID,
		message.Body,
		message.MessageType,
		message.ImageData,
		message.LocationLabel,
		message.LocationLat,
		message.LocationLon,
		message.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert chat message: %w", err)
	}

	_, err = s.DB.ExecContext(
		ctx,
		`UPDATE chat_conversations SET updated_at = ? WHERE id = ?`,
		message.CreatedAt,
		message.ConversationID,
	)
	if err != nil {
		return nil, fmt.Errorf("touch chat conversation: %w", err)
	}

	return message, nil
}
