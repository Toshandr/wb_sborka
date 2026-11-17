package services

import (
	"context"
	"database/sql"
	"errors"

	"wms/internal/models"
	"wms/internal/repositories"
)

type ExecutorService struct {
	db        *sql.DB
	goodsRepo *repositories.GoodsRepo
	finRepo   *repositories.FinanceRepo
}

func NewExecutorService(db *sql.DB, gr *repositories.GoodsRepo, fr *repositories.FinanceRepo) *ExecutorService {
	return &ExecutorService{db: db, goodsRepo: gr, finRepo: fr}
}

func (s *ExecutorService) ClaimOrder(ctx context.Context, executorID, quantity int) (int, []int, error) {
	if quantity <= 0 {
		quantity = 20
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, nil, err
	}
	defer tx.Rollback()

	var orderID int
	if err := tx.QueryRowContext(ctx,
		`INSERT INTO orders (executor_id, status, requested_quantity, created_at)
		 VALUES ($1, 'active', $2, now()) RETURNING id`,
		executorID, quantity).Scan(&orderID); err != nil {
		return 0, nil, err
	}

	rows, err := tx.QueryContext(ctx,
		`SELECT id FROM goods WHERE status = 'available' LIMIT $1 FOR UPDATE SKIP LOCKED`,
		quantity)
	if err != nil {
		return 0, nil, err
	}
	var ids []int
	for rows.Next() {
		var id int
		rows.Scan(&id)
		ids = append(ids, id)
	}
	rows.Close()

	if len(ids) == 0 {
		return 0, nil, errors.New("нет доступных товаров")
	}

	for _, gid := range ids {
		if _, err := tx.ExecContext(ctx,
			`UPDATE goods SET status = 'assigned', assigned_executor_id = $1,
			        order_id = $2, updated_at = now()
			 WHERE id = $3`, executorID, orderID, gid); err != nil {
			return 0, nil, err
		}
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE orders SET assigned_quantity = $1 WHERE id = $2`,
		len(ids), orderID); err != nil {
		return 0, nil, err
	}

	return orderID, ids, tx.Commit()
}

func (s *ExecutorService) SendToContainer(ctx context.Context, executorID, goodsID int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	g, err := s.goodsRepo.GetForUpdate(ctx, tx, goodsID)
	if err != nil {
		return err
	}
	if g.AssignedExecutorID == nil || *g.AssignedExecutorID != executorID {
		return errors.New("товар не закреплён за вами")
	}
	if g.Status != "assigned" {
		return errors.New("товар уже обработан")
	}

	commission := round2(g.Price * 0.05)

	before, err := s.finRepo.GetBalanceForUpdate(ctx, tx, executorID)
	if err != nil {
		return err
	}
	after := round2(before + commission)
	if err := s.finRepo.UpdateBalance(ctx, tx, executorID, after); err != nil {
		return err
	}
	if err := s.finRepo.SaveTransaction(ctx, tx, executorID, &goodsID, "earning", commission, before, after); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO sales (goods_id, seller_id, executor_id, price, commission)
		 VALUES ($1, $2, $3, $4, $5)`,
		goodsID, g.SellerID, executorID, g.Price, commission); err != nil {
		return err
	}

	return s.goodsRepo.UpdateStatus(ctx, tx, goodsID, "sold")
}

