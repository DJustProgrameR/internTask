// Package usecase это юзкейсы и вспомогательная логика
package usecase

import (
	"errors"
	"github.com/google/uuid"
	"internshipPVZ/internal/domain/model"
	"internshipPVZ/internal/domain/repository/dao"
	"internshipPVZ/internal/http/onlymodels"
)

func receptionDaoToDto(reception *dao.Reception) (*onlymodels.Reception, error) {
	if reception == nil {
		return nil, errors.New("reception is nil")
	}
	id, err := validateRawID(reception.ID)
	if err != nil {
		return nil, err
	}
	pvzID, err := uuid.Parse(reception.PVZID)
	if err != nil {
		return nil, err
	}
	dto := &onlymodels.Reception{
		Id:       &id,
		DateTime: reception.DateTime,
		Status:   onlymodels.ReceptionStatus(model.NewReceptionStatus(reception.Status)),
		PvzId:    pvzID,
	}
	return dto, nil
}
