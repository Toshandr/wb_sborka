package handlers

import (
	"net/http"
	"strconv"

	"wms/internal/services"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct{ svc *services.AdminService }

func NewAdminHandler(s *services.AdminService) *AdminHandler { return &AdminHandler{svc: s} }

func (h *AdminHandler) GetUsers(c *gin.Context) {
	u, err := h.svc.GetUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, u)
}

func (h *AdminHandler) GetAllGoods(c *gin.Context) {
	g, err := h.svc.GetGoods(c.Request.Context(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, g)
}

func (h *AdminHandler) GetPendingGoods(c *gin.Context) {
	g, _ := h.svc.GetGoods(c.Request.Context(), "pending")
	c.JSON(http.StatusOK, g)
}

func (h *AdminHandler) GetDefectiveGoods(c *gin.Context) {
	g, _ := h.svc.GetGoods(c.Request.Context(), "defective")
	c.JSON(http.StatusOK, g)
}

func (h *AdminHandler) ApproveGoods(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.svc.Approve(c.Request.Context(), id, c.GetInt("user_id")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Товар одобрен"})
}

func (h *AdminHandler) RejectGoods(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите reason"})
		return
	}
	if err := h.svc.Reject(c.Request.Context(), id, c.GetInt("user_id"), req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Товар отклонён"})
}

func (h *AdminHandler) ResolveDefect(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.svc.ResolveDefect(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Товар возвращён на модерацию"})
}

func (h *AdminHandler) GetExecutorBalance(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	b, err := h.svc.GetBalance(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"executor_id": id, "balance": b})
}

func (h *AdminHandler) GetExecutorTransactions(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	t, err := h.svc.GetTransactions(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, t)
}
