package handlers

import (
	"net/http"
	"strconv"

	"wms/internal/services"

	"github.com/gin-gonic/gin"
)

type SellerHandler struct{ svc *services.SellerService }

func NewSellerHandler(s *services.SellerService) *SellerHandler { return &SellerHandler{svc: s} }

func (h *SellerHandler) CreateGoods(c *gin.Context) {
	var req services.CreateGoodsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	id, err := h.svc.CreateGoods(c.Request.Context(), c.GetInt("user_id"), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id, "status": "pending"})
}

func (h *SellerHandler) GetMyGoods(c *gin.Context) {
	list, err := h.svc.GetMyGoods(c.Request.Context(), c.GetInt("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *SellerHandler) GetMyGoodsByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	g, err := h.svc.GetMyGoodsByID(c.Request.Context(), c.GetInt("user_id"), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, g)
}

func (h *SellerHandler) GetMySales(c *gin.Context) {
	list, err := h.svc.GetMySales(c.Request.Context(), c.GetInt("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *SellerHandler) GetStatistics(c *gin.Context) {
	stats, err := h.svc.GetStatistics(c.Request.Context(), c.GetInt("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}
