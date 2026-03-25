package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"ridex/backend/internal/models"
)

func (s *Store) CreateRideRequest(ctx context.Context, input models.RideRequest) (*models.RideRequest, error) {
	request := &models.RideRequest{
		ID:                  uuid.NewString(),
		UserID:              input.UserID,
		FromLabel:           input.FromLabel,
		FromLat:             input.FromLat,
		FromLon:             input.FromLon,
		ToLabel:             input.ToLabel,
		ToLat:               input.ToLat,
		ToLon:               input.ToLon,
		RideDate:            input.RideDate,
		RideTime:            input.RideTime,
		Flexibility:         input.Flexibility,
		Passengers:          input.Passengers,
		Luggage:             input.Luggage,
		MaxBudget:           input.MaxBudget,
		RideType:            input.RideType,
		VehiclePreference:   input.VehiclePreference,
		MinimumRating:       input.MinimumRating,
		VerifiedDriversOnly: input.VerifiedDriversOnly,
		Notes:               input.Notes,
		RouteMiles:          input.RouteMiles,
		RouteDuration:       input.RouteDuration,
		PriceEstimate:       input.PriceEstimate,
		CreatedAt:           time.Now().UTC(),
	}

	query := `INSERT INTO ride_requests (
		id, user_id, from_label, from_lat, from_lon, to_label, to_lat, to_lon,
		ride_date, ride_time, flexibility, passengers, luggage, max_budget,
		ride_type, vehicle_preference, minimum_rating, verified_drivers_only,
		notes, route_miles, route_duration, price_estimate, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.DB.ExecContext(
		ctx,
		query,
		request.ID, request.UserID, request.FromLabel, request.FromLat, request.FromLon,
		request.ToLabel, request.ToLat, request.ToLon, request.RideDate, request.RideTime,
		request.Flexibility, request.Passengers, request.Luggage, request.MaxBudget,
		request.RideType, request.VehiclePreference, request.MinimumRating,
		request.VerifiedDriversOnly, request.Notes, request.RouteMiles,
		request.RouteDuration, request.PriceEstimate, request.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert ride request: %w", err)
	}

	return request, nil
}

func (s *Store) ListRideRequestsByUser(ctx context.Context, userID string) ([]models.RideRequest, error) {
	query := `SELECT
		id, user_id, from_label, from_lat, from_lon, to_label, to_lat, to_lon,
		ride_date, ride_time, flexibility, passengers, luggage, max_budget,
		ride_type, vehicle_preference, minimum_rating, verified_drivers_only,
		notes, route_miles, route_duration, price_estimate, created_at
	FROM ride_requests
	WHERE user_id = ?
	ORDER BY created_at DESC`

	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list ride requests: %w", err)
	}
	defer rows.Close()

	requests := []models.RideRequest{}
	for rows.Next() {
		var request models.RideRequest
		if err := rows.Scan(
			&request.ID, &request.UserID, &request.FromLabel, &request.FromLat, &request.FromLon,
			&request.ToLabel, &request.ToLat, &request.ToLon, &request.RideDate, &request.RideTime,
			&request.Flexibility, &request.Passengers, &request.Luggage, &request.MaxBudget,
			&request.RideType, &request.VehiclePreference, &request.MinimumRating,
			&request.VerifiedDriversOnly, &request.Notes, &request.RouteMiles,
			&request.RouteDuration, &request.PriceEstimate, &request.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ride request: %w", err)
		}
		requests = append(requests, request)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ride requests: %w", err)
	}

	return requests, nil
}

func (s *Store) GetRideRequestByID(ctx context.Context, requestID string) (*models.RideRequest, error) {
	query := `SELECT
		id, user_id, from_label, from_lat, from_lon, to_label, to_lat, to_lon,
		ride_date, ride_time, flexibility, passengers, luggage, max_budget,
		ride_type, vehicle_preference, minimum_rating, verified_drivers_only,
		notes, route_miles, route_duration, price_estimate, created_at
	FROM ride_requests
	WHERE id = ?`

	var request models.RideRequest
	err := s.DB.QueryRowContext(ctx, query, requestID).Scan(
		&request.ID, &request.UserID, &request.FromLabel, &request.FromLat, &request.FromLon,
		&request.ToLabel, &request.ToLat, &request.ToLon, &request.RideDate, &request.RideTime,
		&request.Flexibility, &request.Passengers, &request.Luggage, &request.MaxBudget,
		&request.RideType, &request.VehiclePreference, &request.MinimumRating,
		&request.VerifiedDriversOnly, &request.Notes, &request.RouteMiles,
		&request.RouteDuration, &request.PriceEstimate, &request.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("select ride request: %w", err)
	}

	return &request, nil
}

func (s *Store) DeleteRideRequest(ctx context.Context, requestID, userID string) error {
	result, err := s.DB.ExecContext(
		ctx,
		`DELETE FROM ride_requests WHERE id = ? AND user_id = ?`,
		requestID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("delete ride request: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete ride request rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *Store) UpdateRideRequest(ctx context.Context, input models.RideRequest) (*models.RideRequest, error) {
	query := `UPDATE ride_requests SET
		from_label = ?, from_lat = ?, from_lon = ?, to_label = ?, to_lat = ?, to_lon = ?,
		ride_date = ?, ride_time = ?, flexibility = ?, passengers = ?, luggage = ?, max_budget = ?,
		ride_type = ?, vehicle_preference = ?, minimum_rating = ?, verified_drivers_only = ?,
		notes = ?, route_miles = ?, route_duration = ?, price_estimate = ?
	WHERE id = ? AND user_id = ?`

	result, err := s.DB.ExecContext(
		ctx,
		query,
		input.FromLabel, input.FromLat, input.FromLon, input.ToLabel, input.ToLat, input.ToLon,
		input.RideDate, input.RideTime, input.Flexibility, input.Passengers, input.Luggage, input.MaxBudget,
		input.RideType, input.VehiclePreference, input.MinimumRating, input.VerifiedDriversOnly,
		input.Notes, input.RouteMiles, input.RouteDuration, input.PriceEstimate,
		input.ID, input.UserID,
	)
	if err != nil {
		return nil, fmt.Errorf("update ride request: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("update ride request rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

	return s.GetRideRequestByID(ctx, input.ID)
}
