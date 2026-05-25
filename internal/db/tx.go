package db

import (
	"context"
	"unitly/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Implement the Transaction Manager
type txManager struct {
	db *pgxpool.Pool
}

func NewSqlTxManager(db *pgxpool.Pool) domain.UnitOfWork {
	return &txManager{db: db}
}

func (m *txManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Inject the transaction into the context
	txCtx := context.WithValue(ctx, txKey, tx)

	// Run the use case with the new context
	err = fn(txCtx)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func GetTxOrNil(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(txKey).(pgx.Tx)
	return tx
}

type contextKey string

const txKey contextKey = "dbContextKey"
