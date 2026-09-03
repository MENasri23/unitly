package building

import (
	"context"
	"errors"
	"unitly/internal/db"
	"unitly/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type buildingRepository struct {
	db    *pgxpool.Pool
	cache domain.Cache
}

func NewBuildingRepository(db *pgxpool.Pool, cache domain.Cache) BuildingRepository {
	if db == nil {
		panic("db cannot be nil")
	}
	if cache == nil {
		panic("cache cannot be nil")
	}
	return &buildingRepository{
		db:    db,
		cache: cache,
	}
}

const buildingSelect = `
	SELECT id, guid, name, address,
			created_by, version, created_at, updated_at
	FROM buildings`

func scanBuilding(row db.Row) (*Building, error) {
	var b Building
	err := row.Scan(&b.ID, &b.GUID, &b.Name, &b.Address,
		&b.CreatedBy, &b.Version, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrBuildingNotFound
		}
		return nil, err
	}
	return &b, nil
}

func (r *buildingRepository) getQuerier(ctx context.Context) db.Querier {
	if tx := db.GetTxOrNil(ctx); tx != nil {
		return tx
	}

	return r.db
}

const createBuilding = `
INSERT INTO buildings (guid, name, address, created_by)
VALUES ($1, $2, $3, $4)
RETURNING id, version, created_at, updated_at
`

// Create implements [BuildingRepository].
func (r *buildingRepository) Create(ctx context.Context, b *Building) error {
	querier := r.getQuerier(ctx)
	return querier.QueryRow(
		ctx,
		createBuilding,
		b.GUID,
		b.Name,
		b.Address,
		b.CreatedBy,
	).Scan(
		&b.ID,
		&b.Version,
		&b.CreatedAt,
		&b.UpdatedAt,
	)
}

const getByID = buildingSelect + ` WHERE id=$1 AND deleted_at IS NULL`

func (r *buildingRepository) GetByID(ctx context.Context, id int64) (*Building, error) {
	return scanBuilding(r.getQuerier(ctx).QueryRow(ctx, getByID, id))
}

const getByGUID = buildingSelect + ` WHERE guid=$1 AND deleted_at IS NULL`

func (r *buildingRepository) GetByGUID(ctx context.Context, guid uuid.UUID) (*Building, error) {
	return scanBuilding(r.getQuerier(ctx).QueryRow(ctx, getByGUID, guid))
}

const softDelete = `UPDATE buildings SET deleted_at=now() WHERE guid=$1 AND deleted_at IS NULL`

func (r *buildingRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	ct, err := r.getQuerier(ctx).Exec(ctx, softDelete, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrBuildingNotFound
	}
	return nil
}
