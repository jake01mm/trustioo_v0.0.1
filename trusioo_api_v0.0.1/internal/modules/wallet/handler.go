package wallet

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Handler 钱包模块HTTP处理器
type Handler struct {
	service Service
	logger  *logrus.Logger
}

// NewHandler 创建新的钱包处理器
func NewHandler(service Service, logger *logrus.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// === 钱包相关接口 ===

// GetWallet 获取钱包信息
// @Summary 获取用户钱包信息
// @Description 获取当前用户的钱包余额、状态等信息
// @Tags 钱包
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} WalletResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/wallet [get]
func (h *Handler) GetWallet(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	wallet, err := h.service.GetWallet(ctx, userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get wallet")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to retrieve wallet information")
		return
	}

	c.JSON(http.StatusOK, wallet)
}

// SetTransactionPin 设置交易密码
// @Summary 设置交易密码
// @Description 用户设置6位数字交易密码，设置后可进行提现等操作
// @Tags 钱包
// @Accept json
// @Produce json
// @Param request body SetTransactionPinRequest true "设置交易密码请求"
// @Security ApiKeyAuth
// @Success 200 {object} OperationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/wallet/transaction-pin [post]
func (h *Handler) SetTransactionPin(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	var req SetTransactionPinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid set transaction pin request")
		h.respondError(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := h.service.SetTransactionPin(ctx, userID, &req); err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to set transaction pin")

		if err.Error() == "transaction pin already set" {
			h.respondError(c, http.StatusConflict, "Transaction pin already set", "Transaction pin has already been configured")
			return
		}

		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to set transaction pin")
		return
	}

	h.respondSuccess(c, "Transaction pin set successfully", nil)
}

// ChangeTransactionPin 修改交易密码
// @Summary 修改交易密码
// @Description 修改用户的交易密码
// @Tags 钱包
// @Accept json
// @Produce json
// @Param request body ChangeTransactionPinRequest true "修改交易密码请求"
// @Security ApiKeyAuth
// @Success 200 {object} OperationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/wallet/transaction-pin [put]
func (h *Handler) ChangeTransactionPin(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	var req ChangeTransactionPinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid change transaction pin request")
		h.respondError(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := h.service.ChangeTransactionPin(ctx, userID, &req); err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to change transaction pin")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to change transaction pin")
		return
	}

	h.respondSuccess(c, "Transaction pin changed successfully", nil)
}

// === 货币和汇率相关接口 ===

// GetCurrencies 获取支持的货币列表
// @Summary 获取支持的货币列表
// @Description 获取平台支持的所有货币信息
// @Tags 货币
// @Accept json
// @Produce json
// @Success 200 {object} CurrencyListResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/wallet/currencies [get]
func (h *Handler) GetCurrencies(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	currencies, err := h.service.GetCurrencies(ctx)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get currencies")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to retrieve currencies")
		return
	}

	c.JSON(http.StatusOK, currencies)
}

// GetExchangeRate 获取汇率
// @Summary 获取货币汇率
// @Description 获取指定货币对的汇率信息
// @Tags 货币
// @Accept json
// @Produce json
// @Param from query string true "源货币代码" example="TRU"
// @Param to query string true "目标货币代码" example="NGN"
// @Success 200 {object} ExchangeRateResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/wallet/exchange-rate [get]
func (h *Handler) GetExchangeRate(c *gin.Context) {
	var req GetExchangeRateRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid get exchange rate request")
		h.respondError(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	rate, err := h.service.GetExchangeRate(ctx, req.FromCurrency, req.ToCurrency)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"from": req.FromCurrency,
			"to":   req.ToCurrency,
		}).Error("Failed to get exchange rate")

		if err.Error() == "exchange rate not found" {
			h.respondError(c, http.StatusNotFound, "Exchange rate not found", "No exchange rate found for the specified currency pair")
			return
		}

		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to retrieve exchange rate")
		return
	}

	c.JSON(http.StatusOK, rate)
}

// === 银行相关接口 ===

