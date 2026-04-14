package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"ridex/backend/internal/middleware"
)

func (h *AuthHandler) Trips(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/trips/") {
		h.TripByID(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDContextKey).(string)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	trips, err := h.store.ListTripViewsByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not get trips")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"trips": trips})
}

func (h *AuthHandler) TripByID(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/cancel") {
		h.CancelTrip(w, r)
		return
	}
	writeError(w, http.StatusNotFound, "Not found")
}

func (h *AuthHandler) CancelTrip(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDContextKey).(string)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	tripID := strings.TrimPrefix(r.URL.Path, "/api/trips/")
	tripID = strings.TrimSuffix(tripID, "/cancel")
	tripID = strings.Trim(tripID, "/")
	if tripID == "" {
		writeError(w, http.StatusBadRequest, "Trip ID is required")
		return
	}

	trip, err := h.store.CancelTrip(r.Context(), tripID, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "Trip not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not cancel trip")
		return
	}

	// Re-activate the published ride so it reappears for everyone.
	if strings.TrimSpace(trip.PublishedRideID) != "" {
		_ = h.store.SetPublishedRideStatus(r.Context(), trip.PublishedRideID, trip.DriverUserID, "active")
	}

	_, _ = h.store.CreateNotification(r.Context(), trip.RiderUserID, "Trip cancelled", "A confirmed trip was cancelled and is available again.")
	_, _ = h.store.CreateNotification(r.Context(), trip.DriverUserID, "Trip cancelled", "A confirmed trip was cancelled and your ride is available again.")

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Trip cancelled.",
		"trip":    trip,
	})
}

