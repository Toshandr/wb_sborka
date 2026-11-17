package services

import (
	"context"
	"database/sql"

	"wms/internal/models"
)

type AdminService struct{ db *sql.DB }

func NewAdminService(db *sql.DB) *AdminService { return &AdminService{db: db} }

func (s *AdminService) GetUsers(ctx context.Context) ([]models.User, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, email, role, is_active FROM users ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Email, &u.Role, &u.IsActive)
		list = append(list, u)
	}
	return list, nil
}

func (s *AdminService) GetGoods(ctx context.Context, status string) ([]models.Goods, error) {
	q := `SELECT id, title, price, description::text, warehouse_id, seller_id, status,
	              assigned_executor_id, rejection_reason FROM goods`
	args := []any{}
	if status != "" {
		q += ` WHERE status = $1`
		args = append(args, status)
	}
	q += ` ORDER BY id DESC`
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Goods
	for rows.Next() {
		var g models.Goods
		rows.Scan(&g.ID, &g.Title, &g.Price, &g.Description, &g.WarehouseID,
			&g.SellerID, &g.Status, &g.AssignedExecutorID, &g.RejectionReason)
		list = append(list, g)
	}
	return list, nil
}

func (s *AdminService) Approve(ctx context.Context, goodsID, adminID int) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE goods SET status = 'available', moderated_by = $1, moderated_at = now(), updated_at = now()
		 WHERE id = $2 AND status = 'pending'`, adminID, goodsID)
	return err
}

func (s *AdminService) Reject(ctx context.Context, goodsID, adminID int, reason string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE goods SET status = 'rejected', rejection_reason = $1, moderated_by = $2,
		        moderated_at = now(), updated_at = now()
		 WHERE id = $3 AND status = 'pending'`, reason, adminID, goodsID)
	return err
}

func (s *AdminService) ResolveDefect(ctx context.Context, goodsID int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx,
		`UPDATE goods SET status = 'pending', updated_at = now()
		 WHERE id = $1 AND status = 'defective'`, goodsID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE defective_goods SET resolved = true, resolved_at = now()
		 WHERE goods_id = $1 AND resolved = false`, goodsID); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *AdminService) GetBalance(ctx context.Context, executorID int) (float64, error) {
	var b float64
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(balance, 0) FROM executor_balances WHERE executor_id = $1`,
		executorID).Scan(&b)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return b, err
}

func (s *AdminService) GetTransactions(ctx context.Context, executorID int) ([]models.ExecutorTransaction, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, executor_id, goods_id, type, amount, balance_before, balance_after, created_at::text
		 FROM executor_transactions WHERE executor_id = $1 ORDER BY id DESC`, executorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.ExecutorTransaction
	for rows.Next() {
		var t models.ExecutorTransaction
		rows.Scan(&t.ID, &t.ExecutorID, &t.GoodsID, &t.Type, &t.Amount,
			&t.BalanceBefore, &t.BalanceAfter, &t.CreatedAt)
		list = append(list, t)
	}
	return list, nil
}
