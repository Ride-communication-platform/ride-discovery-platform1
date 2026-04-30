package handlers

import "testing"

func TestIsValidRideDate(t *testing.T) {
	tests := []struct {
		name string
		date string
		want bool
	}{
		{name: "past date", date: "2025-12-21", want: true},
		{name: "future date", date: "2027-10-05", want: true},
		{name: "leap day", date: "2028-02-29", want: true},
		{name: "invalid date", date: "03/20/2026", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidRideDate(tt.date); got != tt.want {
				t.Fatalf("isValidRideDate(%q) = %v, want %v", tt.date, got, tt.want)
			}
		})
	}
}
