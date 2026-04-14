package models

import "time"

type TripView struct {
	ID              string    `json:"id"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`

	RideRequestID   string `json:"rideRequestId"`
	PublishedRideID string `json:"publishedRideId"`

	RiderUserID  string `json:"riderUserId"`
	RiderName    string `json:"riderName"`
	DriverUserID string `json:"driverUserId"`
	DriverName   string `json:"driverName"`

	FromLabel      string `json:"fromLabel"`
	ToLabel        string `json:"toLabel"`
	RideDate       string `json:"rideDate"`
	RideTime       string `json:"rideTime"`
	Passengers     int    `json:"passengers"`
	RouteMiles     int    `json:"routeMiles"`
	RouteDuration  string `json:"routeDuration"`
	PricePerSeat   float64 `json:"pricePerSeat"`
	EstimatedTotal float64 `json:"estimatedTotal"`
}

