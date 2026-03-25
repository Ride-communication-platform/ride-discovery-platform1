package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"ridex/backend/internal/middleware"
	"ridex/backend/internal/models"
)

type rideRequestPayload struct {
	FromLabel           string  `json:"fromLabel"`
	FromLat             float64 `json:"fromLat"`
	FromLon             float64 `json:"fromLon"`
	ToLabel             string  `json:"toLabel"`
	ToLat               float64 `json:"toLat"`
	ToLon               float64 `json:"toLon"`
	RideDate            string  `json:"rideDate"`
	RideTime            string  `json:"rideTime"`
	Flexibility         string  `json:"flexibility"`
	Passengers          int     `json:"passengers"`
	Luggage             string  `json:"luggage"`
	MaxBudget           float64 `json:"maxBudget"`
	RideType            string  `json:"rideType"`
	VehiclePreference   string  `json:"vehiclePreference"`
	MinimumRating       float64 `json:"minimumRating"`
	VerifiedDriversOnly bool    `json:"verifiedDriversOnly"`
	Notes               string  `json:"notes"`
	RouteMiles          int     `json:"routeMiles"`
	RouteDuration       string  `json:"routeDuration"`
	PriceEstimate       string  `json:"priceEstimate"`
}

func (h *AuthHandler) RideRequests(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.UserIDContextKey).(string)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	switch r.Method {
	case http.MethodPost:
		h.createRideRequest(w, r, userID)
	case http.MethodGet:
		h.listMyRideRequests(w, r, userID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *AuthHandler) createRideRequest(w http.ResponseWriter, r *http.Request, userID string) {
	var payload rideRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	payload.FromLabel = strings.TrimSpace(payload.FromLabel)
	payload.ToLabel = strings.TrimSpace(payload.ToLabel)
	payload.Flexibility = strings.TrimSpace(payload.Flexibility)
	payload.Luggage = strings.TrimSpace(payload.Luggage)
	payload.RideType = strings.TrimSpace(payload.RideType)
	payload.VehiclePreference = strings.TrimSpace(payload.VehiclePreference)
	payload.Notes = strings.TrimSpace(payload.Notes)
	payload.RouteDuration = strings.TrimSpace(payload.RouteDuration)
	payload.PriceEstimate = strings.TrimSpace(payload.PriceEstimate)

	if payload.FromLabel == "" || payload.ToLabel == "" {
		writeError(w, http.StatusBadRequest, "Pickup and destination are required")
		return
	}
	if payload.RideDate == "" || payload.RideTime == "" {
		writeError(w, http.StatusBadRequest, "Date and time are required")
		return
	}
	if payload.Passengers < 1 || payload.Passengers > 6 {
		writeError(w, http.StatusBadRequest, "Passengers must be between 1 and 6")
		return
	}
	if payload.FromLat == 0 && payload.FromLon == 0 {
		writeError(w, http.StatusBadRequest, "Pickup location must be selected from suggestions")
		return
	}
	if payload.ToLat == 0 && payload.ToLon == 0 {
		writeError(w, http.StatusBadRequest, "Destination must be selected from suggestions")
		return
	}
	if strings.EqualFold(payload.FromLabel, payload.ToLabel) {
		writeError(w, http.StatusBadRequest, "Destination must be different from pickup")
		return
	}
	if len([]rune(payload.Notes)) > 220 {
		writeError(w, http.StatusBadRequest, "Notes should be under 220 characters")
		return
	}
	if payload.MaxBudget < 0 {
		writeError(w, http.StatusBadRequest, "Budget must be greater than or equal to 0")
		return
	}

	request, err := h.store.CreateRideRequest(r.Context(), models.RideRequest{
		UserID:              userID,
		FromLabel:           payload.FromLabel,
		FromLat:             payload.FromLat,
		FromLon:             payload.FromLon,
		ToLabel:             payload.ToLabel,
		ToLat:               payload.ToLat,
		ToLon:               payload.ToLon,
		RideDate:            payload.RideDate,
		RideTime:            payload.RideTime,
		Flexibility:         payload.Flexibility,
		Passengers:          payload.Passengers,
		Luggage:             payload.Luggage,
		MaxBudget:           payload.MaxBudget,
		RideType:            payload.RideType,
		VehiclePreference:   payload.VehiclePreference,
		MinimumRating:       payload.MinimumRating,
		VerifiedDriversOnly: payload.VerifiedDriversOnly,
		Notes:               payload.Notes,
		RouteMiles:          payload.RouteMiles,
		RouteDuration:       payload.RouteDuration,
		PriceEstimate:       payload.PriceEstimate,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not save ride request")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "Ride request posted successfully.",
		"request": request,
	})
}

func (h *AuthHandler) listMyRideRequests(w http.ResponseWriter, r *http.Request, userID string) {
	requests, err := h.store.ListRideRequestsByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not get ride requests")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"requests": requests})
}

func (h *AuthHandler) RideRequestByID(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.UserIDContextKey).(string)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	if strings.HasSuffix(r.URL.Path, "/respond") {
		h.RideRequestRespond(w, r)
		return
	}

	requestID := strings.TrimPrefix(r.URL.Path, "/api/ride-requests/")
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		writeError(w, http.StatusBadRequest, "Ride request ID is required")
		return
	}

	if r.Method == http.MethodDelete {
		h.deleteRideRequest(w, r, userID, requestID)
		return
	}
	if r.Method == http.MethodPut {
		h.updateRideRequest(w, r, userID, requestID)
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	request, err := h.store.GetRideRequestByID(r.Context(), requestID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Ride request not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not get ride request")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"request": request})
}

func (h *AuthHandler) deleteRideRequest(w http.ResponseWriter, r *http.Request, userID, requestID string) {
	request, err := h.store.GetRideRequestByID(r.Context(), requestID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Ride request not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not get ride request")
		return
	}
	if request.UserID != userID {
		writeError(w, http.StatusUnauthorized, "You do not have access to this ride request")
		return
	}

	if err := h.store.DeleteRideRequest(r.Context(), requestID, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Ride request not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not cancel ride request")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Ride request cancelled.",
	})
}

