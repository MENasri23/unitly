package bldgdto

import "unitly/internal/building"

type UnitDTO struct {
	UnitNumber string              `json:"unit_number"`
	Caption    *string             `json:"caption"`
	Status     building.UnitStatus `json:"status"`
	HasParking bool                `json:"has_parking"`
}

func MapUnitToDTO(u *building.Unit) UnitDTO {
	return UnitDTO{
		UnitNumber: u.UnitNumber,
		Caption:    u.Caption,
		Status:     u.Status,
		HasParking: u.HasParking,
	}
}

func MapUnitsToUnitDTOs(units []building.Unit) []UnitDTO {
	if len(units) == 0 {
		return nil
	}

	result := make([]UnitDTO, len(units))
	for i, u := range units {
		result[i] = UnitDTO{
			UnitNumber: u.UnitNumber,
			Caption:    u.Caption,
			Status:     u.Status,
			HasParking: u.HasParking,
		}
	}
	return result
}

func MapUnitDetailToDTO(u *building.UnitDetail) UnitDTO {
	return UnitDTO{
		UnitNumber: u.UnitNumber,
		Caption:    u.Caption,
		Status:     u.Status,
		HasParking: u.HasParking,
	}
}

func MapUnitDetailsToDTOs(units []building.UnitDetail) []UnitDTO {
	if len(units) == 0 {
		return nil
	}

	result := make([]UnitDTO, len(units))
	for i, u := range units {
		result[i] = UnitDTO{
			UnitNumber: u.UnitNumber,
			Caption:    u.Caption,
			Status:     u.Status,
			HasParking: u.HasParking,
		}
	}
	return result
}
