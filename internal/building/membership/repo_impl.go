package membership

import (
	"context"
	"errors"
	"unitly/internal/building/role"
	"unitly/internal/db"
	"unitly/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type membershipRepository struct {
	db    *pgxpool.Pool
	cache domain.Cache
}

func (r *membershipRepository) getQuerier(ctx context.Context) db.Querier {
	if tx := db.GetTxOrNil(ctx); tx != nil {
		return tx
	}

	return r.db
}

const membershipSelect = `
	SELECT m.id, m.guid, m.user_id, m.building_id, m.status, m.invited_by, m.version, 
			m.created_at, m.updated_at, r.id, r.name, r.level
	FROM building_memberships m JOIN building_roles r ON r.id = m.role_id`

func scanMembership(row db.Row) (*domain.Membership, error) {
	var m domain.Membership
	var role role.Role
	err := row.Scan(&m.ID, &m.GUID, &m.UserID, &m.BuildingID, &m.Status, &m.InvitedBy, &m.Version,
		&m.CreatedAt, &m.UpdatedAt, &role.ID, &role.Name, &role.Level)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMembershipNotFound
		}
		return nil, err
	}
	m.Role = role
	return &m, nil
}

const createMembership = `
	INSERT INTO building_memberships (guid, user_id, building_id, role_id, status, invited_by, version)
	VALUES ($1,$2,$3,$4,$5,$6,1)
	RETURNING id, created_at, updated_at`

func (r *membershipRepository) Create(ctx context.Context, m *domain.Membership) error {
	return r.getQuerier(ctx).QueryRow(ctx, createMembership,
		m.GUID, m.UserID, m.BuildingID, m.Role.ID, m.Status, m.InvitedBy,
	).Scan(&m.ID, &m.CreatedAt, &m.UpdatedAt)
}

const getByGUID = membershipSelect + ` WHERE m.guid=$1`

func (r *membershipRepository) GetByGUID(ctx context.Context, guid uuid.UUID) (*domain.Membership, error) {
	return scanMembership(r.getQuerier(ctx).QueryRow(ctx, getByGUID, guid))
}

const getByID = membershipSelect + ` WHERE m.id=$1`

func (r *membershipRepository) GetByID(ctx context.Context, id int64) (*domain.Membership, error) {
	return scanMembership(r.getQuerier(ctx).QueryRow(ctx, getByID, id))
}

const getByUserBuildingRole = membershipSelect + ` WHERE m.user_id=$1 AND m.building_id=$2 AND m.role_id=$3`

func (r *membershipRepository) GetByUserBuildingRole(ctx context.Context, userID int64, buildingID int64, roleID int16) (*domain.Membership, error) {
	row := r.getQuerier(ctx).QueryRow(ctx, getByUserBuildingRole, userID, buildingID, roleID)
	return scanMembership(row)
}

const getHighestActiveByUserBuilding = membershipSelect + `
	WHERE m.user_id=$1 AND m.building_id=$2 AND m.status='active'
	ORDER BY r.level DESC LIMIT 1`

func (r *membershipRepository) GetHighestActiveByUserBuilding(ctx context.Context, userID int64, buildingID int64) (*domain.Membership, error) {
	row := r.getQuerier(ctx).QueryRow(ctx, getHighestActiveByUserBuilding, userID, buildingID)
	return scanMembership(row)
}

const listByBuilding = membershipSelect + ` WHERE m.building_id=$1`

func (r *membershipRepository) ListByBuilding(ctx context.Context, buildingID int64, f MembershipFilter) ([]*domain.Membership, error) {
	qb := db.NewQueryBuilder(listByBuilding, buildingID)
	if f.RoleName != nil {
		qb.And("r.name", *f.RoleName)
	}
	if f.Status != nil {
		qb.And("m.status", *f.Status)
	}
	sql, args := qb.OrderBy("m.created_at").Build()

	rows, err := r.db.Query(ctx, sql, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.Membership
	for rows.Next() {
		m, err := scanMembership(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

const updateStatus = `
	UPDATE building_memberships 
	SET 
		status=$1,
		version=version+1 
	WHERE 
		id=$2 AND version=$3`

func (r *membershipRepository) UpdateStatus(ctx context.Context, id int64, status domain.MembershipStatus, version int) error {
	ct, err := r.getQuerier(ctx).Exec(ctx, updateStatus, status, id, version)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrMembershipConflict
	}
	return nil
}

func NewMembershipRepository(db *pgxpool.Pool, cache domain.Cache) MembershipRepository {
	if db == nil {
		panic("db cannot be nil")
	}
	if cache == nil {
		panic("cache cannot be nil")
	}
	return &membershipRepository{
		db:    db,
		cache: cache,
	}
}
