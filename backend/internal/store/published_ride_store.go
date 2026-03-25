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

func (s *Store) CreatePublishedRide(ctx context.Context, input models.PublishedRide) (*models.PublishedRide, error) {
	ride := &models.PublishedRide{
		ID:               uuid.NewString(),
		UserID:           input.UserID,
		FromLabel:        input.FromLabel,
		FromLat:          input.FromLat,
		FromLon:          input.FromLon,
		ToLabel:          input.ToLabel,
		ToLat:            input.ToLat,
		ToLon:            input.ToLon,
		RideDate:         input.RideDate,
		RideTime:         input.RideTime,
		Flexibility:      input.Flexibility,
		AvailableSeats:   input.AvailableSeats,
		TotalSeats:       input.TotalSeats,
		PricePerSeat:     input.PricePerSeat,
		VehicleType:      input.VehicleType,
		LuggageAllowed:   input.LuggageAllowed,
		RideType:         input.RideType,
		VehicleInfo:      input.VehicleInfo,
		Notes:            input.Notes,
		RouteMiles:       input.RouteMiles,
		RouteDuration:    input.RouteDuration,
		EarningsEstimate: input.EarningsEstimate,
		CreatedAt:        time.Now().UTC(),
	}

	query := `INSERT INTO published_rides (
		id, user_id, from_label, from_lat, from_lon, to_label, to_lat, to_lon,
		ride_date, ride_time, flexibility, available_seats, total_seats, price_per_seat,
		vehicle_type, luggage_allowed, ride_type, vehicle_info, notes, route_miles,
		route_duration, earnings_estimate, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.DB.ExecContext(
		ctx,
		query,
		ride.ID, ride.UserID, ride.FromLabel, ride.FromLat, ride.FromLon, ride.ToLabel, ride.ToLat, ride.ToLon,
		ride.RideDate, ride.RideTime, ride.Flexibility, ride.AvailableSeats, ride.TotalSeats, ride.PricePerSeat,
		ride.VehicleType, ride.LuggageAllowed, ride.RideType, ride.VehicleInfo, ride.Notes, ride.RouteMiles,
		ride.RouteDuration, ride.EarningsEstimate, ride.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert published ride: %w", err)
	}

	return ride, nil
}

func (s *Store) ListPublishedRidesByUser(ctx context.Context, userID string) ([]models.PublishedRide, error) {
	query := `SELECT
		id, user_id, from_label, from_lat, from_lon, to_label, to_lat, to_lon,
		ride_date, ride_time, flexibility, available_seats, total_seats, price_per_seat,
		vehicle_type, luggage_allowed, ride_type, vehicle_info, notes, route_miles,
		route_duration, earnings_estimate, created_at
	FROM published_rides
	WHERE user_id = ?
	ORDER BY created_at DESC`

	rows, err := s.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("list published rides: %w", err)
	}
	defer rows.Close()

	rides := []models.PublishedRide{}
	for rows.Next() {
		var ride models.PublishedRide
		if err := rows.Scan(
			&ride.ID, &ride.UserID, &ride.FromLabel, &ride.FromLat, &ride.FromLon, &ride.ToLabel, &ride.ToLat, &ride.ToLon,
			&ride.RideDate, &ride.RideTime, &ride.Flexibility, &ride.AvailableSeats, &ride.TotalSeats, &ride.PricePerSeat,
			&ride.VehicleType, &ride.LuggageAllowed, &ride.RideType, &ride.VehicleInfo, &ride.Notes, &ride.RouteMiles,
			&ride.RouteDuration, &ride.EarningsEstimate, &ride.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan published ride: %w", err)
		}
		rides = append(rides, ride)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate published rides: %w", err)
	}

	return rides, nil
}

func (s *Store) ListRideRequestFeed(ctx context.Context, userID, dateFilter, routeFilter string, passengerCount int, maxBudget float64) ([]models.RideRequestFeedItem, error) {
	query := `SELECT
		rr.id, rr.user_id, rr.from_label, rr.from_lat, rr.from_lon, rr.to_label, rr.to_lat, rr.to_lon,
		rr.ride_date, rr.ride_time, rr.flexibility, rr.passengers, rr.luggage, rr.max_budget,
		rr.ride_type, rr.vehicle_preference, rr.minimum_rating, rr.verified_drivers_only,
		rr.notes, rr.route_miles, rr.route_duration, rr.price_estimate, rr.created_at,
		u.name, u.rating, u.rating_count
	FROM ride_requests rr
	JOIN users u ON u.id = rr.user_id
	WHERE rr.user_id != ?`

	args := []any{userID}
	if strings.TrimSpace(dateFilter) != "" {
		query += ` AND rr.ride_date = ?`
		args = append(args, strings.TrimSpace(dateFilter))
	}
	if strings.TrimSpace(routeFilter) != "" {
		query += ` AND (LOWER(rr.from_label) LIKE ? OR LOWER(rr.to_label) LIKE ?)`
		routeLike := "%" + strings.ToLower(strings.TrimSpace(routeFilter)) + "%"
		args = append(args, routeLike, routeLike)
	}
	if passengerCount > 0 {
		query += ` AND rr.passengers = ?`
		args = append(args, passengerCount)
	}
	if maxBudget > 0 {
		query += ` AND rr.max_budget <= ?`
		args = append(args, maxBudget)
	}

	query += ` ORDER BY rr.created_at DESC`

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list ride request feed: %w", err)
	}
	defer rows.Close()

	items := []models.RideRequestFeedItem{}
	for rows.Next() {
		var item models.RideRequestFeedItem
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.FromLabel, &item.FromLat, &item.FromLon, &item.ToLabel, &item.ToLat, &item.ToLon,
			&item.RideDate, &item.RideTime, &item.Flexibility, &item.Passengers, &item.Luggage, &item.MaxBudget,
			&item.RideType, &item.VehiclePreference, &item.MinimumRating, &item.VerifiedDriversOnly,
			&item.Notes, &item.RouteMiles, &item.RouteDuration, &item.PriceEstimate, &item.CreatedAt,
			&item.RequesterName, &item.RequesterRating, &item.RequesterRatingCount,
		); err != nil {
			return nil, fmt.Errorf("scan ride request feed item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate ride request feed: %w", err)
	}

	return items, nil
}

func (s *Store) CreateRideRequestAction(ctx context.Context, input models.RideRequestAction) (*models.RideRequestAction, error) {
	action := &models.RideRequestAction{
		ID:              uuid.NewString(),
		RideRequestID:   input.RideRequestID,
		DriverUserID:    input.DriverUserID,
		PublishedRideID: input.PublishedRideID,
		Action:          input.Action,
		Message:         input.Message,
		CreatedAt:       time.Now().UTC(),
	}

	query := `INSERT INTO ride_request_actions (
		id, ride_request_id, driver_user_id, published_ride_id, action, message, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.DB.ExecContext(ctx, query, action.ID, action.RideRequestID, action.DriverUserID, action.PublishedRideID, action.Action, action.Message, action.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert ride request action: %w", err)
	}

	return action, nil
}

func (s *Store) GetPublishedRideByID(ctx context.Context, rideID string) (*models.PublishedRide, error) {
	query := `SELECT
		id, user_id, from_label, from_lat, from_lon, to_label, to_lat, to_lon,
		ride_date, ride_time, flexibility, available_seats, total_seats, price_per_seat,
		vehicle_type, luggage_allowed, ride_type, vehicle_info, notes, route_miles,
		route_duration, earnings_estimate, created_at
	FROM published_rides WHERE id = ?`

	var ride models.PublishedRide
	err := s.DB.QueryRowContext(ctx, query, rideID).Scan(
		&ride.ID, &ride.UserID, &ride.FromLabel, &ride.FromLat, &ride.FromLon, &ride.ToLabel, &ride.ToLat, &ride.ToLon,
		&ride.RideDate, &ride.RideTime, &ride.Flexibility, &ride.AvailableSeats, &ride.TotalSeats, &ride.PricePerSeat,
		&ride.VehicleType, &ride.LuggageAllowed, &ride.RideType, &ride.VehicleInfo, &ride.Notes, &ride.RouteMiles,
		&ride.RouteDuration, &ride.EarningsEstimate, &ride.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("select published ride: %w", err)
	}
	return &ride, nil
}
