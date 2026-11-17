package handlers

import (
	"net/http"
	"strconv"

	"wms/internal/services"

	"github.com/gin-gonic/gin"
)

type ExecutorHandler struct{ svc *services.ExecutorService }

func NewExecutorHandler(s *services.ExecutorService) *ExecutorHandler {
	return &ExecutorHandler{svc: s}
}

func (h *ExecutorHandler) ClaimOrder(c *gin.Context) {
	var req struct {
		Quantity int `json:"quantity"`
	}
	c.ShouldBindJSON(&req)
	orderID, ids, err := h.svc.ClaimOrder(c.Request.Context(), c.GetInt("user_id"), req.Quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"order_id": orderID, "goods_ids": ids})
}

func (h *ExecutorHandler) SendToContainer(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.svc.SendToContainer(c.Request.Context(), c.GetInt("user_id"), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Товар в таре, 5% начислено"})
}

func (h *ExecutorHandler) MarkAsLost(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)
	if err := h.svc.MarkAsLost(c.Request.Context(), c.GetInt("user_id"), id, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Товар списан, штраф применён"})
}

func (h *ExecutorHandler) MarkAsDefect(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Reason string `json:"reason"`
	}
	c.ShouldBindJSON(&req)
	if err := h.svc.MarkAsDefect(c.Request.Context(), c.GetInt("user_id"), id, req.Reason); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Товар отправлен в брак"})
}

func (h *ExecutorHandler) GetCart(c *gin.Context) {
	items, total, err := h.svc.GetCart(c.Request.Context(), c.GetInt("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total_commission": total})
}

func (h *ExecutorHandler) GetBalance(c *gin.Context) {
	b, err := h.svc.GetBalance(c.Request.Context(), c.GetInt("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"balance": b})
}

func (h *ExecutorHandler) GetTransactions(c *gin.Context) {
	t, err := h.svc.GetTransactions(c.Request.Context(), c.GetInt("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, t)
}

func (h *ExecutorHandler) GetMyOrders(c *gin.Context) {
	o, err := h.svc.GetOrders(c.Request.Context(), c.GetInt("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, o)
}

func (h *ExecutorHandler) GetMyOrder(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	o, err := h.svc.GetOrder(c.Request.Context(), c.GetInt("user_id"), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, o)
}
