package models

import "time"

type Trip struct {
	ID              string    `json:"id"`
	RideRequestID   string    `json:"rideRequestId"`
	RiderUserID     string    `json:"riderUserId"`
	DriverUserID    string    `json:"driverUserId"`
	PublishedRideID string    `json:"publishedRideId"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"createdAt"`
}

