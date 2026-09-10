package unit

import (
	"context"
	"fmt"
	"unitly/internal/building"
	"unitly/internal/db"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type unitRepository struct {
	db *pgxpool.Pool
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
			u.UnitNumber,
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
			return fmt.Errorf("unit number %q: %w", u.UnitNumber, building.ErrUnitNumberTaken)
		}
		if err != nil {
			return err
		}
	}
	
	return nil
}

func NewUnitRepository(db *pgxpool.Pool) building.UnitRepository {
	return &unitRepository{db: db}
}
