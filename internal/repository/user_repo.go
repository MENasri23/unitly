package repository

import (
	"context"
	"fmt"
	"unitly/internal/db"
	"unitly/internal/domain"
	"unitly/internal/util/hash"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) domain.UserRepository {
	if db == nil {
		panic("db cannot be nil")
	}

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
	INSERT INTO users (phone, public_id, active) 
	VALUES ($1, $2, $3)
	ON CONFLICT (phone) DO NOTHING
	RETURNING id, created_at
	`

	querier := u.getQuerier(ctx)
	err := querier.QueryRow(
		ctx,
		query,
		user.Phone,
		user.GUID,
		user.IsActive,
	).Scan(
		&user.ID,
		&user.CreatedAt,
	)

	if err == pgx.ErrNoRows {
		return domain.ErrAlreadyExist
	}

	return err
}

func (u *userRepository) GetByID(ctx context.Context, userID int64) (*domain.User, error) {
	return u.getBy(ctx, "id", userID)
}

func (u *userRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	return u.getBy(ctx, "phone", phone)
}

func (u *userRepository) getBy(ctx context.Context, column string, value any) (*domain.User, error) {
	query := fmt.Sprintf(`
	SELECT id, public_id, phone, first_name, last_name, active, created_at
	FROM users
	WHERE %s = $1
	`, column,
	)

	querier := u.getQuerier(ctx)

	var user domain.User
	err := querier.QueryRow(
		ctx,
		query,
		value,
	).Scan(
		&user.ID,
		&user.GUID,
		&user.Phone,
		&user.FirstName,
		&user.LastName,
		&user.IsActive,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, mapDBError(err)
	}

	return &user, nil
}

func (u *userRepository) GetUserIDByPhone(ctx context.Context, phone string) (int64, error) {
	return u.getUserIDBy(ctx, phone, false)
}

func (u *userRepository) GetActiveUserIDByPhone(ctx context.Context, phone string) (int64, error) {
	return u.getUserIDBy(ctx, phone, true)
}

func (u *userRepository) getUserIDBy(ctx context.Context, phone string, isActive bool) (int64, error) {
	query := "SELECT id from users WHERE phone = $1 AND active = $2"

	querier := u.getQuerier(ctx)

	var userID int64
	err := querier.QueryRow(ctx, query, phone, isActive).Scan(&userID)
	if err != nil {
		return -1, mapDBError(err)
	}

	return userID, nil
}

func (u *userRepository) UpdatePassword(ctx context.Context, userGUID uuid.UUID, password string) error {
	query := `UPDATE users 
	SET password_hash = $1 
	WHERE public_id = $2`

	passwordHash := hash.SHA256(password)
	querier := u.getQuerier(ctx)

	cmdTag, err := querier.Exec(ctx, query, passwordHash, userGUID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func mapDBError(err error) error {
	if err == nil {
		return nil
	}

	switch err {
	case pgx.ErrNoRows:
		return domain.ErrNotFound
	default:
		return err
	}
}
