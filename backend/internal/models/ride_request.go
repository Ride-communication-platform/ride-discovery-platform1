package models

import "time"

type RideRequest struct {
	ID                  string    `json:"id"`
	UserID              string    `json:"userId"`
	FromLabel           string    `json:"fromLabel"`
	FromLat             float64   `json:"fromLat"`
	FromLon             float64   `json:"fromLon"`
	ToLabel             string    `json:"toLabel"`
	ToLat               float64   `json:"toLat"`
	ToLon               float64   `json:"toLon"`
	RideDate            string    `json:"rideDate"`
	RideTime            string    `json:"rideTime"`
	Flexibility         string    `json:"flexibility"`
	Passengers          int       `json:"passengers"`
	Luggage             string    `json:"luggage"`
	MaxBudget           float64   `json:"maxBudget"`
	RideType            string    `json:"rideType"`
	VehiclePreference   string    `json:"vehiclePreference"`
	MinimumRating       float64   `json:"minimumRating"`
	VerifiedDriversOnly bool      `json:"verifiedDriversOnly"`
	Notes               string    `json:"notes"`
	RouteMiles          int       `json:"routeMiles"`
	RouteDuration       string    `json:"routeDuration"`
	PriceEstimate       string    `json:"priceEstimate"`
	CreatedAt           time.Time `json:"createdAt"`
}
