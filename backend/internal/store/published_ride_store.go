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
		Status:           "active",
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
		id, user_id, status, from_label, from_lat, from_lon, to_label, to_lat, to_lon,
		ride_date, ride_time, flexibility, available_seats, total_seats, price_per_seat,
		vehicle_type, luggage_allowed, ride_type, vehicle_info, notes, route_miles,
		route_duration, earnings_estimate, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.DB.ExecContext(
		ctx,
		query,
		ride.ID, ride.UserID, ride.Status, ride.FromLabel, ride.FromLat, ride.FromLon, ride.ToLabel, ride.ToLat, ride.ToLon,
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
		id, user_id, status, from_label, from_lat, from_lon, to_label, to_lat, to_lon,
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
			&ride.ID, &ride.UserID, &ride.Status, &ride.FromLabel, &ride.FromLat, &ride.FromLon, &ride.ToLabel, &ride.ToLat, &ride.ToLon,
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

func (s *Store) ListPublishedRideFeed(ctx context.Context, userID, dateFilter, routeFilter, rideType, vehicleType, luggageAllowed, flexibility string, seats int, maxPrice float64) ([]models.PublishedRideFeedItem, error) {
	query := `SELECT
		pr.id, pr.user_id, pr.status, pr.from_label, pr.from_lat, pr.from_lon, pr.to_label, pr.to_lat, pr.to_lon,
		pr.ride_date, pr.ride_time, pr.flexibility, pr.available_seats, pr.total_seats, pr.price_per_seat,
		pr.vehicle_type, pr.luggage_allowed, pr.ride_type, pr.vehicle_info, pr.notes, pr.route_miles,
		pr.route_duration, pr.earnings_estimate, pr.created_at,
		u.name, u.rating, u.rating_count
	FROM published_rides pr
	JOIN users u ON u.id = pr.user_id
	WHERE pr.user_id != ? AND pr.status = 'active'`

	args := []any{userID}
	if strings.TrimSpace(dateFilter) != "" {
		query += ` AND pr.ride_date = ?`
		args = append(args, strings.TrimSpace(dateFilter))
	}
	if strings.TrimSpace(routeFilter) != "" {
		query += ` AND (LOWER(pr.from_label) LIKE ? OR LOWER(pr.to_label) LIKE ?)`
		routeLike := "%" + strings.ToLower(strings.TrimSpace(routeFilter)) + "%"
		args = append(args, routeLike, routeLike)
	}
	if strings.TrimSpace(rideType) != "" {
		query += ` AND pr.ride_type = ?`
		args = append(args, strings.TrimSpace(rideType))
	}
	if strings.TrimSpace(vehicleType) != "" {
		query += ` AND pr.vehicle_type = ?`
		args = append(args, strings.TrimSpace(vehicleType))
	}
	if strings.TrimSpace(luggageAllowed) != "" {
		query += ` AND pr.luggage_allowed = ?`
		args = append(args, strings.TrimSpace(luggageAllowed))
	}
	if strings.TrimSpace(flexibility) != "" {
		query += ` AND pr.flexibility = ?`
		args = append(args, strings.TrimSpace(flexibility))
	}
	if seats > 0 {
		query += ` AND pr.available_seats >= ?`
		args = append(args, seats)
	}
	if maxPrice > 0 {
		query += ` AND pr.price_per_seat <= ?`
		args = append(args, maxPrice)
	}

	query += ` ORDER BY pr.created_at DESC`

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list published ride feed: %w", err)
	}
	defer rows.Close()

	items := []models.PublishedRideFeedItem{}
	for rows.Next() {
		var item models.PublishedRideFeedItem
		if err := rows.Scan(
			&item.ID, &item.UserID, &item.Status, &item.FromLabel, &item.FromLat, &item.FromLon, &item.ToLabel, &item.ToLat, &item.ToLon,
			&item.RideDate, &item.RideTime, &item.Flexibility, &item.AvailableSeats, &item.TotalSeats, &item.PricePerSeat,
			&item.VehicleType, &item.LuggageAllowed, &item.RideType, &item.VehicleInfo, &item.Notes, &item.RouteMiles,
			&item.RouteDuration, &item.EarningsEstimate, &item.CreatedAt,
			&item.DriverName, &item.DriverRating, &item.DriverRatingCount,
		); err != nil {
			return nil, fmt.Errorf("scan published ride feed item: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate published ride feed: %w", err)
	}

	return items, nil
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
		id, user_id, status, from_label, from_lat, from_lon, to_label, to_lat, to_lon,
		ride_date, ride_time, flexibility, available_seats, total_seats, price_per_seat,
		vehicle_type, luggage_allowed, ride_type, vehicle_info, notes, route_miles,
		route_duration, earnings_estimate, created_at
	FROM published_rides WHERE id = ?`

	var ride models.PublishedRide
	err := s.DB.QueryRowContext(ctx, query, rideID).Scan(
		&ride.ID, &ride.UserID, &ride.Status, &ride.FromLabel, &ride.FromLat, &ride.FromLon, &ride.ToLabel, &ride.ToLat, &ride.ToLon,
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

func (s *Store) SetPublishedRideStatus(ctx context.Context, rideID, ownerUserID, status string) error {
	result, err := s.DB.ExecContext(ctx, `UPDATE published_rides SET status = ? WHERE id = ? AND user_id = ?`, strings.TrimSpace(status), strings.TrimSpace(rideID), strings.TrimSpace(ownerUserID))
	if err != nil {
		return fmt.Errorf("update published ride status: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("published ride status affected rows: %w", err)
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
