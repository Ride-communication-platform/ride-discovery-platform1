package models

import "time"

type PublishedRide struct {
	ID               string    `json:"id"`
	UserID           string    `json:"userId"`
	Status           string    `json:"status"`
	FromLabel        string    `json:"fromLabel"`
	FromLat          float64   `json:"fromLat"`
	FromLon          float64   `json:"fromLon"`
	ToLabel          string    `json:"toLabel"`
	ToLat            float64   `json:"toLat"`
	ToLon            float64   `json:"toLon"`
	RideDate         string    `json:"rideDate"`
	RideTime         string    `json:"rideTime"`
	Flexibility      string    `json:"flexibility"`
	AvailableSeats   int       `json:"availableSeats"`
	TotalSeats       int       `json:"totalSeats"`
	PricePerSeat     float64   `json:"pricePerSeat"`
	VehicleType      string    `json:"vehicleType"`
	LuggageAllowed   string    `json:"luggageAllowed"`
	RideType         string    `json:"rideType"`
	VehicleInfo      string    `json:"vehicleInfo"`
	Notes            string    `json:"notes"`
	RouteMiles       int       `json:"routeMiles"`
	RouteDuration    string    `json:"routeDuration"`
	EarningsEstimate float64   `json:"earningsEstimate"`
	CreatedAt        time.Time `json:"createdAt"`
}

type PublishedRideFeedItem struct {
	PublishedRide
	DriverName        string  `json:"driverName"`
	DriverRating      float64 `json:"driverRating"`
	DriverRatingCount int     `json:"driverRatingCount"`
}

type RideRequestFeedItem struct {
	RideRequest
	RequesterName       string  `json:"requesterName"`
	RequesterRating     float64 `json:"requesterRating"`
	RequesterRatingCount int    `json:"requesterRatingCount"`
}

type RideRequestAction struct {
	ID              string    `json:"id"`
	RideRequestID   string    `json:"rideRequestId"`
	DriverUserID    string    `json:"driverUserId"`
	PublishedRideID string    `json:"publishedRideId"`
	Action          string    `json:"action"`
	Message         string    `json:"message"`
	CreatedAt       time.Time `json:"createdAt"`
}
