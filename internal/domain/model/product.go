// Package model это доменные сущности и типы
package model

import (
	"github.com/google/uuid"
	"internshipPVZ/internal/domain/repository/dao"
	"time"
)

// ProductType тип продукта
type ProductType string

// типы продуктов
const (
	ProductElectronics ProductType = "электроника"
	ProductClothes     ProductType = "одежда"
	ProductShoes       ProductType = "обувь"
	ProductDefault     ProductType = ""
)

// Get возвращает строковое представление типа продукта.
func (p ProductType) Get() string {
	return string(p)
}

// ToInt возвращает целочисленное представление типа продукта.
func (p ProductType) ToInt() int16 {
	mapp := map[ProductType]int16{
		ProductElectronics: 0,
		ProductShoes:       1,
		ProductClothes:     2,
		ProductDefault:     3,
	}
	return mapp[p]
}

// NewProductType конструктор для создания типа продукта из целочисленного значения.
func NewProductType(num int16) ProductType {
	mapp := map[int16]ProductType{
		0: ProductElectronics,
		1: ProductShoes,
		2: ProductClothes,
		3: ProductDefault,
	}
	return mapp[num]
}

// Product сущность продукта
type Product struct {
	ID          uuid.UUID
	DateTime    time.Time
	ReceptionID uuid.UUID
	Type        ProductType
}

// ToDao преобразует сущность продукта в DAO объект.
func (p Product) ToDao() *dao.Product {
	return &dao.Product{DateTime: p.DateTime, Type: p.Type.ToInt(), ReceptionID: p.ReceptionID.String()}
}
