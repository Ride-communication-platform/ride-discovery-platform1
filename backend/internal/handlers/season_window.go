package handlers

import "time"

func isValidRideDate(value string) bool {
	_, err := time.Parse(time.DateOnly, value)
	return err == nil
}

func rideDateMessage() string {
	return "Ride date must use YYYY-MM-DD"
}
