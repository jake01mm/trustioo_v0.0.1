package wallet

import (
	"fmt"
	"time"
)

// === 请求DTO ===

// SetTransactionPinRequest 设置交易密码请求
type SetTransactionPinRequest struct {
	Pin        string `json:"pin" binding:"required,len=6,numeric" example:"123456"`
	ConfirmPin string `json:"confirm_pin" binding:"required,len=6,numeric" example:"123456"`
}

// VerifyTransactionPinRequest 验证交易密码请求
type VerifyTransactionPinRequest struct {
	Pin string `json:"pin" binding:"required,len=6,numeric" example:"123456"`
}

// ChangeTransactionPinRequest 修改交易密码请求
type ChangeTransactionPinRequest struct {
	CurrentPin string `json:"current_pin" binding:"required,len=6,numeric" example:"123456"`
	NewPin     string `json:"new_pin" binding:"required,len=6,numeric" example:"654321"`
	ConfirmPin string `json:"confirm_pin" binding:"required,len=6,numeric" example:"654321"`
}

// CreateWithdrawalRequest 创建提现申请请求
type CreateWithdrawalRequest struct {
	CurrencyCode   string  `json:"currency_code" binding:"required" example:"NGN"`
	AmountLocal    float64 `json:"amount_local" binding:"required,gt=0" example:"22000.00"`
	BankAccountID  string  `json:"bank_account_id" binding:"required,uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	TransactionPin string  `json:"transaction_pin" binding:"required,len=6,numeric" example:"123456"`
	Description    *string `json:"description" binding:"omitempty" example:"Salary withdrawal"`
}

// AddBankAccountRequest 添加银行账户请求
type AddBankAccountRequest struct {
	BankID        string  `json:"bank_id" binding:"required,uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	AccountNumber string  `json:"account_number" binding:"required" example:"1234567890"`
	AccountName   string  `json:"account_name" binding:"required" example:"John Doe"`
	AccountType   string  `json:"account_type" binding:"omitempty,oneof=savings current checking" example:"savings"`
	SortCode      *string `json:"sort_code" binding:"omitempty" example:"123456"`
	IsDefault     bool    `json:"is_default" example:"false"`
}

// UpdateBankAccountRequest 更新银行账户请求
type UpdateBankAccountRequest struct {
	AccountName string  `json:"account_name" binding:"omitempty" example:"John Doe"`
	AccountType string  `json:"account_type" binding:"omitempty,oneof=savings current checking" example:"savings"`
	SortCode    *string `json:"sort_code" binding:"omitempty" example:"123456"`
	IsDefault   *bool   `json:"is_default" example:"true"`
}

// GetTransactionsRequest 获取交易记录请求
type GetTransactionsRequest struct {
	Page     int     `form:"page" binding:"omitempty,min=1" example:"1"`
	PageSize int     `form:"page_size" binding:"omitempty,min=1,max=100" example:"20"`
	Type     *string `form:"type" binding:"omitempty" example:"withdrawal"`
	Status   *string `form:"status" binding:"omitempty" example:"completed"`
	DateFrom string  `form:"date_from" binding:"omitempty" example:"2024-01-01"`
	DateTo   string  `form:"date_to" binding:"omitempty" example:"2024-12-31"`
	SortBy   string  `form:"sort_by" binding:"omitempty,oneof=created_at amount" example:"created_at"`
	SortDir  string  `form:"sort_dir" binding:"omitempty,oneof=asc desc" example:"desc"`
}

// GetWithdrawalsRequest 获取提现申请请求
type GetWithdrawalsRequest struct {
	Page     int     `form:"page" binding:"omitempty,min=1" example:"1"`
	PageSize int     `form:"page_size" binding:"omitempty,min=1,max=100" example:"20"`
	Status   *string `form:"status" binding:"omitempty" example:"pending"`
	DateFrom string  `form:"date_from" binding:"omitempty" example:"2024-01-01"`
	DateTo   string  `form:"date_to" binding:"omitempty" example:"2024-12-31"`
	SortBy   string  `form:"sort_by" binding:"omitempty,oneof=created_at amount_local" example:"created_at"`
	SortDir  string  `form:"sort_dir" binding:"omitempty,oneof=asc desc" example:"desc"`
}

// GetExchangeRateRequest 获取汇率请求
type GetExchangeRateRequest struct {
	FromCurrency string `form:"from" binding:"required" example:"TRU"`
	ToCurrency   string `form:"to" binding:"required" example:"NGN"`
}

