package model

import "time"

// Device is a grow unit or sensor hub owned by a user.
type Device struct {
	ID         string
	UserID     string
	Name       string
	DeviceID   string
	Type       string
	Location   string
	StreamPath string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
