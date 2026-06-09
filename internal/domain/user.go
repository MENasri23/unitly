package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             int64
	GUID           uuid.UUID
	Phone          string
	FirstName      *string
	LastName       *string
	HashedPassword *string
	IsActive       bool
	CreatedAt      time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, userID int64) (*User, error)
	GetUserIDByPhone(ctx context.Context, phone string) (int64, error)
}