// CalculateWithdrawalRequest 计算提现费用请求
type CalculateWithdrawalRequest struct {
	CurrencyCode string  `json:"currency_code" binding:"required" example:"NGN"`
	AmountLocal  float64 `json:"amount_local" binding:"required,gt=0" example:"22000.00"`
}

// === 管理员请求DTO ===

// AdminReviewWithdrawalRequest 管理员审核提现请求
type AdminReviewWithdrawalRequest struct {
	Action string  `json:"action" binding:"required,oneof=approve reject" example:"approve"`
	Notes  *string `json:"notes" binding:"omitempty" example:"Approved after verification"`
}

// AdminProcessWithdrawalRequest 管理员处理提现请求
type AdminProcessWithdrawalRequest struct {
	TransactionReference string  `json:"transaction_reference" binding:"required" example:"TXN123456789"`
	Notes                *string `json:"notes" binding:"omitempty" example:"Payment processed successfully"`
}

// AdminUpdateExchangeRateRequest 管理员更新汇率请求
type AdminUpdateExchangeRateRequest struct {
	FromCurrencyCode string     `json:"from_currency_code" binding:"required" example:"TRU"`
	ToCurrencyCode   string     `json:"to_currency_code" binding:"required" example:"NGN"`
	Rate             float64    `json:"rate" binding:"required,gt=0" example:"220.00"`
	EffectiveFrom    *time.Time `json:"effective_from" binding:"omitempty" example:"2024-01-01T00:00:00Z"`
	Notes            *string    `json:"notes" binding:"omitempty" example:"Updated market rate"`
}

// AdminWalletAdjustmentRequest 管理员钱包调整请求
type AdminWalletAdjustmentRequest struct {
	UserID      string  `json:"user_id" binding:"required,uuid" example:"123e4567-e89b-12d3-a456-426614174000"`
	Amount      float64 `json:"amount" binding:"required" example:"100.00"`
	Type        string  `json:"type" binding:"required,oneof=adjustment bonus refund" example:"bonus"`
	Description string  `json:"description" binding:"required" example:"Welcome bonus"`
	Reason      string  `json:"reason" binding:"required" example:"User promotion"`
}

// === 响应DTO ===

// WalletResponse 钱包响应
type WalletResponse struct {
	ID                   string     `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Balance              float64    `json:"balance" example:"500.00"`
	FrozenBalance        float64    `json:"frozen_balance" example:"0.00"`
	AvailableBalance     float64    `json:"available_balance" example:"500.00"`
	Status               string     `json:"status" example:"active"`
	IsWithdrawalEnabled  bool       `json:"is_withdrawal_enabled" example:"false"`
	HasTransactionPin    bool       `json:"has_transaction_pin" example:"false"`
	DailyWithdrawalLimit float64    `json:"daily_withdrawal_limit" example:"100000.00"`
	DailyWithdrawnAmount float64    `json:"daily_withdrawn_amount" example:"0.00"`
	RemainingDailyLimit  float64    `json:"remaining_daily_limit" example:"100000.00"`
	WithdrawalCount      int        `json:"withdrawal_count" example:"0"`
	TotalDeposited       float64    `json:"total_deposited" example:"500.00"`
	TotalWithdrawn       float64    `json:"total_withdrawn" example:"0.00"`
	LastTransactionAt    *time.Time `json:"last_transaction_at" example:"2024-01-22T10:30:00Z"`
	CreatedAt            time.Time  `json:"created_at" example:"2024-01-01T08:00:00Z"`
}

// CurrencyResponse 货币响应
type CurrencyResponse struct {
	ID            string  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Code          string  `json:"code" example:"NGN"`
	Name          string  `json:"name" example:"Nigerian Naira"`
	Symbol        *string `json:"symbol" example:"₦"`
	IsFiat        bool    `json:"is_fiat" example:"true"`
	DecimalPlaces int     `json:"decimal_places" example:"2"`
}

// ExchangeRateResponse 汇率响应
type ExchangeRateResponse struct {
	FromCurrency   CurrencyResponse `json:"from_currency"`
	ToCurrency     CurrencyResponse `json:"to_currency"`
	Rate           float64          `json:"rate" example:"220.00"`
	EffectiveFrom  time.Time        `json:"effective_from" example:"2024-01-01T00:00:00Z"`
	EffectiveUntil *time.Time       `json:"effective_until" example:"2024-12-31T23:59:59Z"`
}

// BankResponse 银行响应
type BankResponse struct {
	ID           string           `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name         string           `json:"name" example:"Access Bank"`
	Code         string           `json:"code" example:"ACCESS"`
	CountryCode  string           `json:"country_code" example:"NGA"`
	Currency     CurrencyResponse `json:"currency"`
	SwiftCode    *string          `json:"swift_code" example:"ABNGNGLA"`
	LogoURL      *string          `json:"logo_url" example:"https://example.com/logo.png"`
	SupportPhone *string          `json:"support_phone" example:"+234-800-000-0000"`
}

