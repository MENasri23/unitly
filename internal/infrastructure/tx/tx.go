package tx

import (
	"context"
	"database/sql"
	"unitly/internal/domain"
)

// Implement the Transaction Manager
type txManager struct {
	db *sql.DB
}

func NewSqlTxManager(db *sql.DB) domain.TxManager {
	return &txManager{db: db}
}

func (m *txManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Inject the transaction into the context
	txCtx := context.WithValue(ctx, txKey, tx)

	// Run the use case with the new context
	err = fn(txCtx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func GetTxOrNil(ctx context.Context) *sql.Tx {
	return ctx.Value(txKey).(*sql.Tx)
}

type contextKey string

const txKey contextKey = "dbContextKey"
