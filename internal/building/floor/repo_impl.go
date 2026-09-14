package floor

import (
	"context"
	"errors"
	"fmt"

	"unitly/internal/building"
	"unitly/internal/db"
	"unitly/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type floorRepository struct {
	db *pgxpool.Pool
}

func NewFloorRepository(pool *pgxpool.Pool) building.FloorRepository {
	return &floorRepository{db: pool}
}

const floorSelect = `
	SELECT id, building_id, floor_number, caption, version, created_at, updated_at, deleted_at
	FROM floors`

const createBatch = `
	INSERT INTO floors (building_id, floor_number, caption, version)
		VALUES ($1,$2,$3,1) 
		RETURNING id, created_at, updated_at`

func (r *floorRepository) CreateBatch(ctx context.Context, floors []*building.Floor) error {
	if len(floors) == 0 {
		return nil
	}
	q := db.GetQuerier(ctx, r.db)

	batch := &pgx.Batch{}
	for _, f := range floors {
		batch.Queue(createBatch, f.BuildingID, f.FloorNumber, f.Caption)
	}

	br := q.SendBatch(ctx, batch)
	defer br.Close()

	for _, f := range floors {
		if err := br.QueryRow().Scan(&f.ID, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return mapCreateBatchError(err, f)
		}
	}

	return nil
}

func mapCreateBatchError(err error, f *building.Floor) error {
	err = mapDBError(err)
	switch {
	case errors.Is(err, building.ErrFloorNumberTaken):
		return fmt.Errorf("floor number %d: %w", f.FloorNumber, building.ErrFloorNumberTaken)

	case errors.Is(err, building.ErrFloorCaptionTaken):
		return fmt.Errorf("floor caption '%s': %w", *f.Caption, building.ErrFloorCaptionTaken)

	default:
		return err
	}
}

const getByNumber = floorSelect + ` WHERE building_id = $1 AND floor_number = $2 AND deleted_at IS NULL`

func (r *floorRepository) GetByNumber(ctx context.Context, buildingID int64, floorNumber int16) (*building.Floor, error) {
	q := db.GetQuerier(ctx, r.db)
	var f building.Floor
	err := q.QueryRow(ctx, getByNumber, buildingID, floorNumber).Scan(
		&f.ID,
		&f.BuildingID,
		&f.FloorNumber,
		&f.Caption,
		&f.Version,
		&f.CreatedAt,
		&f.UpdatedAt,
		&f.DeletedAt,
	)

	if err != nil {
		return nil, mapDBError(err)
	}

	return &f, nil
}

const floorDetail = `
SELECT
    f.floor_number,
    f.caption,

    u.unit_number,
    u.caption,
    u.status,
    u.has_parking
FROM floors f
LEFT JOIN units u
    ON u.floor_id = f.id
    AND u.deleted_at IS NULL
WHERE f.building_id = $1 AND f.id = $2
ORDER BY u.id`

func (r *floorRepository) GetDetail(ctx context.Context, buildingID, floorID int64) (*building.FloorDetail, error) {
	q := db.GetQuerier(ctx, r.db)
	rows, err := q.Query(ctx, floorDetail, buildingID, floorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		found  bool
		detail building.FloorDetail
	)

	for rows.Next() {
		var row FloorUnitRow
		if err := rows.Scan(
			&row.FloorNumber,
			&row.FloorCaption,
			&row.UnitNumber,
			&row.UnitCaption,
			&row.UnitStatus,
			&row.UnitHasParking,
		); err != nil {
			return nil, err
		}

		if !found && row.FloorNumber.Valid {
			found = true
			detail.FloorNumber = row.FloorNumber.Int16
			detail.Caption = db.NullString(row.FloorCaption)
		}

		if unit, ok := row.ToUnitDetail(); ok {
			detail.Units = append(detail.Units, unit)
		}
	}

	if err := rows.Err(); err != nil {
		if db.IsNoRows(err) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	if !found {
		return nil, domain.ErrNotFound
	}

	return &detail, nil
}

const softDelete = `
	UPDATE floors SET deleted_at = now(), version = version + 1
	WHERE id = $1 AND version = $2 AND deleted_at IS NULL
		AND NOT EXISTS (
			SELECT 1 FROM units WHERE floor_id = $1 AND deleted_at IS NULL
		)`

func (r *floorRepository) SoftDelete(ctx context.Context, id int64, expectedVersion int) error {
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

const getVersion = `SELECT version FROM floors WHERE id = $1 AND deleted_at IS NULL`

func (r *floorRepository) diagnoseDeleteFailure(ctx context.Context, floorID int64, expectedVersion int) error {
	q := db.GetQuerier(ctx, r.db)
	var version int
	err := q.QueryRow(ctx, getVersion, floorID).Scan(&version)
	if db.IsNoRows(err) {
		return building.ErrFloorNotFound
	}
	if err != nil {
		return err
	}
	if version != expectedVersion {
		return domain.ErrVersionConflict
	}
	return building.ErrFloorHasUnits
}

func mapDBError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case db.IsNoRows(err):
		return building.ErrFloorNotFound
	case db.IsUniqueViolation(err, "uq_floor_building_number_active"):
		return building.ErrFloorNumberTaken
	case db.IsUniqueViolation(err, "uq_floor_building_caption_active"):
		return building.ErrFloorCaptionTaken
	default:
		return err
	}
}
