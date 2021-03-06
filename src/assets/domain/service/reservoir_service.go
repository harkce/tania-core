package service

import (
	"github.com/Tanibox/tania-server/src/assets/domain"
	"github.com/Tanibox/tania-server/src/assets/query"
	"github.com/Tanibox/tania-server/src/assets/storage"
	uuid "github.com/satori/go.uuid"
)

type ReservoirServiceInMemory struct {
	FarmReadQuery query.FarmReadQuery
}

func (s ReservoirServiceInMemory) FindFarmByID(uid uuid.UUID) (domain.ReservoirFarmServiceResult, error) {
	result := <-s.FarmReadQuery.FindByID(uid)

	if result.Error != nil {
		return domain.ReservoirFarmServiceResult{}, result.Error
	}

	farm, ok := result.Result.(storage.FarmRead)

	if !ok {
		return domain.ReservoirFarmServiceResult{}, domain.ReservoirError{Code: domain.ReservoirErrorFarmNotFound}
	}

	if farm == (storage.FarmRead{}) {
		return domain.ReservoirFarmServiceResult{}, domain.ReservoirError{Code: domain.ReservoirErrorFarmNotFound}
	}

	return domain.ReservoirFarmServiceResult{
		UID:  farm.UID,
		Name: farm.Name,
	}, nil
}