// BankAccountResponse 银行账户响应
type BankAccountResponse struct {
	ID            string       `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Bank          BankResponse `json:"bank"`
	AccountNumber string       `json:"account_number" example:"1234567890"`
	AccountName   string       `json:"account_name" example:"John Doe"`
	AccountType   string       `json:"account_type" example:"savings"`
	Status        string       `json:"status" example:"active"`
	IsDefault     bool         `json:"is_default" example:"true"`
	IsVerified    bool         `json:"is_verified" example:"false"`
	UsageCount    int          `json:"usage_count" example:"0"`
	LastUsedAt    *time.Time   `json:"last_used_at" example:"2024-01-22T10:30:00Z"`
	CreatedAt     time.Time    `json:"created_at" example:"2024-01-01T08:00:00Z"`
}

// TransactionResponse 交易响应
type TransactionResponse struct {
	ID             string            `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Type           string            `json:"type" example:"withdrawal"`
	Status         string            `json:"status" example:"completed"`
	Amount         float64           `json:"amount" example:"100.00"`
	Fee            float64           `json:"fee" example:"5.00"`
	NetAmount      float64           `json:"net_amount" example:"95.00"`
	BalanceBefore  float64           `json:"balance_before" example:"500.00"`
	BalanceAfter   float64           `json:"balance_after" example:"405.00"`
	Currency       *CurrencyResponse `json:"currency,omitempty"`
	ExchangeRate   *float64          `json:"exchange_rate,omitempty" example:"220.00"`
	OriginalAmount *float64          `json:"original_amount,omitempty" example:"22000.00"`
	Description    *string           `json:"description,omitempty" example:"Salary withdrawal"`
	ProcessedAt    *time.Time        `json:"processed_at,omitempty" example:"2024-01-22T10:30:00Z"`
	CreatedAt      time.Time         `json:"created_at" example:"2024-01-22T10:00:00Z"`
}

