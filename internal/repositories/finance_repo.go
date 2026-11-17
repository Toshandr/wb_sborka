package repositories

import (
	"context"
	"database/sql"
)

type FinanceRepo struct{ db *sql.DB }

func NewFinanceRepo(db *sql.DB) *FinanceRepo { return &FinanceRepo{db: db} }

func (r *FinanceRepo) GetBalanceForUpdate(ctx context.Context, tx *sql.Tx, executorID int) (float64, error) {
	_, _ = tx.ExecContext(ctx,
		`INSERT INTO executor_balances (executor_id, balance) VALUES ($1, 0)
		 ON CONFLICT (executor_id) DO NOTHING`, executorID)
	var b float64
	err := tx.QueryRowContext(ctx,
		`SELECT balance FROM executor_balances WHERE executor_id = $1 FOR UPDATE`,
		executorID).Scan(&b)
	return b, err
}

func (r *FinanceRepo) UpdateBalance(ctx context.Context, tx *sql.Tx, executorID int, newBalance float64) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE executor_balances SET balance = $1 WHERE executor_id = $2`,
		newBalance, executorID)
	return err
}

func (r *FinanceRepo) SaveTransaction(ctx context.Context, tx *sql.Tx,
	executorID int, goodsID *int, typ string, amount, before, after float64) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO executor_transactions
		 (executor_id, goods_id, type, amount, balance_before, balance_after)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		executorID, goodsID, typ, amount, before, after)
	return err
}
