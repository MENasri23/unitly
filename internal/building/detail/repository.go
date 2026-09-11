package bldgdetail

import (
	"context"
	"database/sql"
	"unitly/internal/building"
	"unitly/internal/db"
	"unitly/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type buildingDetailRepository struct {
	db *pgxpool.Pool
}

func NewBuildingDetailRepository(db *pgxpool.Pool) building.BuildingDetailRepository {
	return &buildingDetailRepository{db: db}
}

// floorUnitRow represents the common columns shared between building and floor detail queries.
type floorUnitRow struct {
	FloorNumber    sql.NullInt16
	FloorCaption   sql.NullString
	UnitNumber     sql.NullString
	UnitCaption    sql.NullString
	UnitStatus     sql.NullString
	UnitHasParking sql.NullBool
}

func (r *floorUnitRow) toUnitDetail() (building.UnitDetail, bool) {
	if !r.UnitNumber.Valid {
		return building.UnitDetail{}, false
	}
	return building.UnitDetail{
		UnitNumber: r.UnitNumber.String,
		Caption:    db.NullString(r.UnitCaption),
		Status:     building.UnitStatus(r.UnitStatus.String),
		HasParking: db.Bool(r.UnitHasParking),
	}, true

}

const buildingDetailByID = `
SELECT 
    b.guid,
    b.name,

    f.floor_number,
    f.caption,
    
	u.unit_number,
    u.caption,
    u.status,
    u.has_parking
FROM buildings b
LEFT JOIN floors f 
    ON f.building_id = b.id
LEFT JOIN units u 
    ON u.floor_id = f.id 
    AND u.deleted_at IS NULL
WHERE b.id = $1
ORDER BY f.floor_number, u.id`

func (r *buildingDetailRepository) GetBuildingDetail(ctx context.Context, buildingID int64) (*building.BuildingDetail, error) {
	q := db.GetQuerier(ctx, r.db)
	rows, err := q.Query(ctx, buildingDetailByID, buildingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		found      bool
		detail     building.BuildingDetail
		floorIndex = make(map[int16]int) // maps floorNumber to its index in detail.Floors
	)

	for rows.Next() {
		var (
			bGUID uuid.UUID
			bName string
			row   floorUnitRow
		)

		if err := rows.Scan(
			&bGUID,
			&bName,
			&row.FloorNumber,
			&row.FloorCaption,
			&row.UnitNumber,
			&row.UnitCaption,
			&row.UnitStatus,
			&row.UnitHasParking,
		); err != nil {
			return nil, err
		}

		if !found {
			found = true
			detail.GUID = bGUID
			detail.Name = bName
		}

		if row.FloorNumber.Valid {
			fNum := row.FloorNumber.Int16
			idx, exists := floorIndex[fNum]
			if !exists {
				idx = len(detail.Floors)
				detail.Floors = append(detail.Floors, building.FloorDetail{
					FloorNumber: fNum,
					Caption:     db.NullString(row.FloorCaption),
				})
				floorIndex[fNum] = idx
			}

			if unit, ok := row.toUnitDetail(); ok {
				detail.Floors[idx].Units = append(detail.Floors[idx].Units, unit)
				detail.TotalUnits++
				if unit.Status == building.UnitStatusOccupied {
					detail.OccupiedUnits++
				}
			}
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

func (r *buildingDetailRepository) GetFloorDetail(ctx context.Context, buildingID, floorID int64) (*building.FloorDetail, error) {
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
		var row floorUnitRow
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

		if unit, ok := row.toUnitDetail(); ok {
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
