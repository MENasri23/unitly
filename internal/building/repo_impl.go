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

const columns = "id, guid, name, address, created_by, version, created_at, updated_at"

const buildingSelect = `
	SELECT ` + columns + `
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

const getInternalID = `SELECT id FROM buildings WHERE guid=$1 AND deleted_at IS NULL`

func (r *buildingRepository) GetInternalID(ctx context.Context, guid uuid.UUID) (int64, error) {
	row := r.getQuerier(ctx).QueryRow(ctx, getInternalID, guid)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrBuildingNotFound
		}
		return 0, err
	}
	return id, nil
}

const getSummaryByID = `
		SELECT DISTINCT b.guid, b.name
		FROM buildings b
		JOIN building_memberships m ON m.building_id = b.id
		WHERE m.user_id = $1 AND m.status = 'active' AND b.deleted_at IS NULL
		ORDER BY b.created_at DESC`

func (r *buildingRepository) ListSummaryByUser(ctx context.Context, userID int64) ([]*BuildingSummary, error) {
	rows, err := r.getQuerier(ctx).Query(ctx, getSummaryByID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*BuildingSummary
	for rows.Next() {
		var summary BuildingSummary
		err := rows.Scan(&summary.ID, &summary.Name)
		if err != nil {
			return nil, err
		}
		out = append(out, &summary)
	}
	return out, rows.Err()
}

const listByUser = `
		SELECT DISTINCT ` + columns + `
		FROM buildings b
		JOIN building_memberships m ON m.building_id = b.id
		WHERE m.user_id = $1 AND m.status = 'active' AND b.deleted_at IS NULL
		ORDER BY b.created_at DESC`

// ListByUser: DISTINCT because a user can hold more than one active
// membership row (e.g. owner + tenant) in the same building.
func (r *buildingRepository) ListByUser(ctx context.Context, userID int64) ([]*Building, error) {
	rows, err := r.getQuerier(ctx).Query(ctx, listByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Building
	for rows.Next() {
		b, err := scanBuilding(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

const softDelete = `UPDATE buildings SET deleted_at=now() WHERE id=$1 AND deleted_at IS NULL`

func (r *buildingRepository) SoftDelete(ctx context.Context, id int64) error {
	ct, err := r.getQuerier(ctx).Exec(ctx, softDelete, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrBuildingNotFound
	}
	return nil
}