func (s *ExecutorService) MarkAsLost(ctx context.Context, executorID, goodsID int, reason string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	g, err := s.goodsRepo.GetForUpdate(ctx, tx, goodsID)
	if err != nil {
		return err
	}
	if g.AssignedExecutorID == nil || *g.AssignedExecutorID != executorID || g.Status != "assigned" {
		return errors.New("нельзя обработать этот товар")
	}

	penalty := round2(500 + g.Price*0.10)

	before, err := s.finRepo.GetBalanceForUpdate(ctx, tx, executorID)
	if err != nil {
		return err
	}
	after := round2(before - penalty)
	if err := s.finRepo.UpdateBalance(ctx, tx, executorID, after); err != nil {
		return err
	}
	if err := s.finRepo.SaveTransaction(ctx, tx, executorID, &goodsID, "lost_penalty", penalty, before, after); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO lost_goods (goods_id, executor_id, reason, penalty)
		 VALUES ($1, $2, $3, $4)`, goodsID, executorID, reason, penalty); err != nil {
		return err
	}

	return s.goodsRepo.UpdateStatus(ctx, tx, goodsID, "lost")
}

func (s *ExecutorService) MarkAsDefect(ctx context.Context, executorID, goodsID int, reason string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	g, err := s.goodsRepo.GetForUpdate(ctx, tx, goodsID)
	if err != nil {
		return err
	}
	if g.AssignedExecutorID == nil || *g.AssignedExecutorID != executorID || g.Status != "assigned" {
		return errors.New("нельзя обработать этот товар")
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT INTO defective_goods (goods_id, executor_id, reason)
		 VALUES ($1, $2, $3)`, goodsID, executorID, reason); err != nil {
		return err
	}
	return s.goodsRepo.UpdateStatus(ctx, tx, goodsID, "defective")
}

func (s *ExecutorService) GetCart(ctx context.Context, executorID int) ([]map[string]any, float64, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT s.id, s.goods_id, g.title, s.price, s.commission, s.sold_at::text
		 FROM sales s JOIN goods g ON g.id = s.goods_id
		 WHERE s.executor_id = $1 ORDER BY s.sold_at DESC`, executorID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []map[string]any
	var total float64
	for rows.Next() {
		var id, goodsID int
		var title, soldAt string
		var price, commission float64
		rows.Scan(&id, &goodsID, &title, &price, &commission, &soldAt)
		total += commission
		list = append(list, map[string]any{
			"id": id, "goods_id": goodsID, "title": title,
			"price": price, "commission": commission, "sold_at": soldAt,
		})
	}
	return list, round2(total), nil
}

func (s *ExecutorService) GetBalance(ctx context.Context, executorID int) (float64, error) {
	var b float64
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(balance, 0) FROM executor_balances WHERE executor_id = $1`,
		executorID).Scan(&b)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return b, err
}

func (s *ExecutorService) GetTransactions(ctx context.Context, executorID int) ([]models.ExecutorTransaction, error) {
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

func (s *ExecutorService) GetOrders(ctx context.Context, executorID int) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, status, requested_quantity, assigned_quantity, created_at::text
		 FROM orders WHERE executor_id = $1 ORDER BY id DESC`, executorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []map[string]any
	for rows.Next() {
		var id, req, asg int
		var status, created string
		rows.Scan(&id, &status, &req, &asg, &created)
		list = append(list, map[string]any{
			"id": id, "status": status, "requested_quantity": req,
			"assigned_quantity": asg, "created_at": created,
		})
	}
	return list, nil
}

func (s *ExecutorService) GetOrder(ctx context.Context, executorID, orderID int) (map[string]any, error) {
	var id, req, asg int
	var status, created string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, status, requested_quantity, assigned_quantity, created_at::text
		 FROM orders WHERE id = $1 AND executor_id = $2`, orderID, executorID).
		Scan(&id, &status, &req, &asg, &created)
	if err != nil {
		return nil, err
	}
	rows, _ := s.db.QueryContext(ctx,
		`SELECT id, title, price, status FROM goods WHERE order_id = $1`, orderID)
	defer rows.Close()
	var items []map[string]any
	for rows.Next() {
		var gid int
		var title, st string
		var price float64
		rows.Scan(&gid, &title, &price, &st)
		items = append(items, map[string]any{
			"id": gid, "title": title, "price": price, "status": st,
		})
	}
	return map[string]any{
		"id": id, "status": status, "requested_quantity": req,
		"assigned_quantity": asg, "created_at": created, "items": items,
	}, nil
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
