package floor

import (
	"database/sql"
	"unitly/internal/building"
	"unitly/internal/db"
)

// floorUnitRow represents the common columns shared between building and floor detail queries.
type FloorUnitRow struct {
	FloorNumber    sql.NullInt16
	FloorCaption   sql.NullString
	UnitNumber     sql.NullString
	UnitCaption    sql.NullString
	UnitStatus     sql.NullString
}

func (r *FloorUnitRow) ToUnitDetail() (building.UnitSummary, bool) {
	if !r.UnitNumber.Valid {
		return building.UnitSummary{}, false
	}
	return building.UnitSummary{
		UnitNumber: r.UnitNumber.String,
		Caption:    db.NullString(r.UnitCaption),
		Status:     building.UnitStatus(r.UnitStatus.String),
	}, true
}
