package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"ridex/backend/internal/middleware"
	"ridex/backend/internal/models"
)

type publishedRidePayload struct {
	FromLabel        string  `json:"fromLabel"`
	FromLat          float64 `json:"fromLat"`
	FromLon          float64 `json:"fromLon"`
	ToLabel          string  `json:"toLabel"`
	ToLat            float64 `json:"toLat"`
	ToLon            float64 `json:"toLon"`
	RideDate         string  `json:"rideDate"`
	RideTime         string  `json:"rideTime"`
	Flexibility      string  `json:"flexibility"`
	AvailableSeats   int     `json:"availableSeats"`
	TotalSeats       int     `json:"totalSeats"`
	PricePerSeat     float64 `json:"pricePerSeat"`
	VehicleType      string  `json:"vehicleType"`
	LuggageAllowed   string  `json:"luggageAllowed"`
	RideType         string  `json:"rideType"`
	VehicleInfo      string  `json:"vehicleInfo"`
	Notes            string  `json:"notes"`
	RouteMiles       int     `json:"routeMiles"`
	RouteDuration    string  `json:"routeDuration"`
	EarningsEstimate float64 `json:"earningsEstimate"`
}

type rideRequestActionPayload struct {
	Action          string `json:"action"`
	Message         string `json:"message"`
	PublishedRideID string `json:"publishedRideId"`
}

func (h *AuthHandler) PublishedRides(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value(middleware.UserIDContextKey).(string)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	switch r.Method {
	case http.MethodPost:
		h.createPublishedRide(w, r, userID)
	case http.MethodGet:
		rides, err := h.store.ListPublishedRidesByUser(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Could not get published rides")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"rides": rides})
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (h *AuthHandler) createPublishedRide(w http.ResponseWriter, r *http.Request, userID string) {
	var payload publishedRidePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	payload.FromLabel = strings.TrimSpace(payload.FromLabel)
	payload.ToLabel = strings.TrimSpace(payload.ToLabel)
	payload.Flexibility = strings.TrimSpace(payload.Flexibility)
	payload.VehicleType = strings.TrimSpace(payload.VehicleType)
	payload.LuggageAllowed = strings.TrimSpace(payload.LuggageAllowed)
	payload.RideType = strings.TrimSpace(payload.RideType)
	payload.VehicleInfo = strings.TrimSpace(payload.VehicleInfo)
	payload.Notes = strings.TrimSpace(payload.Notes)
	payload.RouteDuration = strings.TrimSpace(payload.RouteDuration)

	if payload.FromLabel == "" || payload.ToLabel == "" {
		writeError(w, http.StatusBadRequest, "Pickup and drop-off are required")
		return
	}
	if payload.RideDate == "" || payload.RideTime == "" {
		writeError(w, http.StatusBadRequest, "Date and time are required")
		return
	}
	if payload.AvailableSeats < 1 || payload.AvailableSeats > 8 || payload.TotalSeats < 1 || payload.TotalSeats > 8 {
		writeError(w, http.StatusBadRequest, "Seats must be between 1 and 8")
		return
	}
	if payload.AvailableSeats > payload.TotalSeats {
		writeError(w, http.StatusBadRequest, "Available seats cannot exceed total seats")
		return
	}
	if payload.PricePerSeat <= 0 {
		writeError(w, http.StatusBadRequest, "Price per seat must be greater than 0")
		return
	}
	if payload.FromLat == 0 && payload.FromLon == 0 {
		writeError(w, http.StatusBadRequest, "Pickup location must be selected from suggestions")
		return
	}
	if payload.ToLat == 0 && payload.ToLon == 0 {
		writeError(w, http.StatusBadRequest, "Drop-off location must be selected from suggestions")
		return
	}
	if len([]rune(payload.VehicleInfo)) > 80 {
		writeError(w, http.StatusBadRequest, "Vehicle info must be 80 characters or fewer")
		return
	}
	if len([]rune(payload.Notes)) > 220 {
		writeError(w, http.StatusBadRequest, "Notes should be under 220 characters")
		return
	}

	ride, err := h.store.CreatePublishedRide(r.Context(), models.PublishedRide{
		UserID:           userID,
		FromLabel:        payload.FromLabel,
		FromLat:          payload.FromLat,
		FromLon:          payload.FromLon,
		ToLabel:          payload.ToLabel,
		ToLat:            payload.ToLat,
		ToLon:            payload.ToLon,
		RideDate:         payload.RideDate,
		RideTime:         payload.RideTime,
		Flexibility:      payload.Flexibility,
		AvailableSeats:   payload.AvailableSeats,
		TotalSeats:       payload.TotalSeats,
		PricePerSeat:     payload.PricePerSeat,
		VehicleType:      payload.VehicleType,
		LuggageAllowed:   payload.LuggageAllowed,
		RideType:         payload.RideType,
		VehicleInfo:      payload.VehicleInfo,
		Notes:            payload.Notes,
		RouteMiles:       payload.RouteMiles,
		RouteDuration:    payload.RouteDuration,
		EarningsEstimate: payload.EarningsEstimate,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not publish ride")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "Ride published successfully.",
		"ride":    ride,
	})
}

func (h *AuthHandler) RideRequestFeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDContextKey).(string)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	passengerCount, _ := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get("passengers")))
	maxBudget, _ := strconv.ParseFloat(strings.TrimSpace(r.URL.Query().Get("budget")), 64)
	feed, err := h.store.ListRideRequestFeed(
		r.Context(),
		userID,
		r.URL.Query().Get("date"),
		r.URL.Query().Get("route"),
		passengerCount,
		maxBudget,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not get ride request feed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"requests": feed})
}

func (h *AuthHandler) RideRequestRespond(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDContextKey).(string)
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	requestID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/ride-requests/"), "/respond")
	requestID = strings.Trim(requestID, "/")
	if requestID == "" {
		writeError(w, http.StatusBadRequest, "Ride request ID is required")
		return
	}

	var payload rideRequestActionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	payload.Action = strings.TrimSpace(strings.ToLower(payload.Action))
	payload.Message = strings.TrimSpace(payload.Message)
	payload.PublishedRideID = strings.TrimSpace(payload.PublishedRideID)

	if payload.Action != "accept" && payload.Action != "negotiate" {
		writeError(w, http.StatusBadRequest, "Action must be accept or negotiate")
		return
	}
	if len([]rune(payload.Message)) > 220 {
		writeError(w, http.StatusBadRequest, "Message should be under 220 characters")
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
	if request.UserID == userID {
		writeError(w, http.StatusBadRequest, "You cannot respond to your own ride request")
		return
	}

	if payload.PublishedRideID != "" {
		ride, err := h.store.GetPublishedRideByID(r.Context(), payload.PublishedRideID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusNotFound, "Published ride not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "Could not get published ride")
			return
		}
		if ride.UserID != userID {
			writeError(w, http.StatusUnauthorized, "You do not have access to that ride")
			return
		}
	}

	action, err := h.store.CreateRideRequestAction(r.Context(), models.RideRequestAction{
		RideRequestID:   requestID,
		DriverUserID:    userID,
		PublishedRideID: payload.PublishedRideID,
		Action:          payload.Action,
		Message:         payload.Message,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not save response")
		return
	}

	message := "Request marked for follow-up."
	if payload.Action == "accept" {
		message = "Ride request accepted."
	} else if payload.Action == "negotiate" {
		message = "Negotiation started."
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message": message,
		"action":  action,
	})
}
