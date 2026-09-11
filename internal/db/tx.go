package db

import (
	"context"
	"unitly/internal/domain"

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
	if GetQuerier(ctx, nil) != nil {
		return fn(ctx)
	}

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
