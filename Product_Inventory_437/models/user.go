package models

import "time"

type User struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	APIToken     string    `json:"api_token,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserInput struct {
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
}