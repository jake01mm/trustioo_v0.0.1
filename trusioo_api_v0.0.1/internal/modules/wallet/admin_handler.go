package wallet

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// === 管理员提现管理接口 ===

// GetPendingWithdrawals 获取待处理提现申请
func (h *Handler) GetPendingWithdrawals(c *gin.Context) {
	var req GetWithdrawalsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid get pending withdrawals request")
		h.respondError(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	withdrawals, err := h.service.GetPendingWithdrawals(ctx, &req)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get pending withdrawals")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to retrieve pending withdrawals")
		return
	}

	c.JSON(http.StatusOK, withdrawals)
}

// GetWithdrawalDetail 获取提现详情（管理员）
func (h *Handler) GetWithdrawalDetail(c *gin.Context) {
	withdrawalID := c.Param("withdrawal_id")
	if withdrawalID == "" {
		h.respondError(c, http.StatusBadRequest, "Invalid request", "Withdrawal ID is required")
		return
	}

	// 暂时返回未实现错误
	h.respondError(c, http.StatusNotImplemented, "Not implemented", "This feature is not yet implemented")
}

// ReviewWithdrawal 审核提现申请
func (h *Handler) ReviewWithdrawal(c *gin.Context) {
	adminID := h.getUserID(c)
	if adminID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "Admin not authenticated")
		return
	}

	withdrawalID := c.Param("withdrawal_id")
	if withdrawalID == "" {
		h.respondError(c, http.StatusBadRequest, "Invalid request", "Withdrawal ID is required")
		return
	}

	var req AdminReviewWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid review withdrawal request")
		h.respondError(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := h.service.ReviewWithdrawal(ctx, adminID, withdrawalID, &req); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"admin_id":      adminID,
			"withdrawal_id": withdrawalID,
		}).Error("Failed to review withdrawal")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to review withdrawal")
		return
	}

	h.respondSuccess(c, "Withdrawal reviewed successfully", nil)
}

// ProcessWithdrawal 处理提现申请
func (h *Handler) ProcessWithdrawal(c *gin.Context) {
	adminID := h.getUserID(c)
	if adminID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "Admin not authenticated")
		return
	}

	withdrawalID := c.Param("withdrawal_id")
	if withdrawalID == "" {
		h.respondError(c, http.StatusBadRequest, "Invalid request", "Withdrawal ID is required")
		return
	}

	var req AdminProcessWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid process withdrawal request")
		h.respondError(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := h.service.ProcessWithdrawal(ctx, adminID, withdrawalID, &req); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"admin_id":      adminID,
			"withdrawal_id": withdrawalID,
		}).Error("Failed to process withdrawal")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to process withdrawal")
		return
	}

	h.respondSuccess(c, "Withdrawal processed successfully", nil)
}

// === 汇率管理接口 ===

// CreateExchangeRate 创建汇率
func (h *Handler) CreateExchangeRate(c *gin.Context) {
	adminID := h.getUserID(c)
	if adminID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "Admin not authenticated")
		return
	}

	var req AdminUpdateExchangeRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid create exchange rate request")
		h.respondError(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := h.service.UpdateExchangeRate(ctx, adminID, &req); err != nil {
		h.logger.WithError(err).WithField("admin_id", adminID).Error("Failed to create exchange rate")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to create exchange rate")
		return
	}

	h.respondSuccess(c, "Exchange rate created successfully", nil)
}

// UpdateExchangeRate 更新汇率
func (h *Handler) UpdateExchangeRate(c *gin.Context) {
	adminID := h.getUserID(c)
	if adminID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "Admin not authenticated")
		return
	}

	rateID := c.Param("rate_id")
	if rateID == "" {
		h.respondError(c, http.StatusBadRequest, "Invalid request", "Rate ID is required")
		return
	}

	var req AdminUpdateExchangeRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid update exchange rate request")
		h.respondError(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := h.service.UpdateExchangeRate(ctx, adminID, &req); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"admin_id": adminID,
			"rate_id":  rateID,
		}).Error("Failed to update exchange rate")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to update exchange rate")
		return
	}

	h.respondSuccess(c, "Exchange rate updated successfully", nil)
}

// GetExchangeRates 获取汇率列表（管理员）
func (h *Handler) GetExchangeRates(c *gin.Context) {
	// 暂时返回未实现错误
	h.respondError(c, http.StatusNotImplemented, "Not implemented", "This feature is not yet implemented")
}

// === 钱包管理接口 ===

// AdjustWallet 调整钱包余额
func (h *Handler) AdjustWallet(c *gin.Context) {
	adminID := h.getUserID(c)
	if adminID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "Admin not authenticated")
		return
	}

	var req AdminWalletAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid wallet adjustment request")
		h.respondError(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := h.service.AdjustWallet(ctx, adminID, &req); err != nil {
		h.logger.WithError(err).WithField("admin_id", adminID).Error("Failed to adjust wallet")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to adjust wallet")
		return
	}

	h.respondSuccess(c, "Wallet adjusted successfully", nil)
}

// GetUserWallet 获取用户钱包（管理员）
func (h *Handler) GetUserWallet(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		h.respondError(c, http.StatusBadRequest, "Invalid request", "User ID is required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	wallet, err := h.service.GetWallet(ctx, userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user wallet")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to retrieve user wallet")
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// FreezeWallet 冻结钱包
func (h *Handler) FreezeWallet(c *gin.Context) {
	adminID := h.getUserID(c)
	if adminID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "Admin not authenticated")
		return
	}

	userID := c.Param("user_id")
	if userID == "" {
		h.respondError(c, http.StatusBadRequest, "Invalid request", "User ID is required")
		return
	}

	// 暂时返回未实现错误
	h.respondError(c, http.StatusNotImplemented, "Not implemented", "This feature is not yet implemented")
}

// UnfreezeWallet 解冻钱包
func (h *Handler) UnfreezeWallet(c *gin.Context) {
	adminID := h.getUserID(c)
	if adminID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "Admin not authenticated")
		return
	}

	userID := c.Param("user_id")
	if userID == "" {
		h.respondError(c, http.StatusBadRequest, "Invalid request", "User ID is required")
		return
	}

	// 暂时返回未实现错误
	h.respondError(c, http.StatusNotImplemented, "Not implemented", "This feature is not yet implemented")
}

// === 统计报告接口 ===

// GetWalletStatistics 获取钱包统计
func (h *Handler) GetWalletStatistics(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	statistics, err := h.service.GetWalletStatistics(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get wallet statistics")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to retrieve wallet statistics")
		return
	}

	c.JSON(http.StatusOK, statistics)
}

// GetTransactionStatistics 获取交易统计
func (h *Handler) GetTransactionStatistics(c *gin.Context) {
	// 暂时返回未实现错误
	h.respondError(c, http.StatusNotImplemented, "Not implemented", "This feature is not yet implemented")
}

// GetWithdrawalStatistics 获取提现统计
func (h *Handler) GetWithdrawalStatistics(c *gin.Context) {
	// 暂时返回未实现错误
	h.respondError(c, http.StatusNotImplemented, "Not implemented", "This feature is not yet implemented")
}