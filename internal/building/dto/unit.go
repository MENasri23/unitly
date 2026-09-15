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

func MapUnitsToSummaryDTOs(units []building.Unit) []UnitSummaryDTO {
	if len(units) == 0 {
		return nil
	}

	result := make([]UnitSummaryDTO, len(units))
	for i, u := range units {
		result[i] = UnitSummaryDTO{
			UnitNumber: u.UnitNumber(),
			Caption:    u.Caption,
			Status:     u.Status,
		}
	}
	return result
}

func MapUnitSummariesToDTOs(units []building.UnitSummary) []UnitSummaryDTO {
	if len(units) == 0 {
		return nil
	}

	result := make([]UnitSummaryDTO, len(units))
	for i, u := range units {
		result[i] = UnitSummaryDTO{
			UnitNumber: u.UnitNumber,
			Caption:    u.Caption,
			Status:     u.Status,
		}
	}
	return result
}
