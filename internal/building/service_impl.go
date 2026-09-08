package building

import (
	"context"
	"errors"
	"unitly/internal/building/detail"
	"unitly/internal/building/floor"
	"unitly/internal/building/membership"
	"unitly/internal/domain"

	"github.com/google/uuid"
)

type buildingService struct {
	buildings   BuildingRepository
	memberships membership.MembershipRepository
	floors      floor.FloorRepository
	details     detail.BuildingDetailRepository
	tx          domain.UnitOfWork
}

func (s *buildingService) Create(ctx context.Context, actorUserID int64, in CreateBuildingInput) (*Building, error) {
	buildingID, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	b := &Building{
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

func (s *buildingService) Get(ctx context.Context, id int64) (*Building, error) {
	return s.buildings.GetByID(ctx, id)
}

func (s *buildingService) GetDetail(ctx context.Context, buildingID int64) (*detail.BuildingDetail, error) {
	detail, err := s.details.GetBuildingDetail(ctx, buildingID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrBuildingNotFound
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

func (s *buildingService) ListMine(ctx context.Context, userID int64) ([]*BuildingSummary, error) {
	return s.buildings.ListSummaryByUser(ctx, userID)
}

func (s *buildingService) Delete(ctx context.Context, buildingID int64) error {
	return s.buildings.SoftDelete(ctx, buildingID) // authz (admin-only) enforced by middleware
}

func (s *buildingService) AddFloor(ctx context.Context, buildingID int64, floorNumber int16, caption *string) (*floor.Floor, error) {
	out, err := s.AddFloors(ctx, buildingID, []NewFloor{{FloorNumber: floorNumber, Caption: caption}})
	if err != nil {
		return nil, err
	}
	return out[0], nil
}

func (s *buildingService) AddFloors(ctx context.Context, buildingID int64, in []NewFloor) ([]*floor.Floor, error) {
	if len(in) == 0 {
		return nil, floor.ErrNoFloorsProvided
	}

	floors := make([]*floor.Floor, len(in))
	for i, nf := range in {
		floors[i] = &floor.Floor{
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

	return floors, nil
}

func NewBuildingService(
	buildings BuildingRepository,
	memberships membership.MembershipRepository,
	floors floor.FloorRepository,
	details detail.BuildingDetailRepository,
	tx domain.UnitOfWork,
) BuildingService {
	if buildings == nil {
		panic("buildings cannot be nil")
	}
	if memberships == nil {
		panic("memberships cannot be nil")
	}
	if floors == nil {
		panic("floors cannot be nil")
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
		details:     details,
		tx:          tx,
	}
}
