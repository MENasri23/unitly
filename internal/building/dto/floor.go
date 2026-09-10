package bldgdto

import "unitly/internal/building"

type FloorDTO struct {
	FloorNumber int16
	Caption     *string
	Units       []UnitDTO
}

func MapFloorDetailToDTO(f *building.FloorDetail) FloorDTO {
	return FloorDTO{
		FloorNumber: f.FloorNumber,
		Caption:     f.Caption,
		Units:       MapUnitDetailsToDTOs(f.Units),
	}
}

func MapFloorDetailsToDTOs(floors []building.FloorDetail) []FloorDTO {
	if len(floors) == 0 {
		return nil
	}

	result := make([]FloorDTO, len(floors))
	for i, f := range floors {
		result[i] = FloorDTO{
			FloorNumber: f.FloorNumber,
			Caption:     f.Caption,
			Units:       MapUnitDetailsToDTOs(f.Units),
		}
	}
	return result
}

func MapFloorToDTO(f *building.Floor) FloorDTO {
	return FloorDTO{
		FloorNumber: f.FloorNumber,
		Caption:     f.Caption,
	}
}

func MapFloorsToDTOs(floors []building.Floor) []FloorDTO {
	if len(floors) == 0 {
		return nil
	}

	result := make([]FloorDTO, len(floors))
	for i, f := range floors {
		result[i] = MapFloorToDTO(&f)
	}
	return result
}
