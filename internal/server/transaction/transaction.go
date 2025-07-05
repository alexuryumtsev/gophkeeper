package transaction

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type transactionManager struct {
	db *pgxpool.Pool
}

func NewTransactionManager(db *pgxpool.Pool) TransactionManager {
	return &transactionManager{
		db: db,
	}
}

// mockTransactionManager для тестов
type mockTransactionManager struct{}

func NewMockTransactionManager() TransactionManager {
	return &mockTransactionManager{}
}

func (m *mockTransactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// Для тестов просто выполняем функцию без транзакции
	return fn(ctx)
}

func (tm *transactionManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create a new context with the transaction
	txCtx := context.WithValue(ctx, "tx", tx)

	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Helper function to get transaction from context
func GetTxFromContext(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value("tx").(pgx.Tx); ok {
		return tx
	}
	return nil
}