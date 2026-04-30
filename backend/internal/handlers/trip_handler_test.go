package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"ridex/backend/internal/middleware"
	"ridex/backend/internal/models"
	"ridex/backend/internal/store"
)

func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "ridex-test.db")
	st, err := store.New(dbPath)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	t.Cleanup(func() { _ = st.DB.Close() })
	return st
}

func TestRideRequestRespond_AcceptCreatesTrip(t *testing.T) {
	st := newTestStore(t)

	rider, err := st.CreateOAuthUser(context.Background(), "Rider", "rider@example.com", "google")
	if err != nil {
		t.Fatalf("create rider: %v", err)
	}
	driver, err := st.CreateOAuthUser(context.Background(), "Driver", "driver@example.com", "google")
	if err != nil {
		t.Fatalf("create driver: %v", err)
	}

	req, err := st.CreateRideRequest(context.Background(), models.RideRequest{
		UserID:            rider.ID,
		FromLabel:         "A",
		FromLat:           1,
		FromLon:           1,
		ToLabel:           "B",
		ToLat:             2,
		ToLon:             2,
		RideDate:          "2026-01-13",
		RideTime:          "10:00",
		Flexibility:       "exact",
		Passengers:        1,
		Luggage:           "none",
		MaxBudget:         20,
		RideType:          "shared",
		VehiclePreference: "any",
	})
	if err != nil {
		t.Fatalf("create ride request: %v", err)
	}

	h := &AuthHandler{store: st}

	payload := map[string]any{
		"action":  "accept",
		"message": "ok",
	}
	body, _ := json.Marshal(payload)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/ride-requests/"+req.ID+"/respond", bytes.NewReader(body))
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), middleware.UserIDContextKey, driver.ID))
	rec := httptest.NewRecorder()

	h.RideRequestRespond(rec, httpReq)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	var res struct {
		Message string       `json:"message"`
		Action  any          `json:"action"`
		Trip    *models.Trip `json:"trip"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if res.Trip == nil {
		t.Fatalf("expected trip in response")
	}
	if res.Trip.RideRequestID != req.ID {
		t.Fatalf("expected rideRequestId=%s got %s", req.ID, res.Trip.RideRequestID)
	}
	if res.Trip.RiderUserID != rider.ID || res.Trip.DriverUserID != driver.ID {
		t.Fatalf("unexpected rider/driver ids")
	}

	riderNotes, err := st.ListNotificationsByUser(context.Background(), rider.ID)
	if err != nil {
		t.Fatalf("list rider notifications: %v", err)
	}
	driverNotes, err := st.ListNotificationsByUser(context.Background(), driver.ID)
	if err != nil {
		t.Fatalf("list driver notifications: %v", err)
	}
	if len(riderNotes) == 0 || len(driverNotes) == 0 {
		t.Fatalf("expected notifications for both rider and driver")
	}
}

func TestTrips_ListTripsByUser(t *testing.T) {
	st := newTestStore(t)

	rider, _ := st.CreateOAuthUser(context.Background(), "Rider", "rider2@example.com", "google")
	driver, _ := st.CreateOAuthUser(context.Background(), "Driver", "driver2@example.com", "google")
	req, _ := st.CreateRideRequest(context.Background(), models.RideRequest{
		UserID:            rider.ID,
		FromLabel:         "A",
		FromLat:           1,
		FromLon:           1,
		ToLabel:           "B",
		ToLat:             2,
		ToLon:             2,
		RideDate:          "2026-01-13",
		RideTime:          "10:00",
		Flexibility:       "exact",
		Passengers:        1,
		Luggage:           "none",
		MaxBudget:         20,
		RideType:          "shared",
		VehiclePreference: "any",
	})
	_, err := st.CreateConfirmedTrip(context.Background(), req.ID, rider.ID, driver.ID, "")
	if err != nil {
		t.Fatalf("create trip: %v", err)
	}

	h := &AuthHandler{store: st}
	httpReq := httptest.NewRequest(http.MethodGet, "/api/trips", nil)
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), middleware.UserIDContextKey, rider.ID))
	rec := httptest.NewRecorder()

	h.Trips(rec, httpReq)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	var res struct {
		Trips []models.Trip `json:"trips"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(res.Trips) != 1 {
		t.Fatalf("expected 1 trip, got %d", len(res.Trips))
	}
	if res.Trips[0].RideRequestID != req.ID {
		t.Fatalf("unexpected trip rideRequestId")
	}
}

