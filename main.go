package main

import (
	"database/sql"
	"log"
	"os"

	"wms/internal/handlers"
	"wms/internal/middleware"
	"wms/internal/repositories"
	"wms/internal/services"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET пустой. Проверьте .env")
	}
	log.Println("JWT_SECRET length:", len(secret))

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal("db open:", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatal("db ping:", err)
	}

	// КРИТИЧНО: middleware и AuthService должны использовать ОДИН секрет
	middleware.JwtSecret = []byte(secret)

	goodsRepo := repositories.NewGoodsRepo(db)
	finRepo := repositories.NewFinanceRepo(db)

	authSvc := services.NewAuthService(db, secret)
	sellerSvc := services.NewSellerService(db)
	adminSvc := services.NewAdminService(db)
	execSvc := services.NewExecutorService(db, goodsRepo, finRepo)

	authH := handlers.NewAuthHandler(authSvc)
	sellerH := handlers.NewSellerHandler(sellerSvc)
	adminH := handlers.NewAdminHandler(adminSvc)
	execH := handlers.NewExecutorHandler(execSvc)

	r := gin.Default()

	api := r.Group("/api")
	api.POST("/auth/register", authH.Register)
	api.POST("/auth/login", authH.Login)

	auth := api.Group("")
	auth.Use(middleware.AuthRequired())
	{
		auth.GET("/auth/me", authH.Me)
	}

	seller := api.Group("/seller")
	seller.Use(middleware.AuthRequired(), middleware.RequireRole("seller"))
	{
		seller.POST("/goods", sellerH.CreateGoods)
		seller.GET("/goods", sellerH.GetMyGoods)
		seller.GET("/goods/:id", sellerH.GetMyGoodsByID)
		seller.GET("/sales", sellerH.GetMySales)
		seller.GET("/statistics", sellerH.GetStatistics)
	}

	admin := api.Group("/admin")
	admin.Use(middleware.AuthRequired(), middleware.RequireRole("admin"))
	{
		admin.GET("/users", adminH.GetUsers)
		admin.GET("/goods", adminH.GetAllGoods)
		admin.GET("/goods/pending", adminH.GetPendingGoods)
		admin.GET("/goods/defective", adminH.GetDefectiveGoods)
		admin.POST("/goods/:id/approve", adminH.ApproveGoods)
		admin.POST("/goods/:id/reject", adminH.RejectGoods)
		admin.POST("/goods/defective/:id/resolve", adminH.ResolveDefect)
		admin.GET("/executors/:id/balance", adminH.GetExecutorBalance)
		admin.GET("/executors/:id/transactions", adminH.GetExecutorTransactions)
	}

	executor := api.Group("/executor")
	executor.Use(middleware.AuthRequired(), middleware.RequireRole("executor"))
	{
		executor.POST("/orders/claim", execH.ClaimOrder)
		executor.GET("/orders", execH.GetMyOrders)
		executor.GET("/orders/:id", execH.GetMyOrder)
		executor.POST("/goods/:id/container", execH.SendToContainer)
		executor.POST("/goods/:id/lost", execH.MarkAsLost)
		executor.POST("/goods/:id/defect", execH.MarkAsDefect)
		executor.GET("/cart", execH.GetCart)
		executor.GET("/balance", execH.GetBalance)
		executor.GET("/transactions", execH.GetTransactions)
	}

	log.Println("Server started on :8080")
	r.Run(":8080")
}
