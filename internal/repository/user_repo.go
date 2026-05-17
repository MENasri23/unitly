package repository

import (
	"context"
	"unitly/internal/db"
	"unitly/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (u *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
	INSERT INTO users (phone, public_id, password_hash) VALUES ($1, $2, $3) 
	RETURNING id, created_at
	`

	querier := u.getQuerier(ctx)
	return querier.QueryRow(
		ctx,
		query,
		user.Phone,
		user.GUID,
		user.HashedPassword,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)
}

func (u *userRepository) getQuerier(ctx context.Context) db.Querier {
	if tx := db.GetTxOrNil(ctx); tx != nil {
		return tx
	}

	return u.db
}
