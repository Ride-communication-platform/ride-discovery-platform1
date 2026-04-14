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

func (s *Store) CreateConfirmedTrip(ctx context.Context, rideRequestID, riderUserID, driverUserID, publishedRideID string) (*models.Trip, error) {
	trip := &models.Trip{
		ID:              uuid.NewString(),
		RideRequestID:   strings.TrimSpace(rideRequestID),
		RiderUserID:     strings.TrimSpace(riderUserID),
		DriverUserID:    strings.TrimSpace(driverUserID),
		PublishedRideID: strings.TrimSpace(publishedRideID),
		Status:          "confirmed",
		CreatedAt:       time.Now().UTC(),
	}

	query := `INSERT INTO trips (
		id, ride_request_id, rider_user_id, driver_user_id, published_ride_id, status, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.DB.ExecContext(
		ctx,
		query,
		trip.ID, trip.RideRequestID, trip.RiderUserID, trip.DriverUserID, trip.PublishedRideID, trip.Status, trip.CreatedAt,
	)
	if err != nil {
		// If a trip already exists for this ride request, return the existing trip.
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return s.GetTripByRideRequestID(ctx, trip.RideRequestID)
		}
		return nil, fmt.Errorf("insert trip: %w", err)
	}

	return trip, nil
}

func (s *Store) GetTripByRideRequestID(ctx context.Context, rideRequestID string) (*models.Trip, error) {
	query := `SELECT id, ride_request_id, rider_user_id, driver_user_id, published_ride_id, status, created_at
		FROM trips
		WHERE ride_request_id = ?`

	var trip models.Trip
	err := s.DB.QueryRowContext(ctx, query, strings.TrimSpace(rideRequestID)).Scan(
		&trip.ID,
		&trip.RideRequestID,
		&trip.RiderUserID,
		&trip.DriverUserID,
		&trip.PublishedRideID,
		&trip.Status,
		&trip.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("select trip by ride request id: %w", err)
	}
	return &trip, nil
}

func (s *Store) ListTripsByUser(ctx context.Context, userID string) ([]models.Trip, error) {
	query := `SELECT id, ride_request_id, rider_user_id, driver_user_id, published_ride_id, status, created_at
		FROM trips
		WHERE rider_user_id = ? OR driver_user_id = ?
		ORDER BY created_at DESC`

	rows, err := s.DB.QueryContext(ctx, query, strings.TrimSpace(userID), strings.TrimSpace(userID))
	if err != nil {
		return nil, fmt.Errorf("list trips: %w", err)
	}
	defer rows.Close()

	trips := []models.Trip{}
	for rows.Next() {
		var trip models.Trip
		if err := rows.Scan(
			&trip.ID,
			&trip.RideRequestID,
			&trip.RiderUserID,
			&trip.DriverUserID,
			&trip.PublishedRideID,
			&trip.Status,
			&trip.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan trip: %w", err)
		}
		trips = append(trips, trip)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate trips: %w", err)
	}

	return trips, nil
}

func (s *Store) ListTripViewsByUser(ctx context.Context, userID string) ([]models.TripView, error) {
	query := `SELECT
		t.id, t.status, t.created_at, t.ride_request_id, t.published_ride_id,
		t.rider_user_id, ru.name, t.driver_user_id, du.name,
		rr.from_label, rr.to_label, rr.ride_date, rr.ride_time, rr.passengers, rr.route_miles, rr.route_duration,
		COALESCE(pr.price_per_seat, 0)
	FROM trips t
	JOIN ride_requests rr ON rr.id = t.ride_request_id
	JOIN users ru ON ru.id = t.rider_user_id
	JOIN users du ON du.id = t.driver_user_id
	LEFT JOIN published_rides pr ON pr.id = t.published_ride_id
	WHERE t.rider_user_id = ? OR t.driver_user_id = ?
	ORDER BY t.created_at DESC`

	rows, err := s.DB.QueryContext(ctx, query, strings.TrimSpace(userID), strings.TrimSpace(userID))
	if err != nil {
		return nil, fmt.Errorf("list trip views: %w", err)
	}
	defer rows.Close()

	out := []models.TripView{}
	for rows.Next() {
		var v models.TripView
		if err := rows.Scan(
			&v.ID, &v.Status, &v.CreatedAt, &v.RideRequestID, &v.PublishedRideID,
			&v.RiderUserID, &v.RiderName, &v.DriverUserID, &v.DriverName,
			&v.FromLabel, &v.ToLabel, &v.RideDate, &v.RideTime, &v.Passengers, &v.RouteMiles, &v.RouteDuration,
			&v.PricePerSeat,
		); err != nil {
			return nil, fmt.Errorf("scan trip view: %w", err)
		}
		if v.Passengers <= 0 {
			v.Passengers = 1
		}
		v.EstimatedTotal = v.PricePerSeat * float64(v.Passengers)
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate trip views: %w", err)
	}
	return out, nil
}

func (s *Store) CancelTrip(ctx context.Context, tripID, userID string) (*models.Trip, error) {
	result, err := s.DB.ExecContext(
		ctx,
		`UPDATE trips SET status = 'cancelled' 
		 WHERE id = ? AND (rider_user_id = ? OR driver_user_id = ?) AND status = 'confirmed'`,
		strings.TrimSpace(tripID),
		strings.TrimSpace(userID),
		strings.TrimSpace(userID),
	)
	if err != nil {
		return nil, fmt.Errorf("cancel trip: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("cancel trip affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

	var trip models.Trip
	err = s.DB.QueryRowContext(ctx, `SELECT id, ride_request_id, rider_user_id, driver_user_id, published_ride_id, status, created_at FROM trips WHERE id = ?`, strings.TrimSpace(tripID)).Scan(
		&trip.ID,
		&trip.RideRequestID,
		&trip.RiderUserID,
		&trip.DriverUserID,
		&trip.PublishedRideID,
		&trip.Status,
		&trip.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("select cancelled trip: %w", err)
	}
	return &trip, nil
}

