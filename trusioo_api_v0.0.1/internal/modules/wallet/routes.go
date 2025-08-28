package wallet

import (
	"trusioo_api_v0.0.1/internal/modules/auth"

	"github.com/gin-gonic/gin"
)

// Routes 钱包模块路由
type Routes struct {
	handler    *Handler
	authMiddle *auth.AuthMiddleware
}

// NewRoutes 创建新的钱包路由
func NewRoutes(handler *Handler, authMiddle *auth.AuthMiddleware) *Routes {
	return &Routes{
		handler:    handler,
		authMiddle: authMiddle,
	}
}

// RegisterRoutes 注册钱包路由
func (r *Routes) RegisterRoutes(router *gin.RouterGroup) {
	// 钱包路由组
	walletGroup := router.Group("/wallet")
	{
		// 公开接口（不需要认证）
		r.registerPublicRoutes(walletGroup)

		// 用户接口（需要用户认证）
		r.registerUserRoutes(walletGroup)

		// 管理员接口（需要管理员认证）
		r.registerAdminRoutes(walletGroup)
	}
}

// registerPublicRoutes 注册公开路由
func (r *Routes) registerPublicRoutes(group *gin.RouterGroup) {
	// 公开接口组
	public := group.Group("")
	{
		// 货币相关（公开）
		public.GET("/currencies", r.handler.GetCurrencies)
		public.GET("/exchange-rate", r.handler.GetExchangeRate)

		// 银行相关（公开）
		public.GET("/banks", r.handler.GetBanks)
	}
}

// registerUserRoutes 注册用户路由
func (r *Routes) registerUserRoutes(group *gin.RouterGroup) {
	// 用户接口组 - 需要用户认证
	user := group.Group("")
	user.Use(r.authMiddle.RequireAuth())
	user.Use(r.authMiddle.RequireUserType("user"))
	{
		// === 钱包基础接口 ===

		// 获取钱包信息
		user.GET("", r.handler.GetWallet)

		// 交易密码管理
		user.POST("/transaction-pin", r.handler.SetTransactionPin)
		user.PUT("/transaction-pin", r.handler.ChangeTransactionPin)

		// === 银行账户管理 ===

		// 银行账户CRUD
		user.GET("/bank-accounts", r.handler.GetBankAccounts)
		user.POST("/bank-accounts", r.handler.AddBankAccount)
		user.PUT("/bank-accounts/:account_id", r.handler.UpdateBankAccount)
		user.DELETE("/bank-accounts/:account_id", r.handler.DeleteBankAccount)

		// === 提现相关 ===

		// 提现费用计算
		user.POST("/withdrawal/calculate", r.handler.CalculateWithdrawal)

		// 提现申请
		user.POST("/withdrawals", r.handler.CreateWithdrawal)
		user.GET("/withdrawals", r.handler.GetUserWithdrawals)
		user.GET("/withdrawals/:withdrawal_id", r.handler.GetWithdrawal)
		user.POST("/withdrawals/:withdrawal_id/cancel", r.handler.CancelWithdrawal)

		// === 交易记录 ===

		// 交易记录查询
		user.GET("/transactions", r.handler.GetUserTransactions)
		user.GET("/transactions/:transaction_id", r.handler.GetTransaction)
	}
}

// registerAdminRoutes 注册管理员路由
func (r *Routes) registerAdminRoutes(group *gin.RouterGroup) {
	// 管理员接口组 - 需要管理员认证
	admin := group.Group("/admin")
	admin.Use(r.authMiddle.RequireAuth())
	admin.Use(r.authMiddle.RequireUserType("admin"))
	{
		// === 提现管理 ===

		// 提现申请管理
		admin.GET("/withdrawals", r.handler.GetPendingWithdrawals)
		admin.GET("/withdrawals/:withdrawal_id", r.handler.GetWithdrawalDetail)
		admin.POST("/withdrawals/:withdrawal_id/review", r.handler.ReviewWithdrawal)
		admin.POST("/withdrawals/:withdrawal_id/process", r.handler.ProcessWithdrawal)

		// === 汇率管理 ===

		// 汇率设置
		admin.POST("/exchange-rates", r.handler.CreateExchangeRate)
		admin.PUT("/exchange-rates/:rate_id", r.handler.UpdateExchangeRate)
		admin.GET("/exchange-rates", r.handler.GetExchangeRates)

		// === 钱包管理 ===

		// 钱包调整
		admin.POST("/wallets/adjust", r.handler.AdjustWallet)
		admin.GET("/wallets/:user_id", r.handler.GetUserWallet)
		admin.POST("/wallets/:user_id/freeze", r.handler.FreezeWallet)
		admin.POST("/wallets/:user_id/unfreeze", r.handler.UnfreezeWallet)

		// === 统计报告 ===

		// 统计信息
		admin.GET("/statistics/wallets", r.handler.GetWalletStatistics)
		admin.GET("/statistics/transactions", r.handler.GetTransactionStatistics)
		admin.GET("/statistics/withdrawals", r.handler.GetWithdrawalStatistics)
	}
}

// === 未来扩展路由占位 ===

// 以下路由在第一阶段不实现，仅作为规划参考：

/*
// === 高级功能路由（未来实现） ===

// 转账功能
user.POST("/transfer", r.handler.CreateTransfer)
user.GET("/transfers", r.handler.GetUserTransfers)

// 充值功能
user.POST("/deposits", r.handler.CreateDeposit)
user.GET("/deposits", r.handler.GetUserDeposits)

// 钱包设置
user.GET("/settings", r.handler.GetWalletSettings)
user.PUT("/settings", r.handler.UpdateWalletSettings)

// 通知设置
user.GET("/notifications", r.handler.GetNotificationSettings)
user.PUT("/notifications", r.handler.UpdateNotificationSettings)

// === 管理员高级功能（未来实现） ===

// 批量操作
admin.POST("/wallets/batch/freeze", r.handler.BatchFreezeWallets)
admin.POST("/wallets/batch/adjust", r.handler.BatchAdjustWallets)

// 审计日志
admin.GET("/audit-logs", r.handler.GetAuditLogs)
admin.GET("/audit-logs/:log_id", r.handler.GetAuditLogDetail)

// 风控管理
admin.GET("/risk/suspicious-transactions", r.handler.GetSuspiciousTransactions)
admin.POST("/risk/rules", r.handler.CreateRiskRule)
admin.GET("/risk/rules", r.handler.GetRiskRules)

// 报表导出
admin.GET("/reports/wallets/export", r.handler.ExportWalletReport)
admin.GET("/reports/transactions/export", r.handler.ExportTransactionReport)
admin.GET("/reports/withdrawals/export", r.handler.ExportWithdrawalReport)
*/
