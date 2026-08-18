package repository

import (
	"context"
	"errors"
	"time"
	"unitly/internal/db"
	"unitly/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sessionRepository struct {
	db *pgxpool.Pool
}

func NewSessionRepository(db *pgxpool.Pool) domain.SessionRepository {
	if db == nil {
		panic("db cannot be nil")
	}

	return &sessionRepository{db: db}
}

func (s *sessionRepository) getQuerier(ctx context.Context) db.Querier {
	if tx := db.GetTxOrNil(ctx); tx != nil {
		return tx
	}

	return s.db
}

const sessionColumns = `
id, user_id, family_id, parent_session_id, refresh_token_hash,
ip_address, user_agent, device_type, os_name, os_version, 
label, created_at, last_used_at, expires_at, revoked_at`

const createSession = `
INSERT INTO sessions (` + sessionColumns + `) 
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`

func (s *sessionRepository) Create(ctx context.Context, session *domain.Session) error {
	querier := s.getQuerier(ctx)
	_, err := querier.Exec(
		ctx,
		createSession,
		session.ID, session.UserID, session.FamilyID, session.ParentSessionID, session.RefreshTokenHash,
		session.IPAddress, session.UserAgent, session.DeviceType, session.OSName, session.OSVersion,
		session.Label, session.CreatedAt, session.LastUsedAt, session.ExpiresAt, session.RevokedAt,
	)

	return err
}

const getSessionByID = `SELECT ` + sessionColumns + ` FROM sessions WHERE id = $1`

func (s *sessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	querier := s.getQuerier(ctx)
	return scanOne(querier.QueryRow(ctx, getSessionByID, id))
}

const getSessionByRefreshToken = `SELECT ` + sessionColumns + ` FROM sessions WHERE refresh_token_hash = $1`

func (s *sessionRepository) GetByRefreshToken(ctx context.Context, refreshTokenHash string) (*domain.Session, error) {
	querier := s.getQuerier(ctx)
	return scanOne(querier.QueryRow(ctx, getSessionByRefreshToken, refreshTokenHash))
}

const revokeSession = `
UPDATE sessions
SET revoked_at = $1
WHERE id = $2 AND revoked_at IS NULL
`

func (s *sessionRepository) Revoke(ctx context.Context, session *domain.Session) error {
	querier := s.getQuerier(ctx)

	cmdTag, err := querier.Exec(
		ctx,
		revokeSession,
		session.RevokedAt,
		session.ID,
	)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

const revokeFamilySessions = `
UPDATE sessions
SET revoked_at = now()
WHERE family_id = $1 
	AND revoked_at IS NULL
	AND expires_at > now()
`

func (s *sessionRepository) RevokeFamily(ctx context.Context, familyID uuid.UUID) error {
	querier := s.getQuerier(ctx)
	_, err := querier.Exec(ctx, revokeFamilySessions, familyID)
	return err
}

const revokeAllSessions = `
UPDATE sessions
SET revoked_at = now()
WHERE user_id = $1 
	AND revoked_at IS NULL
	AND expires_at > now()
`

func (s *sessionRepository) RevokeAll(ctx context.Context, userID uuid.UUID) error {
	querier := s.getQuerier(ctx)
	_, err := querier.Exec(ctx, revokeAllSessions, userID)
	return err
}

const revokeAllSessionsExcept = `
UPDATE sessions
SET revoked_at = now()
WHERE user_id = $1 
	AND revoked_at IS NULL 
	AND expires_at > now()
	AND id != $2 OR $2 IS NULL
`

func (s *sessionRepository) RevokeAllExcept(ctx context.Context, userID, sessionID uuid.UUID) error {
	querier := s.getQuerier(ctx)
	_, err := querier.Exec(ctx, revokeAllSessionsExcept, userID, sessionID)
	return err
}

const countActiveSessions = `
SELECT COUNT(*) FROM sessions
WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > now()`

const listActiveByUserID = `
SELECT ` + sessionColumns + `
FROM sessions
WHERE user_id = $1
	AND revoked_at IS NULL
	AND expires_at > now()
ORDER BY last_used_at DESC, id DESC
LIMIT $2 OFFSET $3
`

func (s *sessionRepository) ListActiveByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Session, int, error) {
	querier := s.getQuerier(ctx)

	var total int
	if err := querier.QueryRow(ctx, countActiveSessions, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	rows, err := querier.Query(ctx, listActiveByUserID, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		s, err := scanRow(rows)
		if err != nil {
			return nil, 0, err
		}
		sessions = append(sessions, s)
	}

	rows.Close()

	return sessions, total, rows.Err()
}

const cleanupOldSession = `
WITH sessions_to_delete AS (
    SELECT id
    FROM sessions
    WHERE expires_at < $1
       OR (revoked_at IS NOT NULL AND revoked_at < $2)
    ORDER BY expires_at
    LIMIT 5000
)
DELETE FROM sessions s
USING sessions_to_delete d
WHERE s.id = d.id`

func (s *sessionRepository) CleanupOldSessions(ctx context.Context, expiredThreshold, revokedThreshold time.Duration) (int64, error) {
	now := time.Now()
	expiredCutoff := now.Add(-expiredThreshold)
	revokedCutoff := now.Add(-revokedThreshold)

	querier := s.getQuerier(ctx)
	cmdTag, err := querier.Exec(ctx, cleanupOldSession, expiredCutoff, revokedCutoff)
	if err != nil {
		return 0, err
	}

	return cmdTag.RowsAffected(), nil
}

func scanOne(row db.Row) (*domain.Session, error) {
	s, err := scanRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrSessionNotFound
	}
	return s, err
}

func scanRow(row db.Row) (*domain.Session, error) {
	var s domain.Session
	err := row.Scan(
		&s.ID, &s.UserID, &s.FamilyID, &s.ParentSessionID, &s.RefreshTokenHash,
		&s.IPAddress, &s.UserAgent, &s.DeviceType, &s.OSName, &s.OSVersion, &s.Label,
		&s.CreatedAt, &s.LastUsedAt, &s.ExpiresAt, &s.RevokedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