func (h *AuthHandler) updateRideRequest(w http.ResponseWriter, r *http.Request, userID, requestID string) {
	request, err := h.store.GetRideRequestByID(r.Context(), requestID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Ride request not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not get ride request")
		return
	}
	if request.UserID != userID {
		writeError(w, http.StatusUnauthorized, "You do not have access to this ride request")
		return
	}

	var payload rideRequestPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	payload.FromLabel = strings.TrimSpace(payload.FromLabel)
	payload.ToLabel = strings.TrimSpace(payload.ToLabel)
	payload.Flexibility = strings.TrimSpace(payload.Flexibility)
	payload.Luggage = strings.TrimSpace(payload.Luggage)
	payload.RideType = strings.TrimSpace(payload.RideType)
	payload.VehiclePreference = strings.TrimSpace(payload.VehiclePreference)
	payload.Notes = strings.TrimSpace(payload.Notes)
	payload.RouteDuration = strings.TrimSpace(payload.RouteDuration)
	payload.PriceEstimate = strings.TrimSpace(payload.PriceEstimate)

	if payload.FromLabel == "" || payload.ToLabel == "" {
		writeError(w, http.StatusBadRequest, "Pickup and destination are required")
		return
	}
	if payload.RideDate == "" || payload.RideTime == "" {
		writeError(w, http.StatusBadRequest, "Date and time are required")
		return
	}
	if payload.Passengers < 1 || payload.Passengers > 6 {
		writeError(w, http.StatusBadRequest, "Passengers must be between 1 and 6")
		return
	}
	if payload.FromLat == 0 && payload.FromLon == 0 {
		writeError(w, http.StatusBadRequest, "Pickup location must be selected from suggestions")
		return
	}
	if payload.ToLat == 0 && payload.ToLon == 0 {
		writeError(w, http.StatusBadRequest, "Destination must be selected from suggestions")
		return
	}
	if strings.EqualFold(payload.FromLabel, payload.ToLabel) {
		writeError(w, http.StatusBadRequest, "Destination must be different from pickup")
		return
	}
	if len([]rune(payload.Notes)) > 220 {
		writeError(w, http.StatusBadRequest, "Notes should be under 220 characters")
		return
	}
	if payload.MaxBudget < 0 {
		writeError(w, http.StatusBadRequest, "Budget must be greater than or equal to 0")
		return
	}

	updatedRequest, err := h.store.UpdateRideRequest(r.Context(), models.RideRequest{
		ID:                  requestID,
		UserID:              userID,
		FromLabel:           payload.FromLabel,
		FromLat:             payload.FromLat,
		FromLon:             payload.FromLon,
		ToLabel:             payload.ToLabel,
		ToLat:               payload.ToLat,
		ToLon:               payload.ToLon,
		RideDate:            payload.RideDate,
		RideTime:            payload.RideTime,
		Flexibility:         payload.Flexibility,
		Passengers:          payload.Passengers,
		Luggage:             payload.Luggage,
		MaxBudget:           payload.MaxBudget,
		RideType:            payload.RideType,
		VehiclePreference:   payload.VehiclePreference,
		MinimumRating:       payload.MinimumRating,
		VerifiedDriversOnly: payload.VerifiedDriversOnly,
		Notes:               payload.Notes,
		RouteMiles:          payload.RouteMiles,
		RouteDuration:       payload.RouteDuration,
		PriceEstimate:       payload.PriceEstimate,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Ride request not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not update ride request")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Ride request updated successfully.",
		"request": updatedRequest,
	})
}
