package repository

import (
	"context"
	"errors"
	"unitly/internal/db"
	"unitly/internal/domain"

	"github.com/jackc/pgx/v5"
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

func (u *userRepository) getQuerier(ctx context.Context) db.Querier {
	if tx := db.GetTxOrNil(ctx); tx != nil {
		return tx
	}

	return u.db
}

func (u *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
	INSERT INTO users (phone, public_id) VALUES ($1, $2) 
	RETURNING id, created_at
	`

	querier := u.getQuerier(ctx)
	return querier.QueryRow(
		ctx,
		query,
		user.Phone,
		user.GUID,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)
}

func (u *userRepository) GetByID(ctx context.Context, userID int64) (*domain.User, error) {
	query := `
	SELECT id, public_id, phone, first_name, last_name, created_at
	FROM users
	WHERE id = $1
	`

	querier := u.getQuerier(ctx)

	var user domain.User
	err := querier.QueryRow(
		ctx,
		query,
		userID,
	).Scan(
		&user.ID,
		&user.GUID,
		&user.Phone,
		&user.FirstName,
		&user.LastName,
		&user.CreatedAt,
	)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, domain.ErrNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
