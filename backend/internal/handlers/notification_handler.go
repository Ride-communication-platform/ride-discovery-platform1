package handlers

import (
	"net/http"

	"ridex/backend/internal/middleware"
)

func (h *AuthHandler) Notifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDContextKey).(string)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	notifications, err := h.store.ListNotificationsByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not get notifications")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"notifications": notifications})
}

