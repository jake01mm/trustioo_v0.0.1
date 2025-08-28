package wallet

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// === 枚举类型定义 ===

// WalletStatus 钱包状态
type WalletStatus string

const (
	WalletStatusActive    WalletStatus = "active"
	WalletStatusInactive  WalletStatus = "inactive"
	WalletStatusSuspended WalletStatus = "suspended"
	WalletStatusFrozen    WalletStatus = "frozen"
)

// Value 实现 driver.Valuer 接口
func (ws WalletStatus) Value() (driver.Value, error) {
	return string(ws), nil
}

// Scan 实现 sql.Scanner 接口
func (ws *WalletStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		*ws = WalletStatus(v)
		return nil
	case []byte:
		*ws = WalletStatus(v)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into WalletStatus", value)
	}
}

// TransactionType 交易类型
type TransactionType string

const (
	TransactionTypeDeposit     TransactionType = "deposit"
	TransactionTypeWithdrawal  TransactionType = "withdrawal"
	TransactionTypeTransferIn  TransactionType = "transfer_in"
	TransactionTypeTransferOut TransactionType = "transfer_out"
	TransactionTypeBonus       TransactionType = "bonus"
	TransactionTypeRefund      TransactionType = "refund"
	TransactionTypeFee         TransactionType = "fee"
	TransactionTypeAdjustment  TransactionType = "adjustment"
	TransactionTypeFreeze      TransactionType = "freeze"
	TransactionTypeUnfreeze    TransactionType = "unfreeze"
)

// Value 实现 driver.Valuer 接口
func (tt TransactionType) Value() (driver.Value, error) {
	return string(tt), nil
}

// Scan 实现 sql.Scanner 接口
func (tt *TransactionType) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		*tt = TransactionType(v)
		return nil
	case []byte:
		*tt = TransactionType(v)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into TransactionType", value)
	}
}

// TransactionStatus 交易状态
type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "pending"
	TransactionStatusProcessing TransactionStatus = "processing"
	TransactionStatusCompleted  TransactionStatus = "completed"
	TransactionStatusFailed     TransactionStatus = "failed"
	TransactionStatusCancelled  TransactionStatus = "cancelled"
	TransactionStatusExpired    TransactionStatus = "expired"
)

// Value 实现 driver.Valuer 接口
func (ts TransactionStatus) Value() (driver.Value, error) {
	return string(ts), nil
}

// Scan 实现 sql.Scanner 接口
func (ts *TransactionStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		*ts = TransactionStatus(v)
		return nil
	case []byte:
		*ts = TransactionStatus(v)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into TransactionStatus", value)
	}
}

// WithdrawalStatus 提现状态
type WithdrawalStatus string

const (
	WithdrawalStatusPending    WithdrawalStatus = "pending"
	WithdrawalStatusApproved   WithdrawalStatus = "approved"
	WithdrawalStatusProcessing WithdrawalStatus = "processing"
	WithdrawalStatusCompleted  WithdrawalStatus = "completed"
	WithdrawalStatusRejected   WithdrawalStatus = "rejected"
	WithdrawalStatusCancelled  WithdrawalStatus = "cancelled"
	WithdrawalStatusFailed     WithdrawalStatus = "failed"
)

// Value 实现 driver.Valuer 接口
func (ws WithdrawalStatus) Value() (driver.Value, error) {
	return string(ws), nil
}

// Scan 实现 sql.Scanner 接口
func (ws *WithdrawalStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		*ws = WithdrawalStatus(v)
		return nil
	case []byte:
		*ws = WithdrawalStatus(v)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into WithdrawalStatus", value)
	}
}

// BankAccountStatus 银行账户状态
type BankAccountStatus string

const (
	BankAccountStatusActive              BankAccountStatus = "active"
	BankAccountStatusInactive            BankAccountStatus = "inactive"
	BankAccountStatusSuspended           BankAccountStatus = "suspended"
	BankAccountStatusPendingVerification BankAccountStatus = "pending_verification"
)

// Value 实现 driver.Valuer 接口
func (bas BankAccountStatus) Value() (driver.Value, error) {
	return string(bas), nil
}

// Scan 实现 sql.Scanner 接口
func (bas *BankAccountStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		*bas = BankAccountStatus(v)
		return nil
	case []byte:
		*bas = BankAccountStatus(v)
		return nil
	default:
		return fmt.Errorf("cannot scan %T into BankAccountStatus", value)
	}
}

