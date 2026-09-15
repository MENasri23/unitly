package bldgdto

import "unitly/internal/building"

type UnitDTO struct {
	UnitNumber string              `json:"unit_number"`
	Caption    *string             `json:"caption"`
	Status     building.UnitStatus `json:"status"`
	HasParking bool                `json:"has_parking"`
}

type UnitSummaryDTO struct {
	UnitNumber string              `json:"unit_number"`
	Caption    *string             `json:"caption"`
	Status     building.UnitStatus `json:"status"`
}

func MapUnitToDTO(u *building.Unit) UnitDTO {
	return UnitDTO{
		UnitNumber: u.UnitNumber(),
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
			UnitNumber: u.UnitNumber(),
			Caption:    u.Caption,
			Status:     u.Status,
			HasParking: u.HasParking,
		}
	}
	return result
}

func MapUnitSummaryToDTO(u *building.UnitSummary) UnitDTO {
	return UnitDTO{
		UnitNumber: u.UnitNumber,
		Caption:    u.Caption,
		Status:     u.Status,
	}
}

func MapUnitSummariesToDTOs(units []building.UnitSummary) []UnitDTO {
	if len(units) == 0 {
		return nil
	}

	result := make([]UnitDTO, len(units))
	for i, u := range units {
		result[i] = UnitDTO{
			UnitNumber: u.UnitNumber,
			Caption:    u.Caption,
			Status:     u.Status,
		}
	}
	return result
}