// WithdrawalResponse 提现申请响应
type WithdrawalResponse struct {
	ID                   string              `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	AmountTRU            float64             `json:"amount_tru" example:"100.00"`
	AmountLocal          float64             `json:"amount_local" example:"22000.00"`
	Currency             CurrencyResponse    `json:"currency"`
	ExchangeRate         float64             `json:"exchange_rate" example:"220.00"`
	FeeTRU               float64             `json:"fee_tru" example:"5.00"`
	NetAmountTRU         float64             `json:"net_amount_tru" example:"105.00"`
	Status               string              `json:"status" example:"pending"`
	BankAccount          BankAccountResponse `json:"bank_account"`
	TransactionReference *string             `json:"transaction_reference,omitempty" example:"TXN123456789"`
	ReviewNotes          *string             `json:"review_notes,omitempty" example:"Approved after verification"`
	ProcessingNotes      *string             `json:"processing_notes,omitempty" example:"Payment processed"`
	FailureReason        *string             `json:"failure_reason,omitempty" example:"Bank error"`
	RejectionReason      *string             `json:"rejection_reason,omitempty" example:"Insufficient verification"`
	ExpiresAt            time.Time           `json:"expires_at" example:"2024-01-29T10:00:00Z"`
	CreatedAt            time.Time           `json:"created_at" example:"2024-01-22T10:00:00Z"`
	ReviewedAt           *time.Time          `json:"reviewed_at,omitempty" example:"2024-01-22T11:00:00Z"`
	ProcessedAt          *time.Time          `json:"processed_at,omitempty" example:"2024-01-22T12:00:00Z"`
	CompletedAt          *time.Time          `json:"completed_at,omitempty" example:"2024-01-22T13:00:00Z"`
}

// WithdrawalCalculationResponse 提现费用计算响应
type WithdrawalCalculationResponse struct {
	AmountTRU    float64          `json:"amount_tru" example:"100.00"`
	AmountLocal  float64          `json:"amount_local" example:"22000.00"`
	Currency     CurrencyResponse `json:"currency"`
	ExchangeRate float64          `json:"exchange_rate" example:"220.00"`
	FeeTRU       float64          `json:"fee_tru" example:"5.00"`
	NetAmountTRU float64          `json:"net_amount_tru" example:"105.00"`
	CanWithdraw  bool             `json:"can_withdraw" example:"true"`
	ErrorMessage *string          `json:"error_message,omitempty" example:"Insufficient balance"`
}

// === 分页响应DTO ===

// TransactionListResponse 交易列表响应
type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Total        int64                 `json:"total" example:"100"`
	Page         int                   `json:"page" example:"1"`
	PageSize     int                   `json:"page_size" example:"20"`
	TotalPages   int                   `json:"total_pages" example:"5"`
	HasNext      bool                  `json:"has_next" example:"true"`
	HasPrev      bool                  `json:"has_prev" example:"false"`
}

// WithdrawalListResponse 提现申请列表响应
type WithdrawalListResponse struct {
	Withdrawals []WithdrawalResponse `json:"withdrawals"`
	Total       int64                `json:"total" example:"50"`
	Page        int                  `json:"page" example:"1"`
	PageSize    int                  `json:"page_size" example:"20"`
	TotalPages  int                  `json:"total_pages" example:"3"`
	HasNext     bool                 `json:"has_next" example:"true"`
	HasPrev     bool                 `json:"has_prev" example:"false"`
}

// BankAccountListResponse 银行账户列表响应
type BankAccountListResponse struct {
	BankAccounts []BankAccountResponse `json:"bank_accounts"`
	Total        int64                 `json:"total" example:"5"`
}

// BankListResponse 银行列表响应
type BankListResponse struct {
	Banks []BankResponse `json:"banks"`
	Total int64          `json:"total" example:"20"`
}

// CurrencyListResponse 货币列表响应
type CurrencyListResponse struct {
	Currencies []CurrencyResponse `json:"currencies"`
	Total      int64              `json:"total" example:"6"`
}

// === 统计响应DTO ===

// WalletStatisticsResponse 钱包统计响应
type WalletStatisticsResponse struct {
	TotalWallets         int64     `json:"total_wallets" example:"1000"`
	ActiveWallets        int64     `json:"active_wallets" example:"950"`
	TotalBalance         float64   `json:"total_balance" example:"500000.00"`
	TotalFrozenBalance   float64   `json:"total_frozen_balance" example:"1000.00"`
	TotalDeposited       float64   `json:"total_deposited" example:"1000000.00"`
	TotalWithdrawn       float64   `json:"total_withdrawn" example:"500000.00"`
	PendingWithdrawals   int64     `json:"pending_withdrawals" example:"25"`
	CompletedWithdrawals int64     `json:"completed_withdrawals" example:"1500"`
	GeneratedAt          time.Time `json:"generated_at" example:"2024-01-22T15:30:00Z"`
}

// === 通用响应DTO ===

// OperationResponse 操作响应
type OperationResponse struct {
	Success   bool        `json:"success" example:"true"`
	Message   string      `json:"message" example:"Operation completed successfully"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp" example:"2024-01-22T10:15:00Z"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error     string    `json:"error" example:"Validation failed"`
	Message   string    `json:"message" example:"Invalid input data"`
	Details   *string   `json:"details,omitempty" example:"Pin must be 6 digits"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-22T10:15:00Z"`
}

// === 验证和转换方法 ===

// Validate 验证设置交易密码请求
func (req *SetTransactionPinRequest) Validate() error {
	if req.Pin != req.ConfirmPin {
		return fmt.Errorf("pin and confirm_pin do not match")
	}
	return nil
}

// Validate 验证修改交易密码请求
func (req *ChangeTransactionPinRequest) Validate() error {
	if req.NewPin != req.ConfirmPin {
		return fmt.Errorf("new_pin and confirm_pin do not match")
	}
	if req.CurrentPin == req.NewPin {
		return fmt.Errorf("new_pin must be different from current_pin")
	}
	return nil
}

