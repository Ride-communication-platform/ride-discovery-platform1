package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	AvatarData       string    `json:"avatarData"`
	Interests        []string  `json:"interests"`
	PasswordHash     string    `json:"-"`
	Rating           float64   `json:"rating"`
	RatingCount      int       `json:"ratingCount"`
	TripsCompleted   int       `json:"tripsCompleted"`
	EmailVerified    bool      `json:"emailVerified"`
	VerificationCode string    `json:"-"`
	PasswordResetCode   string       `json:"-"`
	PasswordResetSentAt sql.NullTime `json:"-"`
	AuthProvider     string    `json:"authProvider"`
	CreatedAt        time.Time `json:"createdAt"`
}
