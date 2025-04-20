// Package dao это dao для общения с репозиториями
package dao

import (
	"time"
)

// Reception dao
type Reception struct {
	ID       string
	PVZID    string
	DateTime time.Time
	Status   int8
}
