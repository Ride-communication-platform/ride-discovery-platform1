package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"ridex/backend/internal/middleware"
	"ridex/backend/internal/models"
)

type sendChatMessagePayload struct {
	Body          string  `json:"body"`
	MessageType   string  `json:"messageType"`
	ImageData     string  `json:"imageData"`
	LocationLabel string  `json:"locationLabel"`
	LocationLat   float64 `json:"locationLat"`
	LocationLon   float64 `json:"locationLon"`
}

func (h *AuthHandler) Chats(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/chats/") {
		h.ChatByID(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	conversations, err := h.store.ListChatSummariesByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not get chats")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"conversations": conversations})
}

func (h *AuthHandler) ChatByID(w http.ResponseWriter, r *http.Request) {
	if !strings.HasSuffix(r.URL.Path, "/messages") {
		writeError(w, http.StatusNotFound, "Not found")
		return
	}

	userID, ok := currentUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	conversationID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/chats/"), "/messages")
	conversationID = strings.TrimSpace(strings.Trim(conversationID, "/"))
	if conversationID == "" {
		writeError(w, http.StatusBadRequest, "Chat ID is required")
		return
	}

	switch r.Method {
	case http.MethodGet:
		messages, err := h.store.ListChatMessagesByConversation(r.Context(), conversationID, userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Could not get messages")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"messages": messages})
	case http.MethodPost:
		var payload sendChatMessagePayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, "Invalid JSON payload")
			return
		}

		payload.Body = strings.TrimSpace(payload.Body)
		payload.MessageType = strings.TrimSpace(strings.ToLower(payload.MessageType))
		payload.ImageData = strings.TrimSpace(payload.ImageData)
		payload.LocationLabel = strings.TrimSpace(payload.LocationLabel)
		if payload.MessageType == "" {
			payload.MessageType = "text"
		}

		switch payload.MessageType {
		case "text":
			if payload.Body == "" {
				writeError(w, http.StatusBadRequest, "Message body is required")
				return
			}
			if len([]rune(payload.Body)) > 1000 {
				writeError(w, http.StatusBadRequest, "Message must be 1000 characters or fewer")
				return
			}
		case "image":
			if payload.ImageData == "" {
				writeError(w, http.StatusBadRequest, "Image data is required")
				return
			}
			if !strings.HasPrefix(payload.ImageData, "data:image/") {
				writeError(w, http.StatusBadRequest, "Image must use a data URL")
				return
			}
			if len(payload.ImageData) > 1_500_000 {
				writeError(w, http.StatusBadRequest, "Image is too large")
				return
			}
			if payload.Body != "" && len([]rune(payload.Body)) > 280 {
				writeError(w, http.StatusBadRequest, "Image caption must be 280 characters or fewer")
				return
			}
		case "location":
			if payload.LocationLabel == "" {
				writeError(w, http.StatusBadRequest, "Location label is required")
				return
			}
			if payload.LocationLat == 0 && payload.LocationLon == 0 {
				writeError(w, http.StatusBadRequest, "Location coordinates are required")
				return
			}
			if len([]rune(payload.LocationLabel)) > 160 {
				writeError(w, http.StatusBadRequest, "Location label must be 160 characters or fewer")
				return
			}
			if payload.Body != "" && len([]rune(payload.Body)) > 280 {
				writeError(w, http.StatusBadRequest, "Location note must be 280 characters or fewer")
				return
			}
		default:
			writeError(w, http.StatusBadRequest, "Unsupported message type")
			return
		}

		conversation, err := h.store.GetConversationByIDForUser(r.Context(), conversationID, userID)
		if err != nil {
			if err == sql.ErrNoRows {
				writeError(w, http.StatusNotFound, "Chat not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "Could not send message")
			return
		}

		message, err := h.store.CreateChatMessage(r.Context(), models.ChatMessage{
			ConversationID: conversationID,
			SenderUserID:   userID,
			Body:           payload.Body,
			MessageType:    payload.MessageType,
			ImageData:      payload.ImageData,
			LocationLabel:  payload.LocationLabel,
			LocationLat:    payload.LocationLat,
			LocationLon:    payload.LocationLon,
		})
		if err != nil {
			if err == sql.ErrNoRows {
				writeError(w, http.StatusNotFound, "Chat not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "Could not send message")
			return
		}

		recipientUserID := conversation.RiderUserID
		if recipientUserID == userID {
			recipientUserID = conversation.DriverUserID
		}
		notificationBody := payload.Body
		if notificationBody == "" {
			if payload.MessageType == "image" {
				notificationBody = "Sent an image"
			} else if payload.MessageType == "location" {
				notificationBody = payload.LocationLabel
			}
		}
		_, _ = h.store.CreateNotification(r.Context(), recipientUserID, "New chat message", notificationBody)

		writeJSON(w, http.StatusCreated, map[string]any{
			"message":     "Message sent.",
			"chatMessage": message,
		})
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func currentUserID(r *http.Request) (string, bool) {
	userID, _ := r.Context().Value(middleware.UserIDContextKey).(string)
	return userID, userID != ""
}
