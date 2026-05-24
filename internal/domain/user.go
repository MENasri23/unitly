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
	CreatedAt      time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
}
