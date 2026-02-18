package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Rating       float64   `json:"rating"`
	CreatedAt    time.Time `json:"createdAt"`
}
