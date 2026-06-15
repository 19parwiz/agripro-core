package model

import "time"

// Plant is a crop record owned by a user.
type Plant struct {
	ID           string
	UserID       string
	Name         string
	Variety      string
	PlantingDate time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
