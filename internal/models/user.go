package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Wins         int       `json:"wins"`
	Losses       int       `json:"losses"`
	CreatedAt    time.Time `json:"created_at"`
}
