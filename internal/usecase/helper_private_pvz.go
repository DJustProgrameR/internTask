// Package usecase это юзкейсы и вспомогательная логика
package usecase

import (
	"errors"
	"google.golang.org/protobuf/types/known/timestamppb"
	"internshipPVZ/internal/domain/model"
	"internshipPVZ/internal/domain/repository/dao"
	pb "internshipPVZ/internal/grpc/models"
	"internshipPVZ/internal/http/onlymodels"
	"sort"
)

func pvzDaoToDto(pvzDao *dao.PVZ) (*onlymodels.PVZ, error) {
	if pvzDao == nil {
		return nil, errors.New("pvzDao is nil")
	}
	id, err := validateRawID(pvzDao.ID)
	if err != nil {
		return nil, err
	}
	dto := &onlymodels.PVZ{
		Id:               &id,
		RegistrationDate: &pvzDao.RegistrationDate,
		City:             onlymodels.PVZCity(model.NewCity(pvzDao.City)),
	}
	return dto, nil
}

func pvzsToGrpcDto(input []*dao.PVZ) ([]*pb.PVZ, error) {
	if input == nil {
		return nil, errors.New("input is nil")
	}
	var result []*pb.PVZ
	for _, item := range input {
		if item == nil {
			return nil, errors.New("pvzs is nil")
		}
		dto := &pb.PVZ{
			Id:               item.ID,
			RegistrationDate: timestamppb.New(item.RegistrationDate),
			City:             model.NewCity(item.City).Get(),
		}
		result = append(result, dto)
	}

	return result, nil
}

func pvzListToDto(input []*dao.PVZList) (*onlymodels.GetFilteredResponse, error) {
	if input == nil {
		return nil, errors.New("input is nil")
	}
	pvzGroups := make(map[string][]*dao.PVZList)
	for _, item := range input {
		if item == nil {
			return nil, errors.New("pvz list is nil")
		}
		pvzGroups[item.PvzID] = append(pvzGroups[item.PvzID], item)
	}

	var result onlymodels.GetFilteredResponse
	for _, items := range pvzGroups {
		first := items[0]
		id, err := validateRawID(first.PvzID)
		if err != nil {
			return nil, err
		}
		pvz := &onlymodels.PVZ{
			Id:               &id,
			RegistrationDate: &first.RegistrationDate,
			City:             onlymodels.PVZCity(model.NewCity(first.City).Get()),
		}

		receptionGroups := make(map[string][]*dao.PVZList)
		for _, item := range items {
			receptionGroups[item.ReceptionID] = append(receptionGroups[item.ReceptionID], item)
		}

		var receptions onlymodels.GetFilteredResponseReceptions
		for _, recItems := range receptionGroups {
			firstRec := recItems[0]
			id, err := validateRawID(firstRec.ReceptionID)
			if err != nil {
				return nil, err
			}
			pvzID, err := validateRawID(firstRec.PvzID)
			if err != nil {
				return nil, err
			}
			reception := &onlymodels.Reception{
				Id:       &id,
				DateTime: firstRec.ReceptionDateTime,
				PvzId:    pvzID,
				Status:   onlymodels.ReceptionStatus(model.NewReceptionStatus(firstRec.Status)),
			}

			var products []onlymodels.Product
			for _, r := range recItems {
				if r.ProductID.Valid {
					id, err := validateRawID(r.ProductID.String)
					if err != nil {
						return nil, err
					}
					receptionID, err := validateRawID(r.ReceptionID)
					if err != nil {
						return nil, err
					}
					p := onlymodels.Product{
						Id:          &id,
						DateTime:    &r.ProductDateTime.Time,
						Type:        onlymodels.ProductType(model.NewProductType(r.Type.Int16)),
						ReceptionId: receptionID,
					}
					products = append(products, p)
				}
			}

			prodPointer := &products
			if len(products) == 0 {
				prodPointer = nil
			}

			receptions = append(receptions, struct {
				Products  *[]onlymodels.Product `json:"products,omitempty"`
				Reception *onlymodels.Reception `json:"reception,omitempty"`
			}{
				Reception: reception,
				Products:  prodPointer,
			})
		}

		sort.Slice(receptions, func(i, j int) bool {
			return receptions[i].Reception.DateTime.Before(receptions[j].Reception.DateTime)
		})

		result = append(result, struct {
			Pvz        *onlymodels.PVZ                           `json:"pvz,omitempty"`
			Receptions *onlymodels.GetFilteredResponseReceptions `json:"receptions,omitempty"`
		}{
			Pvz:        pvz,
			Receptions: &receptions,
		})
	}

	return &result, nil
}
