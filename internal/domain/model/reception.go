// Package model это доменные сущности и типы
package model

import (
	"github.com/google/uuid"
	"internshipPVZ/internal/domain/repository/dao"
	"time"
)

// ReceptionStatus статус приёмки
type ReceptionStatus string

// статусы приёмок
const (
	ReceptionInProgress ReceptionStatus = "in_progress"
	ReceptionClosed     ReceptionStatus = "close"
)

// Get возвращает строковое представление статуса приёмки.
func (rs ReceptionStatus) Get() string {
	return string(rs)
}

// ToInt возвращает целочисленное представление статуса приёмки.
func (rs ReceptionStatus) ToInt() int8 {
	mapp := map[ReceptionStatus]int8{
		ReceptionInProgress: 0,
		ReceptionClosed:     1,
	}
	return mapp[rs]
}

// NewReceptionStatus конструктор для создания статуса приёмки из целочисленного значения.
func NewReceptionStatus(num int8) ReceptionStatus {
	mapp := map[int8]ReceptionStatus{
		0: ReceptionInProgress,
		1: ReceptionClosed,
	}
	return mapp[num]
}

// Reception сущность приёмки
type Reception struct {
	ID       uuid.UUID
	PVZID    uuid.UUID
	DateTime time.Time
	Status   ReceptionStatus
}

// ToDao преобразует сущность приёмки в DAO объект.
func (r Reception) ToDao() *dao.Reception {
	return &dao.Reception{ID: r.ID.String(), PVZID: r.PVZID.String(), DateTime: r.DateTime, Status: r.Status.ToInt()}
}
