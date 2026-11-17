package repositories

import (
	"context"
	"database/sql"
)

type GoodsRepo struct{ db *sql.DB }

func NewGoodsRepo(db *sql.DB) *GoodsRepo { return &GoodsRepo{db: db} }

type GoodsLocked struct {
	ID                 int
	Title              string
	Price              float64
	SellerID           int
	AssignedExecutorID *int
	Status             string
}

func (r *GoodsRepo) GetForUpdate(ctx context.Context, tx *sql.Tx, id int) (*GoodsLocked, error) {
	var g GoodsLocked
	err := tx.QueryRowContext(ctx,
		`SELECT id, title, price, seller_id, assigned_executor_id, status
		 FROM goods WHERE id = $1 FOR UPDATE`, id).
		Scan(&g.ID, &g.Title, &g.Price, &g.SellerID, &g.AssignedExecutorID, &g.Status)
	return &g, err
}

func (r *GoodsRepo) UpdateStatus(ctx context.Context, tx *sql.Tx, id int, status string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE goods SET status = $1, updated_at = now() WHERE id = $2`, status, id)
	return err
}
