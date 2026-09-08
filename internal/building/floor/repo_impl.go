package floor

import (
	"context"
	"errors"
	"fmt"

	"unitly/internal/db"
	"unitly/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type floorRepository struct {
	db *pgxpool.Pool
}

func NewFloorRepository(pool *pgxpool.Pool) FloorRepository {
	return &floorRepository{db: pool}
}

const floorSelect = `
	SELECT id, building_id, floor_number, caption, version, created_at, updated_at
	FROM floors`

const createFloor = `
	INSERT INTO floors (building_id, floor_number, caption)
		VALUES ($1, $2, $3)
		RETURNING id, version, created_at, updated_at`

func (r *floorRepository) Create(ctx context.Context, f *Floor) error {
	q := db.GetQuerier(ctx, r.db)
	err := q.QueryRow(
		ctx, createFloor,
		f.BuildingID,
		f.FloorNumber,
		f.Caption,
	).Scan(
		&f.ID,
		&f.Version,
		&f.CreatedAt,
		&f.UpdatedAt,
	)

	return mapDBError(err)
}

const createBatch = `
	INSERT INTO floors (building_id, floor_number caption, version)
		VALUES ($1,$2,$3,1) 
		RETURNING id, created_at, updated_at`

func (r *floorRepository) CreateBatch(ctx context.Context, floors []*Floor) error {
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

func mapCreateBatchError(err error, f *Floor) error {
	err = mapDBError(err)
	switch {
	case errors.Is(err, ErrFloorNumberTaken):
		msg := fmt.Sprintf("%s. floor_number: %d", err.Error(), f.FloorNumber)
		return domain.NewError(ErrFloorNumberTaken.Code(), msg)

	case errors.Is(err, ErrFloorCaptionTaken):
		msg := fmt.Sprintf("%s. caption: %v", err.Error(), f.Caption)
		return domain.NewError(ErrFloorCaptionTaken.Code(), msg)

	default:
		return err
	}
}

func mapDBError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case db.IsUniqueViolation(err, "uq_floor_building_number_active"):
		return ErrFloorNumberTaken
	case db.IsUniqueViolation(err, "uq_floor_building_caption_active"):
		return ErrFloorCaptionTaken
	default:
		return err
	}
}
