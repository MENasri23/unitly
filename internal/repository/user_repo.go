package repository

import (
	"context"
	"database/sql"
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

const createUser = `
INSERT INTO users (phone, public_id, active) 
VALUES ($1, $2, $3)
ON CONFLICT (phone) DO NOTHING
RETURNING id, created_at, version
`

func (u *userRepository) Create(ctx context.Context, user *domain.User) error {
	querier := u.getQuerier(ctx)
	err := querier.QueryRow(
		ctx,
		createUser,
		user.Phone,
		user.GUID,
		user.IsActive,
	).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Version,
	)

	if err == pgx.ErrNoRows {
		return domain.ErrAlreadyExist
	}

	return err
}

func (u *userRepository) GetByID(ctx context.Context, userID int64) (*domain.User, error) {
	return u.getBy(ctx, "id", userID)
}

func (u *userRepository) GetByGUID(ctx context.Context, guid uuid.UUID) (*domain.User, error) {
	return u.getBy(ctx, "public_id", guid)
}

func (u *userRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	return u.getBy(ctx, "phone", phone)
}

func (u *userRepository) getBy(ctx context.Context, column string, value any) (*domain.User, error) {
	query := fmt.Sprintf(`
	SELECT id, public_id, phone, first_name, last_name, active, password_hash, created_at, version
	FROM users
	WHERE %s = $1
	`, column,
	)

	querier := u.getQuerier(ctx)

	var user domain.User
	var passwordHash sql.NullString
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
		&passwordHash,
		&user.CreatedAt,
		&user.Version,
	)

	if passwordHash.Valid {
		user.HashedPassword = &passwordHash.String
	}

	if err != nil {
		return nil, mapDBError(err)
	}

	return &user, nil
}

const updateUserPassword = `
UPDATE users 
SET password_hash = $1, version = version + 1
WHERE id = $2 AND version = $3
`

func (u *userRepository) UpdatePassword(ctx context.Context, user *domain.User, password string) error {
	passwordHash := hash.SHA256(password)
	querier := u.getQuerier(ctx)

	cmdTag, err := querier.Exec(ctx, updateUserPassword, passwordHash, user.ID, user.Version)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

const existsUserByPhone = `
SELECT EXISTS(
    SELECT 1 FROM users WHERE phone = $1
)
`

func (u *userRepository) ExistsByPhone(ctx context.Context, phone string) (bool, error) {
	querier := u.getQuerier(ctx)

	var exists bool
	err := querier.QueryRow(ctx, existsUserByPhone, phone).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

const updateUserFullName = `
UPDATE users
SET
    first_name = COALESCE(NULLIF($1, ''), first_name),
    last_name  = COALESCE(NULLIF($2, ''), last_name),
	version = version + 1
WHERE id = $3 AND version = $4
`

func (u *userRepository) UpdateFullName(ctx context.Context, user *domain.User) error {
	querier := u.getQuerier(ctx)
	cmdTag, err := querier.Exec(
		ctx,
		updateUserFullName,
		user.FirstName,
		user.LastName,
		user.ID,
		user.Version,
	)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

const updatePhone = `
UPDATE users
SET phone = $1, version = version + 1
WHERE id = $2 AND version = $3
`

func (u *userRepository) UpdatePhone(ctx context.Context, user *domain.User) error {
	querier := u.getQuerier(ctx)
	cmdTag, err := querier.Exec(ctx, updatePhone, user.Phone, user.ID, user.Version)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func mapDBError(err error) error {
	if err == nil {
		return nil
	}

	switch err {
	case pgx.ErrNoRows:
		return domain.ErrUserNotFound
	default:
		return err
	}
}
