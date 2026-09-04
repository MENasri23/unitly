package domain

import (
	"time"
	"unitly/internal/building/role"

	"github.com/google/uuid"
)

var ErrMembershipNotFound = NewNotFoundError("membership")
var ErrMembershipConflict = NewConflictError("membership")

type MembershipStatus string

const (
	MembershipStatusActive  MembershipStatus = "active"
	MembershipStatusRemoved MembershipStatus = "removed"
)

type Membership struct {
	ID         int64
	GUID       uuid.UUID
	UserID     int64
	BuildingID int64
	Role       role.Role
	Status     MembershipStatus
	InvitedBy  *int64
	Version    int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (m *Membership) HasMinLevel(level int16) bool {
	return m.Status == MembershipStatusActive && m.Role.Level >= level
}