func TestPublishedRideRespond_AcceptCreatesTripAndNotifications(t *testing.T) {
	st := newTestStore(t)

	driver, err := st.CreateOAuthUser(context.Background(), "Driver", "driver3@example.com", "google")
	if err != nil {
		t.Fatalf("create driver: %v", err)
	}
	rider, err := st.CreateOAuthUser(context.Background(), "Rider", "rider3@example.com", "google")
	if err != nil {
		t.Fatalf("create rider: %v", err)
	}

	ride, err := st.CreatePublishedRide(context.Background(), models.PublishedRide{
		UserID:         driver.ID,
		FromLabel:      "Miami",
		FromLat:        25.7617,
		FromLon:        -80.1918,
		ToLabel:        "Orlando",
		ToLat:          28.5383,
		ToLon:          -81.3792,
		RideDate:       "2026-01-15",
		RideTime:       "07:30",
		Flexibility:    "exact",
		AvailableSeats: 3,
		TotalSeats:     3,
		PricePerSeat:   25,
		VehicleType:    "sedan",
		LuggageAllowed: "small",
		RideType:       "shared",
		RouteMiles:     234,
		RouteDuration:  "4h 24m",
	})
	if err != nil {
		t.Fatalf("create published ride: %v", err)
	}

	h := &AuthHandler{store: st}
	payload := map[string]any{
		"action":     "accept",
		"message":    "Ready to book",
		"passengers": 2,
	}
	body, _ := json.Marshal(payload)
	httpReq := httptest.NewRequest(http.MethodPost, "/api/published-rides/"+ride.ID+"/respond", bytes.NewReader(body))
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), middleware.UserIDContextKey, rider.ID))
	rec := httptest.NewRecorder()

	h.PublishedRideRespond(rec, httpReq)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	// Trip should exist for rider.
	trips, err := st.ListTripsByUser(context.Background(), rider.ID)
	if err != nil {
		t.Fatalf("list trips: %v", err)
	}
	if len(trips) != 1 {
		t.Fatalf("expected 1 trip, got %d", len(trips))
	}

	// Notifications should exist for both users.
	riderNotes, err := st.ListNotificationsByUser(context.Background(), rider.ID)
	if err != nil {
		t.Fatalf("list rider notifications: %v", err)
	}
	driverNotes, err := st.ListNotificationsByUser(context.Background(), driver.ID)
	if err != nil {
		t.Fatalf("list driver notifications: %v", err)
	}
	if len(riderNotes) == 0 || len(driverNotes) == 0 {
		t.Fatalf("expected notifications for both users")
	}
}

func TestCancelTrip_ReactivatesPublishedRide(t *testing.T) {
	st := newTestStore(t)

	driver, err := st.CreateOAuthUser(context.Background(), "Driver", "driver4@example.com", "google")
	if err != nil {
		t.Fatalf("create driver: %v", err)
	}
	rider, err := st.CreateOAuthUser(context.Background(), "Rider", "rider4@example.com", "google")
	if err != nil {
		t.Fatalf("create rider: %v", err)
	}

	ride, err := st.CreatePublishedRide(context.Background(), models.PublishedRide{
		UserID:         driver.ID,
		FromLabel:      "Miami",
		FromLat:        25.7617,
		FromLon:        -80.1918,
		ToLabel:        "Orlando",
		ToLat:          28.5383,
		ToLon:          -81.3792,
		RideDate:       "2026-01-15",
		RideTime:       "07:30",
		Flexibility:    "exact",
		AvailableSeats: 3,
		TotalSeats:     3,
		PricePerSeat:   25,
		VehicleType:    "sedan",
		LuggageAllowed: "small",
		RideType:       "shared",
		RouteMiles:     234,
		RouteDuration:  "4h 24m",
	})
	if err != nil {
		t.Fatalf("create published ride: %v", err)
	}

	// Accept published ride => trip confirmed and ride set inactive.
	h := &AuthHandler{store: st}
	payload := map[string]any{"action": "accept", "message": "Ready", "passengers": 1}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/published-rides/"+ride.ID+"/respond", bytes.NewReader(body))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDContextKey, rider.ID))
	rec := httptest.NewRecorder()
	h.PublishedRideRespond(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("accept status=%d body=%s", rec.Code, rec.Body.String())
	}

	// Find trip ID.
	trips, err := st.ListTripsByUser(context.Background(), rider.ID)
	if err != nil || len(trips) != 1 {
		t.Fatalf("expected 1 trip, got %d (err=%v)", len(trips), err)
	}
	tripID := trips[0].ID

	// Cancel trip should re-activate ride.
	cancelReq := httptest.NewRequest(http.MethodPost, "/api/trips/"+tripID+"/cancel", nil)
	cancelReq = cancelReq.WithContext(context.WithValue(cancelReq.Context(), middleware.UserIDContextKey, rider.ID))
	cancelRec := httptest.NewRecorder()
	h.CancelTrip(cancelRec, cancelReq)
	if cancelRec.Code != http.StatusOK {
		t.Fatalf("cancel status=%d body=%s", cancelRec.Code, cancelRec.Body.String())
	}

	updated, err := st.GetPublishedRideByID(context.Background(), ride.ID)
	if err != nil {
		t.Fatalf("get updated ride: %v", err)
	}
	if updated.Status != "active" {
		t.Fatalf("expected ride status active, got %s", updated.Status)
	}
}
