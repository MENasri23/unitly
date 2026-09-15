package bldgimpl

import (
	"context"
	"errors"
	"strconv"
	b "unitly/internal/building"
	"unitly/internal/domain"

	"github.com/google/uuid"
)

type buildingService struct {
	buildings   b.BuildingRepository
	memberships b.MembershipRepository
	floors      b.FloorRepository
	units       b.UnitRepository
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

		return s.provision(ctx, b.ID, in)
	})
	if err != nil {
		return nil, err
	}

	return b, nil
}

// provision auto-generates floors/units from the optional counts supplied
// at creation. Floor 0 is always ground; floors_ground_and_top
// is the total count from ground upward inclusive (e.g. 5 -> floors
// 0,1,2,3,4), floors_under_ground the count below it, numbered -1,-2,...
// Runs inside the same transaction as Create, via the shared ctx.
func (s *buildingService) provision(ctx context.Context, buildingID int64, in b.CreateBuildingInput) error {
	if in.FloorsGroundAndTop == nil {
		return nil
	}

	newFloors := make([]b.NewFloor, 0, int(*in.FloorsGroundAndTop))
	for n := int16(0); n < *in.FloorsGroundAndTop; n++ {
		newFloors = append(newFloors, b.NewFloor{FloorNumber: n})
	}
	if in.FloorsUnderGround != nil {
		for n := int16(1); n <= *in.FloorsUnderGround; n++ {
			newFloors = append(newFloors, b.NewFloor{FloorNumber: -n})
		}
	}

	floors, err := s.AddFloors(ctx, buildingID, newFloors)
	if err != nil {
		return err
	}
	if in.DefaultUnitsPerFloor == nil {
		return nil
	}

	unitNumbers := make([]string, *in.DefaultUnitsPerFloor)
	for i := range unitNumbers {
		unitNumbers[i] = strconv.Itoa(i + 1)
	}
	for _, f := range floors {
		if _, err := s.AddUnits(ctx, f.ID, unitNumbers); err != nil {
			return err
		}
	}
	return nil
}

func (s *buildingService) Get(ctx context.Context, id int64) (*b.Building, error) {
	return s.buildings.GetByID(ctx, id)
}

func (s *buildingService) GetDetail(ctx context.Context, buildingID int64) (*b.BuildingDetail, error) {
	detail, err := s.buildings.GetDetail(ctx, buildingID)
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

	if err := s.validateNewFloors(in); err != nil {
		return nil, err
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

func (s *buildingService) validateNewFloors(in []b.NewFloor) error {
	numbers := make(map[int16]bool, len(in))
	captions := make(map[string]bool, len(in))
	for _, nf := range in {
		if numbers[nf.FloorNumber] {
			return b.ErrFloorNumberDuplicate
		}
		numbers[nf.FloorNumber] = true

		if nf.Caption != nil {
			if captions[*nf.Caption] {
				return b.ErrFloorCaptionDuplicate
			}
			captions[*nf.Caption] = true
		}
	}
	return nil
}

func (s *buildingService) GetFloor(ctx context.Context, buildingID int64, floorNumber int16) (*b.Floor, error) {
	return s.floors.GetByNumber(ctx, buildingID, floorNumber)
}

func (s *buildingService) GetFloorDetail(ctx context.Context, buildingID, floorID int64) (*b.FloorDetail, error) {
	return s.floors.GetDetail(ctx, buildingID, floorID)
}

func (s *buildingService) RemoveFloor(ctx context.Context, floorID int64, expectedVersion int, force bool) error {
	if force {
		return s.floors.ForceSoftDelete(ctx, floorID, expectedVersion)
	}
	return s.floors.SoftDelete(ctx, floorID, expectedVersion)
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
			Status:     b.UnitStatusVacant,
		}
		if err := units[i].SetUnitNumber(num); err != nil {
			return nil, err
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

func (s *buildingService) RemoveUnit(ctx context.Context, floorID int64, unitNumber string, force bool) error {
	u, err := s.units.GetByNumber(ctx, floorID, unitNumber)
	if err != nil {
		return err
	}
	return s.units.SoftDelete(ctx, u.ID, u.Version)
}

func NewBuildingService(
	buildings b.BuildingRepository,
	memberships b.MembershipRepository,
	floors b.FloorRepository,
	units b.UnitRepository,
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
	if tx == nil {
		panic("tx cannot be nil")
	}
	return &buildingService{
		buildings:   buildings,
		memberships: memberships,
		floors:      floors,
		units:       units,
		tx:          tx,
	}
}
