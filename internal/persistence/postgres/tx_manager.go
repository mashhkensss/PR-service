package postgres

import (
	"context"
	"database/sql"
)

type txKey struct{}

// DBTX represents the minimal subset of *sql.DB / *sql.Tx needed for queries.
type DBTX interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// TxManager implements service.TxRunner.
type TxManager struct {
	db *sql.DB
}

// NewTxManager constructs TxManager for the given DB.
func NewTxManager(db *sql.DB) *TxManager {
	return &TxManager{db: db}
}

// WithinTx starts a database transaction and injects it into context.
func (m *TxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	ctx = context.WithValue(ctx, txKey{}, tx)
	if err := fn(ctx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// ExecutorFromContext returns the current transaction or base DB.
func ExecutorFromContext(ctx context.Context, db *sql.DB) DBTX {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok && tx != nil {
		return tx
	}
	return db
}
