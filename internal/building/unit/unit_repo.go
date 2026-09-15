package unit

import (
	"context"
	"fmt"
	"unitly/internal/building"
	"unitly/internal/db"
	"unitly/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type unitRepository struct {
	db *pgxpool.Pool
}

const selectColumns = `
	SELECT id, building_id, floor_id, unit_number, caption, status, has_parking,
		version, created_at, updated_at, deleted_at
	FROM units`

func (r *unitRepository) scanUnit(ctx context.Context, query string, args ...any) (*building.Unit, error) {
	q := db.GetQuerier(ctx, r.db)
	var u building.Unit
	var unitNumber string
	err := q.QueryRow(ctx, query, args...).Scan(
		&u.ID, &u.BuildingID, &u.FloorID, &unitNumber, &u.Caption, &u.Status, &u.HasParking,
		&u.Version, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt,
	)
	if db.IsNoRows(err) {
		return nil, building.ErrUnitNotFound
	}
	if err != nil {
		return nil, err
	}
	if err = u.SetUnitNumber(unitNumber); err != nil {
		return nil, err
	}
	return &u, nil
}

const createBatch = `
	INSERT INTO units 
		(floor_id, unit_number, caption, status, has_parking)
		VALUES ($1, $2, $3, $4, $5)
	RETURNING id, building_id, version, created_at, updated_at`

func (r *unitRepository) CreateBatch(ctx context.Context, units []*building.Unit) error {
	if len(units) == 0 {
		return nil
	}

	q := db.GetQuerier(ctx, r.db)

	batch := &pgx.Batch{}
	for _, u := range units {
		batch.Queue(
			createBatch,
			u.FloorID,
			u.UnitNumber(),
			u.Caption,
			u.Status,
			u.HasParking,
		)
	}

	br := q.SendBatch(ctx, batch)
	defer br.Close()

	for _, u := range units {
		err := br.QueryRow().Scan(&u.ID, &u.BuildingID, &u.Version, &u.CreatedAt, &u.UpdatedAt)
		if db.IsUniqueViolation(err, "uq_unit_floor_number_active") {
			return fmt.Errorf("unit number %q: %w", u.UnitNumber(), building.ErrUnitNumberTaken)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

const getByID = selectColumns + ` WHERE floor_id = $1 AND unit_number = $2 AND deleted_at IS NULL`

func (r *unitRepository) GetByNumber(ctx context.Context, floorID int64, unitNumber string) (*building.Unit, error) {
	return r.scanUnit(ctx, getByID, floorID, unitNumber)
}

const softDelete = `
	UPDATE units
	SET
		deleted_at = now(),
		version = version + 1
	WHERE
		id = $1
		AND version = $2
		AND deleted_at IS NULL`

func (r *unitRepository) SoftDelete(ctx context.Context, id int64, expectedVersion int) error {
	q := db.GetQuerier(ctx, r.db)
	tag, err := q.Exec(ctx, softDelete, id, expectedVersion)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return nil
	}

	return r.diagnoseDeleteFailure(ctx, id, expectedVersion)
}

const diagnoseDelete = `SELECT version FROM units WHERE id = $1 AND deleted_at IS NULL`

func (r *unitRepository) diagnoseDeleteFailure(ctx context.Context, id int64, expectedVersion int) error {
	q := db.GetQuerier(ctx, r.db)
	var version int
	err := q.QueryRow(ctx, diagnoseDelete, id).Scan(&version)
	if db.IsNoRows(err) {
		return building.ErrUnitNotFound
	}
	if err != nil {
		return err
	}
	if version != expectedVersion {
		return domain.ErrVersionConflict
	}
	return nil
}

func NewUnitRepository(db *pgxpool.Pool) building.UnitRepository {
	return &unitRepository{db: db}
}