// GetBanks 获取银行列表
// @Summary 获取银行列表
// @Description 获取支持的银行列表
// @Tags 银行
// @Accept json
// @Produce json
// @Param country_code query string false "国家代码" example="NGA"
// @Success 200 {object} BankListResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/wallet/banks [get]
func (h *Handler) GetBanks(c *gin.Context) {
	countryCode := c.Query("country_code")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	banks, err := h.service.GetBanks(ctx, countryCode)
	if err != nil {
		h.logger.WithError(err).WithField("country_code", countryCode).Error("Failed to get banks")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to retrieve banks")
		return
	}

	c.JSON(http.StatusOK, banks)
}

// === 银行账户相关接口 ===

// GetBankAccounts 获取用户银行账户列表
// @Summary 获取用户银行账户列表
// @Description 获取当前用户的所有银行账户
// @Tags 银行账户
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} BankAccountListResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/wallet/bank-accounts [get]
func (h *Handler) GetBankAccounts(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	accounts, err := h.service.GetUserBankAccounts(ctx, userID)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get bank accounts")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to retrieve bank accounts")
		return
	}

	c.JSON(http.StatusOK, accounts)
}

// AddBankAccount 添加银行账户
// @Summary 添加银行账户
// @Description 为当前用户添加新的银行账户
// @Tags 银行账户
// @Accept json
// @Produce json
// @Param request body AddBankAccountRequest true "添加银行账户请求"
// @Security ApiKeyAuth
// @Success 201 {object} BankAccountResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/wallet/bank-accounts [post]
func (h *Handler) AddBankAccount(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	var req AddBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid add bank account request")
		h.respondError(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	account, err := h.service.AddBankAccount(ctx, userID, &req)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to add bank account")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to add bank account")
		return
	}

	c.JSON(http.StatusCreated, account)
}

// UpdateBankAccount 更新银行账户
// @Summary 更新银行账户
// @Description 更新用户的银行账户信息
// @Tags 银行账户
// @Accept json
// @Produce json
// @Param account_id path string true "银行账户ID"
// @Param request body UpdateBankAccountRequest true "更新银行账户请求"
// @Security ApiKeyAuth
// @Success 200 {object} BankAccountResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/wallet/bank-accounts/{account_id} [put]
func (h *Handler) UpdateBankAccount(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	accountID := c.Param("account_id")
	if accountID == "" {
		h.respondError(c, http.StatusBadRequest, "Invalid request", "Account ID is required")
		return
	}

	var req UpdateBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid update bank account request")
		h.respondError(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	account, err := h.service.UpdateBankAccount(ctx, userID, accountID, &req)
	if err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":    userID,
			"account_id": accountID,
		}).Error("Failed to update bank account")

		if err.Error() == "account does not belong to user" {
			h.respondError(c, http.StatusNotFound, "Account not found", "Bank account not found or does not belong to user")
			return
		}

		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to update bank account")
		return
	}

	c.JSON(http.StatusOK, account)
}

// DeleteBankAccount 删除银行账户
// @Summary 删除银行账户
// @Description 删除用户的银行账户
// @Tags 银行账户
// @Accept json
// @Produce json
// @Param account_id path string true "银行账户ID"
// @Security ApiKeyAuth
// @Success 200 {object} OperationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/wallet/bank-accounts/{account_id} [delete]
func (h *Handler) DeleteBankAccount(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	accountID := c.Param("account_id")
	if accountID == "" {
		h.respondError(c, http.StatusBadRequest, "Invalid request", "Account ID is required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := h.service.DeleteBankAccount(ctx, userID, accountID); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":    userID,
			"account_id": accountID,
		}).Error("Failed to delete bank account")

		if err.Error() == "account does not belong to user" || err.Error() == "bank account not found" {
			h.respondError(c, http.StatusNotFound, "Account not found", "Bank account not found or does not belong to user")
			return
		}

		if err.Error() == "cannot delete default account" {
			h.respondError(c, http.StatusBadRequest, "Cannot delete default account", "Please set another account as default before deleting this one")
			return
		}

		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to delete bank account")
		return
	}

	h.respondSuccess(c, "Bank account deleted successfully", nil)
}

