package bldgimpl

import (
	"context"
	"errors"
	b "unitly/internal/building"
	"unitly/internal/domain"

	"github.com/google/uuid"
)

type buildingService struct {
	buildings   b.BuildingRepository
	memberships b.MembershipRepository
	floors      b.FloorRepository
	units       b.UnitRepository
	details     b.BuildingDetailRepository
	tx          domain.UnitOfWork
}

func (s *buildingService) Create(ctx context.Context, actorUserID int64, in b.CreateBuildingInput) (*b.Building, error) {
	buildingID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	b := &b.Building{
		GUID:      buildingID,
		Name:      in.Name,
		Address:   in.Address,
		CreatedBy: actorUserID,
	}

	membershipID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	err = s.tx.Do(ctx, func(ctx context.Context) error {
		if err := s.buildings.Create(ctx, b); err != nil {
			return err
		}

		m := &domain.Membership{
			GUID:       membershipID,
			UserID:     actorUserID,
			BuildingID: b.ID,
			Role:       domain.RoleAdmin,
			Status:     domain.MembershipStatusActive,
		}
		if err := s.memberships.Create(ctx, m); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (s *buildingService) Get(ctx context.Context, id int64) (*b.Building, error) {
	return s.buildings.GetByID(ctx, id)
}

func (s *buildingService) GetDetail(ctx context.Context, buildingID int64) (*b.BuildingDetail, error) {
	detail, err := s.details.GetBuildingDetail(ctx, buildingID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, b.ErrBuildingNotFound
		}
	}
	return detail, err
}

func (s *buildingService) GetHighestLevelMembership(ctx context.Context, userID int64, buildingID uuid.UUID) (*domain.Membership, error) {
	id, err := s.buildings.GetInternalID(ctx, buildingID)
	if err != nil {
		return nil, err
	}
	return s.memberships.GetHighestActiveByUserBuilding(ctx, userID, id)
}

func (s *buildingService) ListMine(ctx context.Context, userID int64) ([]*b.BuildingSummary, error) {
	return s.buildings.ListSummaryByUser(ctx, userID)
}

func (s *buildingService) Delete(ctx context.Context, buildingID int64) error {
	return s.buildings.SoftDelete(ctx, buildingID) // authz (admin-only) enforced by middleware
}

func (s *buildingService) AddFloors(ctx context.Context, buildingID int64, in []b.NewFloor) ([]b.Floor, error) {
	if len(in) == 0 {
		return nil, b.ErrNoFloorsProvided
	}

	floors := make([]*b.Floor, len(in))
	for i, nf := range in {
		floors[i] = &b.Floor{
			BuildingID:  buildingID,
			FloorNumber: nf.FloorNumber,
			Caption:     nf.Caption,
		}
	}

	err := s.tx.Do(ctx, func(ctx context.Context) error {
		return s.floors.CreateBatch(ctx, floors)
	})
	if err != nil {
		return nil, err
	}

	out := make([]b.Floor, len(floors))
	for i, f := range floors {
		out[i] = *f
	}

	return out, nil
}

func (s *buildingService) GetFloor(ctx context.Context, buildingID int64, floorNumber int16) (*b.Floor, error) {
	return s.floors.GetByNumber(ctx, buildingID, floorNumber)
}

func (s *buildingService) GetFloorDetail(ctx context.Context, buildingID, floorID int64) (*b.FloorDetail, error) {
	return s.details.GetFloorDetail(ctx, buildingID, floorID)
}

func (s *buildingService) AddUnits(ctx context.Context, floorID int64, unitNumbers []string) ([]b.Unit, error) {
	if len(unitNumbers) == 0 {
		return nil, b.ErrNoUnitsProvided
	}

	seen := make(map[string]bool, len(unitNumbers))
	for _, n := range unitNumbers {
		if seen[n] {
			return nil, b.ErrUnitNumbersDuplicate
		}
		seen[n] = true
	}

	units := make([]*b.Unit, len(unitNumbers))
	for i, num := range unitNumbers {
		units[i] = &b.Unit{
			FloorID:    floorID,
			UnitNumber: num,
			Status:     b.UnitStatusVacant,
		}
	}

	if err := s.tx.Do(ctx, func(ctx context.Context) error {
		return s.units.CreateBatch(ctx, units)
	}); err != nil {
		return nil, err
	}

	out := make([]b.Unit, len(units))
	for i, u := range units {
		out[i] = *u
	}

	return out, nil
}

func NewBuildingService(
	buildings b.BuildingRepository,
	memberships b.MembershipRepository,
	floors b.FloorRepository,
	units b.UnitRepository,
	details b.BuildingDetailRepository,
	tx domain.UnitOfWork,
) b.BuildingService {
	if buildings == nil {
		panic("buildings cannot be nil")
	}
	if memberships == nil {
		panic("memberships cannot be nil")
	}
	if floors == nil {
		panic("floors cannot be nil")
	}
	if units == nil {
		panic("units cannot be nil")
	}
	if details == nil {
		panic("details cannot be nil")
	}
	if tx == nil {
		panic("tx cannot be nil")
	}
	return &buildingService{
		buildings:   buildings,
		memberships: memberships,
		floors:      floors,
		units:       units,
		details:     details,
		tx:          tx,
	}
}
