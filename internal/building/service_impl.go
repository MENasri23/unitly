package building

import (
	"context"
	"unitly/internal/building/membership"
	"unitly/internal/domain"

	"github.com/google/uuid"
)

type buildingService struct {
	buildings   BuildingRepository
	memberships membership.MembershipRepository
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

func (s *buildingService) GetInternalID(ctx context.Context, buildingID uuid.UUID) (int64, error) {
	return s.buildings.GetInternalID(ctx, buildingID)
}

func (s *buildingService) ListMine(ctx context.Context, userID int64) ([]*BuildingSummary, error) {
	return s.buildings.ListSummaryByUser(ctx, userID)
}
func (s *buildingService) Delete(ctx context.Context, buildingID int64) error {
	return s.buildings.SoftDelete(ctx, buildingID) // authz (admin-only) enforced by middleware
}

func NewBuildingService(buildings BuildingRepository, memberships membership.MembershipRepository, tx domain.UnitOfWork) BuildingService {
	if buildings == nil {
		panic("buildings cannot be nil")
	}
	if memberships == nil {
		panic("memberships cannot be nil")
	}
	if tx == nil {
		panic("tx cannot be nil")
	}
	return &buildingService{
		buildings:   buildings,
		memberships: memberships,
		tx:          tx,
	}
}
