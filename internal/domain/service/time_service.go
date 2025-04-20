// Package service это вспомогательные сервисы
package service

import (
	"time"
)

// TimeService -
type TimeService struct{}

// NewTimeService конструктор
func NewTimeService() TimeService {
	return TimeService{}
}

// GetTime возвращает время
func (t TimeService) GetTime() time.Time {
	return time.Now()
}
