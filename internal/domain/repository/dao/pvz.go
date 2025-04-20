// Package dao это dao для общения с репозиториями
package dao

import (
	"database/sql"
	"time"
)

// PVZ dao
type PVZ struct {
	ID               string
	RegistrationDate time.Time
	City             int32
}

// PVZList dao
type PVZList struct {
	PvzID             string
	RegistrationDate  time.Time
	City              int32
	ReceptionID       string
	ReceptionDateTime time.Time
	Status            int8
	ProductID         sql.NullString
	ProductDateTime   sql.NullTime
	Type              sql.NullInt16
}
