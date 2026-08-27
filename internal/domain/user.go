package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

var ErrUserNotFound = NewNotFoundError("user")

type User struct {
	ID             int64
	GUID           uuid.UUID
	Phone          string
	FirstName      *string
	LastName       *string
	HashedPassword *string
	IsActive       bool
	Version        int
	CreatedAt      time.Time
}

func (u *User) IsPasswordSet() bool {
	return u.HashedPassword != nil && *u.HashedPassword != ""
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, userID int64) (*User, error)
	GetByPhone(ctx context.Context, phone string) (*User, error)
	GetByGUID(ctx context.Context, guid uuid.UUID) (*User, error)
	GetIDByGUID(ctx context.Context, guid uuid.UUID) (int64, error)
	ExistsByPhone(ctx context.Context, phone string) (bool, error)
	UpdatePassword(ctx context.Context, user *User, password string) error
	UpdateFullName(ctx context.Context, user *User) error
	UpdatePhone(ctx context.Context, user *User) error
}
