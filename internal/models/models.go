package models

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	IsActive bool   `json:"is_active"`
}

type Goods struct {
	ID                 int     `json:"id"`
	Title              string  `json:"title"`
	Price              float64 `json:"price"`
	Description        string  `json:"description"`
	WarehouseID        int     `json:"warehouse_id"`
	SellerID           int     `json:"seller_id"`
	Status             string  `json:"status"`
	AssignedExecutorID *int    `json:"assigned_executor_id"`
	RejectionReason    *string `json:"rejection_reason"`
}

type ExecutorTransaction struct {
	ID            int     `json:"id"`
	ExecutorID    int     `json:"executor_id"`
	GoodsID       *int    `json:"goods_id"`
	Type          string  `json:"type"`
	Amount        float64 `json:"amount"`
	BalanceBefore float64 `json:"balance_before"`
	BalanceAfter  float64 `json:"balance_after"`
	CreatedAt     string  `json:"created_at"`
}