// === 辅助方法 ===

// CalculateWithdrawal 计算提现费用
func (h *Handler) CalculateWithdrawal(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	var req CalculateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid calculate withdrawal request")
		h.respondError(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	calculation, err := h.service.CalculateWithdrawal(ctx, userID, &req)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to calculate withdrawal")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to calculate withdrawal")
		return
	}

	c.JSON(http.StatusOK, calculation)
}

// CreateWithdrawal 创建提现申请
func (h *Handler) CreateWithdrawal(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	var req CreateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid create withdrawal request")
		h.respondError(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	withdrawal, err := h.service.CreateWithdrawalRequest(ctx, userID, &req)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to create withdrawal")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to create withdrawal")
		return
	}

	c.JSON(http.StatusCreated, withdrawal)
}

// GetUserWithdrawals 获取用户提现记录
func (h *Handler) GetUserWithdrawals(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	var req GetWithdrawalsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid get withdrawals request")
		h.respondError(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	withdrawals, err := h.service.GetUserWithdrawals(ctx, userID, &req)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get withdrawals")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to retrieve withdrawals")
		return
	}

	c.JSON(http.StatusOK, withdrawals)
}

// GetWithdrawal 获取提现详情
func (h *Handler) GetWithdrawal(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	withdrawalID := c.Param("withdrawal_id")
	if withdrawalID == "" {
		h.respondError(c, http.StatusBadRequest, "Invalid request", "Withdrawal ID is required")
		return
	}

	// 暂时返回未实现错误
	h.respondError(c, http.StatusNotImplemented, "Not implemented", "This feature is not yet implemented")
}

// CancelWithdrawal 取消提现申请
func (h *Handler) CancelWithdrawal(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	withdrawalID := c.Param("withdrawal_id")
	if withdrawalID == "" {
		h.respondError(c, http.StatusBadRequest, "Invalid request", "Withdrawal ID is required")
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	if err := h.service.CancelWithdrawal(ctx, userID, withdrawalID); err != nil {
		h.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":       userID,
			"withdrawal_id": withdrawalID,
		}).Error("Failed to cancel withdrawal")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to cancel withdrawal")
		return
	}

	h.respondSuccess(c, "Withdrawal cancelled successfully", nil)
}

// GetUserTransactions 获取用户交易记录
func (h *Handler) GetUserTransactions(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	var req GetTransactionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.WithError(err).Warn("Invalid get transactions request")
		h.respondError(c, http.StatusBadRequest, "Invalid request", err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	transactions, err := h.service.GetUserTransactions(ctx, userID, &req)
	if err != nil {
		h.logger.WithError(err).WithField("user_id", userID).Error("Failed to get transactions")
		h.respondError(c, http.StatusInternalServerError, "Internal server error", "Failed to retrieve transactions")
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// GetTransaction 获取交易详情
func (h *Handler) GetTransaction(c *gin.Context) {
	userID := h.getUserID(c)
	if userID == "" {
		h.respondError(c, http.StatusUnauthorized, "Unauthorized", "User not authenticated")
		return
	}

	transactionID := c.Param("transaction_id")
	if transactionID == "" {
		h.respondError(c, http.StatusBadRequest, "Invalid request", "Transaction ID is required")
		return
	}

	// 暂时返回未实现错误
	h.respondError(c, http.StatusNotImplemented, "Not implemented", "This feature is not yet implemented")
}

// === 辅助方法 ===

// getUserID 从上下文获取用户ID
func (h *Handler) getUserID(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}

	if id, ok := userID.(string); ok {
		return id
	}

	return ""
}

// respondError 响应错误
func (h *Handler) respondError(c *gin.Context, statusCode int, error, message string) {
	response := ErrorResponse{
		Error:     error,
		Message:   message,
		Timestamp: time.Now(),
	}
	c.JSON(statusCode, response)
}

// respondSuccess 响应成功
func (h *Handler) respondSuccess(c *gin.Context, message string, data interface{}) {
	response := OperationResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}
	c.JSON(http.StatusOK, response)
}
