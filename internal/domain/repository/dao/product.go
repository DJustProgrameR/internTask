// Package dao это dao для общения с репозиториями
package dao

import (
	"time"
)

// Product dao
type Product struct {
	ID          string
	DateTime    time.Time
	ReceptionID string
	Type        int16
}