// === 数据模型定义 ===

// Currency 货币模型
type Currency struct {
	ID            string    `json:"id" db:"id"`
	Code          string    `json:"code" db:"code"`
	Name          string    `json:"name" db:"name"`
	Symbol        *string   `json:"symbol" db:"symbol"`
	IsFiat        bool      `json:"is_fiat" db:"is_fiat"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	DecimalPlaces int       `json:"decimal_places" db:"decimal_places"`
	DisplayOrder  int       `json:"display_order" db:"display_order"`
	Description   *string   `json:"description" db:"description"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// ExchangeRate 汇率模型
type ExchangeRate struct {
	ID             string     `json:"id" db:"id"`
	FromCurrencyID string     `json:"from_currency_id" db:"from_currency_id"`
	ToCurrencyID   string     `json:"to_currency_id" db:"to_currency_id"`
	Rate           float64    `json:"rate" db:"rate"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	EffectiveFrom  time.Time  `json:"effective_from" db:"effective_from"`
	EffectiveUntil *time.Time `json:"effective_until" db:"effective_until"`
	CreatedBy      *string    `json:"created_by" db:"created_by"`
	Notes          *string    `json:"notes" db:"notes"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`

	// 关联数据
	FromCurrency *Currency `json:"from_currency,omitempty"`
	ToCurrency   *Currency `json:"to_currency,omitempty"`
}

// Wallet 钱包模型
type Wallet struct {
	ID                   string       `json:"id" db:"id"`
	UserID               string       `json:"user_id" db:"user_id"`
	Balance              float64      `json:"balance" db:"balance"`
	FrozenBalance        float64      `json:"frozen_balance" db:"frozen_balance"`
	Status               WalletStatus `json:"status" db:"status"`
	IsWithdrawalEnabled  bool         `json:"is_withdrawal_enabled" db:"is_withdrawal_enabled"`
	TransactionPinHash   *string      `json:"-" db:"transaction_pin_hash"` // 不返回给前端
	PinAttempts          int          `json:"pin_attempts" db:"pin_attempts"`
	PinLockedUntil       *time.Time   `json:"pin_locked_until" db:"pin_locked_until"`
	MaxPinAttempts       int          `json:"max_pin_attempts" db:"max_pin_attempts"`
	LastTransactionAt    *time.Time   `json:"last_transaction_at" db:"last_transaction_at"`
	DailyWithdrawalLimit float64      `json:"daily_withdrawal_limit" db:"daily_withdrawal_limit"`
	DailyWithdrawnAmount float64      `json:"daily_withdrawn_amount" db:"daily_withdrawn_amount"`
	LastWithdrawalReset  time.Time    `json:"last_withdrawal_reset" db:"last_withdrawal_reset"`
	WithdrawalCount      int          `json:"withdrawal_count" db:"withdrawal_count"`
	TotalDeposited       float64      `json:"total_deposited" db:"total_deposited"`
	TotalWithdrawn       float64      `json:"total_withdrawn" db:"total_withdrawn"`
	Notes                *string      `json:"notes" db:"notes"`
	CreatedAt            time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time    `json:"updated_at" db:"updated_at"`
}

// Bank 银行模型
type Bank struct {
	ID            string    `json:"id" db:"id"`
	Name          string    `json:"name" db:"name"`
	Code          string    `json:"code" db:"code"`
	CountryCode   string    `json:"country_code" db:"country_code"`
	CurrencyID    string    `json:"currency_id" db:"currency_id"`
	SwiftCode     *string   `json:"swift_code" db:"swift_code"`
	RoutingNumber *string   `json:"routing_number" db:"routing_number"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	LogoURL       *string   `json:"logo_url" db:"logo_url"`
	WebsiteURL    *string   `json:"website_url" db:"website_url"`
	SupportPhone  *string   `json:"support_phone" db:"support_phone"`
	SupportEmail  *string   `json:"support_email" db:"support_email"`
	Description   *string   `json:"description" db:"description"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`

	// 关联数据
	Currency *Currency `json:"currency,omitempty"`
}

// UserBankAccount 用户银行账户模型
type UserBankAccount struct {
	ID                 string            `json:"id" db:"id"`
	UserID             string            `json:"user_id" db:"user_id"`
	BankID             string            `json:"bank_id" db:"bank_id"`
	AccountNumber      string            `json:"account_number" db:"account_number"`
	AccountName        string            `json:"account_name" db:"account_name"`
	AccountType        string            `json:"account_type" db:"account_type"`
	SortCode           *string           `json:"sort_code" db:"sort_code"`
	IBAN               *string           `json:"iban" db:"iban"`
	BICCode            *string           `json:"bic_code" db:"bic_code"`
	Status             BankAccountStatus `json:"status" db:"status"`
	IsDefault          bool              `json:"is_default" db:"is_default"`
	IsVerified         bool              `json:"is_verified" db:"is_verified"`
	VerificationMethod *string           `json:"verification_method" db:"verification_method"`
	VerifiedAt         *time.Time        `json:"verified_at" db:"verified_at"`
	VerifiedBy         *string           `json:"verified_by" db:"verified_by"`
	VerificationNotes  *string           `json:"verification_notes" db:"verification_notes"`
	UsageCount         int               `json:"usage_count" db:"usage_count"`
	LastUsedAt         *time.Time        `json:"last_used_at" db:"last_used_at"`
	Notes              *string           `json:"notes" db:"notes"`
	CreatedAt          time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at" db:"updated_at"`

	// 关联数据
	Bank *Bank `json:"bank,omitempty"`
}

// WalletTransaction 钱包交易模型
type WalletTransaction struct {
	ID              string                 `json:"id" db:"id"`
	WalletID        string                 `json:"wallet_id" db:"wallet_id"`
	UserID          string                 `json:"user_id" db:"user_id"`
	Type            TransactionType        `json:"type" db:"type"`
	Status          TransactionStatus      `json:"status" db:"status"`
	Amount          float64                `json:"amount" db:"amount"`
	Fee             float64                `json:"fee" db:"fee"`
	NetAmount       float64                `json:"net_amount" db:"net_amount"`
	BalanceBefore   float64                `json:"balance_before" db:"balance_before"`
	BalanceAfter    float64                `json:"balance_after" db:"balance_after"`
	CurrencyID      *string                `json:"currency_id" db:"currency_id"`
	ExchangeRate    *float64               `json:"exchange_rate" db:"exchange_rate"`
	OriginalAmount  *float64               `json:"original_amount" db:"original_amount"`
	ReferenceID     *string                `json:"reference_id" db:"reference_id"`
	ReferenceType   *string                `json:"reference_type" db:"reference_type"`
	TransactionHash *string                `json:"transaction_hash" db:"transaction_hash"`
	Description     *string                `json:"description" db:"description"`
	Metadata        map[string]interface{} `json:"metadata" db:"metadata"`
	ProcessedAt     *time.Time             `json:"processed_at" db:"processed_at"`
	ProcessedBy     *string                `json:"processed_by" db:"processed_by"`
	ExpiresAt       *time.Time             `json:"expires_at" db:"expires_at"`
	Notes           *string                `json:"notes" db:"notes"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`

	// 关联数据
	Currency *Currency `json:"currency,omitempty"`
}

// WithdrawalRequest 提现申请模型
type WithdrawalRequest struct {
	ID                   string                 `json:"id" db:"id"`
	UserID               string                 `json:"user_id" db:"user_id"`
	WalletID             string                 `json:"wallet_id" db:"wallet_id"`
	BankAccountID        string                 `json:"bank_account_id" db:"bank_account_id"`
	CurrencyID           string                 `json:"currency_id" db:"currency_id"`
	AmountTRU            float64                `json:"amount_tru" db:"amount_tru"`
	AmountLocal          float64                `json:"amount_local" db:"amount_local"`
	ExchangeRate         float64                `json:"exchange_rate" db:"exchange_rate"`
	FeeTRU               float64                `json:"fee_tru" db:"fee_tru"`
	NetAmountTRU         float64                `json:"net_amount_tru" db:"net_amount_tru"`
	Status               WithdrawalStatus       `json:"status" db:"status"`
	Priority             int                    `json:"priority" db:"priority"`
	ReviewedBy           *string                `json:"reviewed_by" db:"reviewed_by"`
	ReviewedAt           *time.Time             `json:"reviewed_at" db:"reviewed_at"`
	ReviewNotes          *string                `json:"review_notes" db:"review_notes"`
	ProcessedBy          *string                `json:"processed_by" db:"processed_by"`
	ProcessedAt          *time.Time             `json:"processed_at" db:"processed_at"`
	ProcessingNotes      *string                `json:"processing_notes" db:"processing_notes"`
	CompletedAt          *time.Time             `json:"completed_at" db:"completed_at"`
	TransactionReference *string                `json:"transaction_reference" db:"transaction_reference"`
	TransactionID        *string                `json:"transaction_id" db:"transaction_id"`
	FailureReason        *string                `json:"failure_reason" db:"failure_reason"`
	RejectionReason      *string                `json:"rejection_reason" db:"rejection_reason"`
	UserName             string                 `json:"user_name" db:"user_name"`
	UserEmail            string                 `json:"user_email" db:"user_email"`
	BankName             string                 `json:"bank_name" db:"bank_name"`
	AccountNumber        string                 `json:"account_number" db:"account_number"`
	AccountName          string                 `json:"account_name" db:"account_name"`
	IPAddress            *string                `json:"ip_address" db:"ip_address"`
	UserAgent            *string                `json:"user_agent" db:"user_agent"`
	ExpiresAt            time.Time              `json:"expires_at" db:"expires_at"`
	Metadata             map[string]interface{} `json:"metadata" db:"metadata"`
	Notes                *string                `json:"notes" db:"notes"`
	CreatedAt            time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time              `json:"updated_at" db:"updated_at"`

	// 关联数据
	Currency    *Currency          `json:"currency,omitempty"`
	BankAccount *UserBankAccount   `json:"bank_account,omitempty"`
	Transaction *WalletTransaction `json:"transaction,omitempty"`
}

// === 统计模型 ===

// WalletStatistics 钱包统计
type WalletStatistics struct {
	TotalWallets         int64     `json:"total_wallets"`
	ActiveWallets        int64     `json:"active_wallets"`
	TotalBalance         float64   `json:"total_balance"`
	TotalFrozenBalance   float64   `json:"total_frozen_balance"`
	TotalDeposited       float64   `json:"total_deposited"`
	TotalWithdrawn       float64   `json:"total_withdrawn"`
	PendingWithdrawals   int64     `json:"pending_withdrawals"`
	CompletedWithdrawals int64     `json:"completed_withdrawals"`
	GeneratedAt          time.Time `json:"generated_at"`
}

// TransactionStatistics 交易统计
type TransactionStatistics struct {
	TotalTransactions      int64     `json:"total_transactions"`
	TodayTransactions      int64     `json:"today_transactions"`
	TotalVolume            float64   `json:"total_volume"`
	TodayVolume            float64   `json:"today_volume"`
	SuccessfulTransactions int64     `json:"successful_transactions"`
	FailedTransactions     int64     `json:"failed_transactions"`
	PendingTransactions    int64     `json:"pending_transactions"`
	GeneratedAt            time.Time `json:"generated_at"`
}

// === 辅助方法 ===

// CanWithdraw 检查钱包是否可以提现
func (w *Wallet) CanWithdraw() bool {
	return w.Status == WalletStatusActive &&
		w.IsWithdrawalEnabled &&
		w.TransactionPinHash != nil &&
		(w.PinLockedUntil == nil || w.PinLockedUntil.Before(time.Now()))
}

// AvailableBalance 获取可用余额
func (w *Wallet) AvailableBalance() float64 {
	return w.Balance - w.FrozenBalance
}

// CanWithdrawAmount 检查是否可以提现指定金额
func (w *Wallet) CanWithdrawAmount(amount float64) bool {
	if !w.CanWithdraw() {
		return false
	}

	// 检查余额是否足够
	if w.AvailableBalance() < amount {
		return false
	}

	// 检查每日限额
	if w.DailyWithdrawnAmount+amount > w.DailyWithdrawalLimit {
		return false
	}

	return true
}

// IsExpired 检查提现申请是否已过期
func (wr *WithdrawalRequest) IsExpired() bool {
	return time.Now().After(wr.ExpiresAt)
}

// CanCancel 检查提现申请是否可以取消
func (wr *WithdrawalRequest) CanCancel() bool {
	return wr.Status == WithdrawalStatusPending && !wr.IsExpired()
}

// CanApprove 检查提现申请是否可以批准
func (wr *WithdrawalRequest) CanApprove() bool {
	return wr.Status == WithdrawalStatusPending && !wr.IsExpired()
}

// CanReject 检查提现申请是否可以拒绝
func (wr *WithdrawalRequest) CanReject() bool {
	return wr.Status == WithdrawalStatusPending && !wr.IsExpired()
}