// ToWalletResponse 将钱包模型转换为响应
func (w *Wallet) ToWalletResponse() *WalletResponse {
	return &WalletResponse{
		ID:                   w.ID,
		Balance:              w.Balance,
		FrozenBalance:        w.FrozenBalance,
		AvailableBalance:     w.AvailableBalance(),
		Status:               string(w.Status),
		IsWithdrawalEnabled:  w.IsWithdrawalEnabled,
		HasTransactionPin:    w.TransactionPinHash != nil,
		DailyWithdrawalLimit: w.DailyWithdrawalLimit,
		DailyWithdrawnAmount: w.DailyWithdrawnAmount,
		RemainingDailyLimit:  w.DailyWithdrawalLimit - w.DailyWithdrawnAmount,
		WithdrawalCount:      w.WithdrawalCount,
		TotalDeposited:       w.TotalDeposited,
		TotalWithdrawn:       w.TotalWithdrawn,
		LastTransactionAt:    w.LastTransactionAt,
		CreatedAt:            w.CreatedAt,
	}
}

// ToCurrencyResponse 将货币模型转换为响应
func (c *Currency) ToCurrencyResponse() *CurrencyResponse {
	return &CurrencyResponse{
		ID:            c.ID,
		Code:          c.Code,
		Name:          c.Name,
		Symbol:        c.Symbol,
		IsFiat:        c.IsFiat,
		DecimalPlaces: c.DecimalPlaces,
	}
}

// ToBankResponse 将银行模型转换为响应
func (b *Bank) ToBankResponse() *BankResponse {
	resp := &BankResponse{
		ID:           b.ID,
		Name:         b.Name,
		Code:         b.Code,
		CountryCode:  b.CountryCode,
		SwiftCode:    b.SwiftCode,
		LogoURL:      b.LogoURL,
		SupportPhone: b.SupportPhone,
	}

	if b.Currency != nil {
		resp.Currency = *b.Currency.ToCurrencyResponse()
	}

	return resp
}

// ToBankAccountResponse 将银行账户模型转换为响应
func (ba *UserBankAccount) ToBankAccountResponse() *BankAccountResponse {
	resp := &BankAccountResponse{
		ID:            ba.ID,
		AccountNumber: ba.AccountNumber,
		AccountName:   ba.AccountName,
		AccountType:   ba.AccountType,
		Status:        string(ba.Status),
		IsDefault:     ba.IsDefault,
		IsVerified:    ba.IsVerified,
		UsageCount:    ba.UsageCount,
		LastUsedAt:    ba.LastUsedAt,
		CreatedAt:     ba.CreatedAt,
	}

	if ba.Bank != nil {
		resp.Bank = *ba.Bank.ToBankResponse()
	}

	return resp
}

// ToTransactionResponse 将交易模型转换为响应
func (t *WalletTransaction) ToTransactionResponse() *TransactionResponse {
	resp := &TransactionResponse{
		ID:             t.ID,
		Type:           string(t.Type),
		Status:         string(t.Status),
		Amount:         t.Amount,
		Fee:            t.Fee,
		NetAmount:      t.NetAmount,
		BalanceBefore:  t.BalanceBefore,
		BalanceAfter:   t.BalanceAfter,
		ExchangeRate:   t.ExchangeRate,
		OriginalAmount: t.OriginalAmount,
		Description:    t.Description,
		ProcessedAt:    t.ProcessedAt,
		CreatedAt:      t.CreatedAt,
	}

	if t.Currency != nil {
		resp.Currency = t.Currency.ToCurrencyResponse()
	}

	return resp
}

// ToWithdrawalResponse 将提现申请模型转换为响应
func (wr *WithdrawalRequest) ToWithdrawalResponse() *WithdrawalResponse {
	resp := &WithdrawalResponse{
		ID:                   wr.ID,
		AmountTRU:            wr.AmountTRU,
		AmountLocal:          wr.AmountLocal,
		ExchangeRate:         wr.ExchangeRate,
		FeeTRU:               wr.FeeTRU,
		NetAmountTRU:         wr.NetAmountTRU,
		Status:               string(wr.Status),
		TransactionReference: wr.TransactionReference,
		ReviewNotes:          wr.ReviewNotes,
		ProcessingNotes:      wr.ProcessingNotes,
		FailureReason:        wr.FailureReason,
		RejectionReason:      wr.RejectionReason,
		ExpiresAt:            wr.ExpiresAt,
		CreatedAt:            wr.CreatedAt,
		ReviewedAt:           wr.ReviewedAt,
		ProcessedAt:          wr.ProcessedAt,
		CompletedAt:          wr.CompletedAt,
	}

	if wr.Currency != nil {
		resp.Currency = *wr.Currency.ToCurrencyResponse()
	}

	if wr.BankAccount != nil {
		resp.BankAccount = *wr.BankAccount.ToBankAccountResponse()
	}

	return resp
}
