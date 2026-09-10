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

type buildingDetailRow struct {
	BuildingGUID uuid.UUID
	BuildingName string

	FloorNumber  sql.NullInt16
	FloorCaption sql.NullString

	UnitNumber        sql.NullString
	UnitCaption       sql.NullString
	UnitStatus        sql.NullString
	UnitHasParking        sql.NullBool
	ParkingSpotNumber sql.NullString
}

func scanRow(row db.Row) (*buildingDetailRow, error) {
	var detail buildingDetailRow
	err := row.Scan(
		&detail.BuildingGUID,
		&detail.BuildingName,
		&detail.FloorNumber,
		&detail.FloorCaption,

		&detail.UnitNumber,
		&detail.UnitCaption,
		&detail.UnitStatus,
		&detail.UnitHasParking,
	)
	return &detail, err
}

const detailByID = `
SELECT 
	b.guid AS building_guid,
	b.name AS building_name,

	f.floor_number AS floor_number,
	f.caption AS floor_caption,

	u.unit_number AS unit_number,
	u.caption AS unit_caption,
	u.status AS unit_status,
	u.has_parking AS has_parking

FROM buildings AS b
LEFT JOIN floors f 
	ON f.building_id = b.id
LEFT JOIN units u
    ON u.floor_id = f.id
WHERE b.id = $1 AND b.deleted_at IS NULL
ORDER BY
	f.floor_number,
    u.id`

func (r *buildingDetailRepository) GetBuildingDetail(ctx context.Context, buildingID int64) (*building.BuildingDetail, error) {
	var (
		founded  bool
		detail   building.BuildingDetail
		floors   []*building.FloorDetail
		floorMap = make(map[int16]*building.FloorDetail)
	)

	q := db.GetQuerier(ctx, r.db)
	rows, err := q.Query(ctx, detailByID, buildingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		row, err := scanRow(rows)
		if err != nil {
			break
		}

		if !founded {
			founded = true
			detail.GUID = row.BuildingGUID
			detail.Name = row.BuildingName
		}

		if row.FloorNumber.Valid {
			floorNumber := row.FloorNumber.Int16
			floor, exist := floorMap[floorNumber]
			if !exist {
				floor = &building.FloorDetail{
					FloorNumber: floorNumber,
				}
				if row.FloorCaption.Valid {
					floor.Caption = &row.FloorCaption.String
				}
				floors = append(floors, floor)
				floorMap[floorNumber] = floor
			}

			if row.UnitNumber.Valid {
				unit := building.UnitDetail{
					UnitNumber: row.UnitNumber.String,
				}
				if row.UnitCaption.Valid {
					unit.Caption = &row.UnitCaption.String
				}
				if row.UnitStatus.Valid {
					unit.Status = building.UnitStatus(row.UnitStatus.String)
				}
				if row.UnitHasParking.Valid {
					unit.HasParking = row.UnitHasParking.Bool
				}

				floor.Units = append(floor.Units, unit)
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

	if !founded {
		return nil, domain.ErrNotFound
	}

	detail.Floors = make([]building.FloorDetail, 0, len(floors))
	for _, f := range floors {
		detail.Floors = append(detail.Floors, *f)
	}

	return &detail, nil
}
