package membership

import (
	"context"
	"unitly/internal/building/role"

	"github.com/google/uuid"
)

type MembershipFilter struct {
	RoleName *role.RoleName
	Status   *MembershipStatus
}

// MembershipRepository manages a user's role-scoped access to a building.
// A user can hold more than one row per building (e.g. owner of one unit,
// tenant of another) — uniqueness is on (user_id, building_id, role_id).
type MembershipRepository interface {
	Create(ctx context.Context, m *Membership) error
	GetByID(ctx context.Context, id int64) (*Membership, error)
	GetByGUID(ctx context.Context, guid uuid.UUID) (*Membership, error)
	GetByUserBuildingRole(ctx context.Context, userID, buildingID int64, roleID int16) (*Membership, error)

	// GetHighestActiveByUserBuilding returns the caller's most-privileged active
	// row for a building.
	GetHighestActiveByUserBuilding(ctx context.Context, userID, buildingID int64) (*Membership, error)

	ListByBuilding(ctx context.Context, buildingID int64, filter MembershipFilter) ([]*Membership, error)

	UpdateStatus(ctx context.Context, id int64, status MembershipStatus, version int) error
}
