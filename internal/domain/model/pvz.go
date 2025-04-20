// Package model это доменные сущности и типы
package model

import (
	"github.com/google/uuid"
	"internshipPVZ/internal/domain/repository/dao"
	"time"
)

// City тип города
type City string

// названия городов
const (
	CityMoscow  City = "Москва"
	CitySPB     City = "Санкт-Петербург"
	CityKazan   City = "Казань"
	CityDefault City = ""
)

// Get возвращает строковое представление города.
func (c City) Get() string {
	return string(c)
}

// ToInt возвращает целочисленное представление города.
func (c City) ToInt() int32 {
	mapp := map[City]int32{
		CitySPB:     0,
		CityKazan:   1,
		CityMoscow:  2,
		CityDefault: 3,
	}
	return mapp[c]
}

// NewCity конструктор для создания типа города из целочисленного значения.
func NewCity(num int32) City {
	mapp := map[int32]City{
		0: CitySPB,
		1: CityKazan,
		2: CityMoscow,
		3: CityDefault,
	}
	return mapp[num]
}

// PVZ сущность пункта выдачи заказов (ПВЗ)
type PVZ struct {
	ID               uuid.UUID
	RegistrationDate time.Time
	City             City
}

// ToDao преобразует сущность ПВЗ в DAO объект.
func (pvz PVZ) ToDao() *dao.PVZ {
	return &dao.PVZ{ID: pvz.ID.String(), RegistrationDate: pvz.RegistrationDate, City: pvz.City.ToInt()}
}
