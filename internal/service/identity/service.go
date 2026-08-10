package identity

import (
	"context"
	"unitly/internal/domain"

	"github.com/google/uuid"
)

var ErrEmptyFullName = domain.NewError("EMPTY_FULLNAME", "first name and last name cannot both be empty")

// IdentityService manages user accounts, credentials, and profile state.
type IdentityService interface {
	GetByPhone(ctx context.Context, phone string) (*domain.User, error)
	HasPassword(ctx context.Context, phone string) (bool, error)
	GetOrCreateUser(ctx context.Context, phone string) (*domain.User, error)

	Authenticate(ctx context.Context, phone, password string) (*domain.User, error)

	ResetPassword(ctx context.Context, userID uuid.UUID, newPassword string) error
	UpdatePassword(ctx context.Context, user *domain.User, sessionID uuid.UUID, oldPassword, newPassword string) error
	UpdateFullName(ctx context.Context, user *domain.User) error
}
