package services

import (
	"context"
	"database/sql"

	"wms/internal/models"
)

type SellerService struct{ db *sql.DB }

func NewSellerService(db *sql.DB) *SellerService { return &SellerService{db: db} }

type CreateGoodsRequest struct {
	Title       string  `json:"title" binding:"required"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Description string  `json:"description"`
	WarehouseID int     `json:"warehouse_id" binding:"required"`
}

func (s *SellerService) CreateGoods(ctx context.Context, sellerID int, req CreateGoodsRequest) (int, error) {
	if req.Description == "" {
		req.Description = "{}"
	}
	var id int
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO goods (title, price, description, warehouse_id, salesman_id, seller_id, status, stored_at, created_at, updated_at)
		 VALUES ($1, $2, $3::jsonb, $4, $5, $5, 'pending', now(), now(), now())
		 RETURNING id`,
		req.Title, req.Price, req.Description, req.WarehouseID, sellerID).Scan(&id)
	return id, err
}

func (s *SellerService) GetMyGoods(ctx context.Context, sellerID int) ([]models.Goods, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, title, price, description::text, warehouse_id, seller_id, status,
		        assigned_executor_id, rejection_reason
		 FROM goods WHERE seller_id = $1 ORDER BY id DESC`, sellerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []models.Goods
	for rows.Next() {
		var g models.Goods
		if err := rows.Scan(&g.ID, &g.Title, &g.Price, &g.Description, &g.WarehouseID,
			&g.SellerID, &g.Status, &g.AssignedExecutorID, &g.RejectionReason); err != nil {
			return nil, err
		}
		list = append(list, g)
	}
	return list, nil
}

func (s *SellerService) GetMyGoodsByID(ctx context.Context, sellerID, id int) (*models.Goods, error) {
	var g models.Goods
	err := s.db.QueryRowContext(ctx,
		`SELECT id, title, price, description::text, warehouse_id, seller_id, status,
		        assigned_executor_id, rejection_reason
		 FROM goods WHERE id = $1 AND seller_id = $2`, id, sellerID).
		Scan(&g.ID, &g.Title, &g.Price, &g.Description, &g.WarehouseID,
			&g.SellerID, &g.Status, &g.AssignedExecutorID, &g.RejectionReason)
	return &g, err
}

func (s *SellerService) GetMySales(ctx context.Context, sellerID int) ([]map[string]any, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT s.id, s.goods_id, g.title, s.price, s.commission, s.sold_at::text
		 FROM sales s JOIN goods g ON g.id = s.goods_id
		 WHERE s.seller_id = $1 ORDER BY s.sold_at DESC`, sellerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id, goodsID int
		var title, soldAt string
		var price, commission float64
		rows.Scan(&id, &goodsID, &title, &price, &commission, &soldAt)
		out = append(out, map[string]any{
			"id": id, "goods_id": goodsID, "title": title,
			"price": price, "commission": commission, "sold_at": soldAt,
		})
	}
	return out, nil
}

func (s *SellerService) GetStatistics(ctx context.Context, sellerID int) (map[string]any, error) {
	var count int
	var total float64
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*), COALESCE(SUM(price), 0) FROM sales WHERE seller_id = $1`,
		sellerID).Scan(&count, &total)
	return map[string]any{"sold_count": count, "total_sum": total}, err
}
