package models

import "time"

// User represents a user account
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never send password in JSON
	GroupID      int       `json:"group_id"`
	IsAdmin      bool      `json:"is_admin"`
	CreatedAt    time.Time `json:"created_at"`
}

// Group represents a research group
type Group struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CPUQuota  int       `json:"cpu_quota"`
	Priority  int       `json:"priority"`
	CreatedAt time.Time `json:"created_at"`
}