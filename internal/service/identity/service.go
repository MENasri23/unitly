package identitysvc

import (
	"context"
	"github.com/google/uuid"
	"unitly/internal/domain"
)

// IdentityService manages user accounts, credentials, and profile state.
type IdentityService interface {
	GetByPhone(ctx context.Context, phone string) (*domain.User, error)
	HasPassword(ctx context.Context, phone string) (bool, error)
	GetOrCreateUser(ctx context.Context, phone string) (*domain.User, error)

	Authenticate(ctx context.Context, phone, password string) (*domain.User, error)

	UpdatePassword(ctx context.Context, userID uuid.UUID, newPassword string) error
}
