package repository

import (
	"context"
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

func (s *sessionRepository) Create(ctx context.Context, session *domain.Session) error {
	query := `
    INSERT INTO sessions (id, user_id, refresh_token_hash, expires_at)
    VALUES ($1, $2, $3, $4)
    RETURNING created_at
    `

	querier := s.getQuerier(ctx)
	err := querier.QueryRow(
		ctx,
		query,
		session.ID,
		session.UserID,
		session.RefreshTokenHash,
		session.ExpiresAt,
	).Scan(&session.CreatedAt)

	return err
}

func (s *sessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	query := `
    SELECT id, user_id, refresh_token_hash, created_at, expires_at, revoked_at, version
    FROM sessions
    WHERE id = $1
    `

	querier := s.getQuerier(ctx)

	var session domain.Session
	err := querier.QueryRow(ctx, query, id).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.Version,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &session, nil
}

func (s *sessionRepository) GetByRefreshToken(ctx context.Context, refreshTokenHash string) (*domain.Session, error) {
	query := `
    SELECT id, user_id, refresh_token_hash, created_at, expires_at, revoked_at, version
    FROM sessions
    WHERE refresh_token_hash = $1
	`

	querier := s.getQuerier(ctx)

	var session domain.Session
	err := querier.QueryRow(ctx, query, refreshTokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.RevokedAt,
		&session.Version,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}

	return &session, nil
}

func (s *sessionRepository) Revoke(ctx context.Context, session *domain.Session) error {
	query := `
    UPDATE sessions
    SET revoked_at = $1, version = version + 1
    WHERE id = $2 AND revoked_at IS NULL
    `

	querier := s.getQuerier(ctx)
	cmdTag, err := querier.Exec(ctx, query, session.RevokedAt, session.ID)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (s *sessionRepository) RevokeAll(ctx context.Context, userID uuid.UUID) error {
	query := `
    UPDATE sessions
    SET revoked_at = NOW(), version = version + 1
    WHERE user_id = $1 AND revoked_at IS NULL
    `

	querier := s.getQuerier(ctx)
	_, err := querier.Exec(ctx, query, userID)
	return err
}

func (s *sessionRepository) CleanupOldSessions(ctx context.Context, expiredThreshold, revokedThreshold time.Duration) (int64, error) {
	query := `
		DELETE FROM sessions 
		WHERE expires_at < $1
		   OR (revoked_at IS NOT NULL AND created_at < $2);
	`

	now := time.Now()
	expiredCutoff := now.Add(-expiredThreshold)
	revokedCutoff := now.Add(-revokedThreshold)

	querier := s.getQuerier(ctx)
	cmdTag, err := querier.Exec(ctx, query, expiredCutoff, revokedCutoff)
	if err != nil {
		return 0, err
	}

	return cmdTag.RowsAffected(), nil
}
