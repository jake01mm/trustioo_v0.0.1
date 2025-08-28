package wallet

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"trusioo_api_v0.0.1/internal/infrastructure/database"
	"trusioo_api_v0.0.1/pkg/cryptoutil"

	"github.com/sirupsen/logrus"
)

// Repository 钱包数据访问层接口
type Repository interface {
	// 钱包相关
	GetWalletByUserID(ctx context.Context, userID string) (*Wallet, error)
	GetWalletByID(ctx context.Context, walletID string) (*Wallet, error)
	UpdateWallet(ctx context.Context, wallet *Wallet) error
	SetTransactionPin(ctx context.Context, userID, pinHash string) error
	VerifyTransactionPin(ctx context.Context, userID, pinHash string) error

	// 货币相关
	GetCurrencies(ctx context.Context, isActive bool) ([]*Currency, error)
	GetCurrencyByCode(ctx context.Context, code string) (*Currency, error)
	GetCurrencyByID(ctx context.Context, id string) (*Currency, error)

	// 汇率相关
	GetExchangeRate(ctx context.Context, fromCurrencyID, toCurrencyID string) (*ExchangeRate, error)
	GetExchangeRateByCode(ctx context.Context, fromCode, toCode string) (*ExchangeRate, error)
	CreateExchangeRate(ctx context.Context, rate *ExchangeRate) error
	UpdateExchangeRate(ctx context.Context, rate *ExchangeRate) error

	// 银行相关
	GetBanks(ctx context.Context, countryCode string) ([]*Bank, error)
	GetBankByID(ctx context.Context, bankID string) (*Bank, error)

	// 银行账户相关
	GetUserBankAccounts(ctx context.Context, userID string) ([]*UserBankAccount, error)
	GetBankAccountByID(ctx context.Context, accountID string) (*UserBankAccount, error)
	CreateBankAccount(ctx context.Context, account *UserBankAccount) error
	UpdateBankAccount(ctx context.Context, account *UserBankAccount) error
	DeleteBankAccount(ctx context.Context, accountID string) error

	// 交易相关
	CreateTransaction(ctx context.Context, tx *WalletTransaction) error
	GetTransactionByID(ctx context.Context, transactionID string) (*WalletTransaction, error)
	GetUserTransactions(ctx context.Context, userID string, filter *TransactionFilter) ([]*WalletTransaction, int64, error)
	UpdateTransaction(ctx context.Context, tx *WalletTransaction) error

	// 提现相关
	CreateWithdrawalRequest(ctx context.Context, req *WithdrawalRequest) error
	GetWithdrawalByID(ctx context.Context, withdrawalID string) (*WithdrawalRequest, error)
	GetUserWithdrawals(ctx context.Context, userID string, filter *WithdrawalFilter) ([]*WithdrawalRequest, int64, error)
	GetPendingWithdrawals(ctx context.Context, filter *WithdrawalFilter) ([]*WithdrawalRequest, int64, error)
	UpdateWithdrawalRequest(ctx context.Context, req *WithdrawalRequest) error

	// 统计相关
	GetWalletStatistics(ctx context.Context) (*WalletStatistics, error)
	GetTransactionStatistics(ctx context.Context) (*TransactionStatistics, error)
}

// repository 钱包数据访问层实现
type repository struct {
	db     *database.Database
	logger *logrus.Logger
}

// NewRepository 创建新的钱包数据访问层
func NewRepository(db *database.Database, logger *logrus.Logger) Repository {
	return &repository{
		db:     db,
		logger: logger,
	}
}

// === 过滤器结构体 ===

// TransactionFilter 交易过滤器
type TransactionFilter struct {
	Type     *TransactionType   `json:"type"`
	Status   *TransactionStatus `json:"status"`
	DateFrom *time.Time         `json:"date_from"`
	DateTo   *time.Time         `json:"date_to"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
	SortBy   string             `json:"sort_by"`
	SortDir  string             `json:"sort_dir"`
}

// WithdrawalFilter 提现过滤器
type WithdrawalFilter struct {
	Status   *WithdrawalStatus `json:"status"`
	DateFrom *time.Time        `json:"date_from"`
	DateTo   *time.Time        `json:"date_to"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	SortBy   string            `json:"sort_by"`
	SortDir  string            `json:"sort_dir"`
}

// === 钱包相关实现 ===

// GetWalletByUserID 根据用户ID获取钱包
func (r *repository) GetWalletByUserID(ctx context.Context, userID string) (*Wallet, error) {
	query := `
		SELECT id, user_id, balance, frozen_balance, status, is_withdrawal_enabled,
			   transaction_pin_hash, pin_attempts, pin_locked_until, max_pin_attempts,
			   last_transaction_at, daily_withdrawal_limit, daily_withdrawn_amount,
			   last_withdrawal_reset, withdrawal_count, total_deposited, total_withdrawn,
			   notes, created_at, updated_at
		FROM wallets 
		WHERE user_id = $1`

	var wallet Wallet
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.FrozenBalance,
		&wallet.Status, &wallet.IsWithdrawalEnabled, &wallet.TransactionPinHash,
		&wallet.PinAttempts, &wallet.PinLockedUntil, &wallet.MaxPinAttempts,
		&wallet.LastTransactionAt, &wallet.DailyWithdrawalLimit, &wallet.DailyWithdrawnAmount,
		&wallet.LastWithdrawalReset, &wallet.WithdrawalCount, &wallet.TotalDeposited,
		&wallet.TotalWithdrawn, &wallet.Notes, &wallet.CreatedAt, &wallet.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found for user %s", userID)
		}
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to get wallet by user ID")
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return &wallet, nil
}

// GetWalletByID 根据钱包ID获取钱包
func (r *repository) GetWalletByID(ctx context.Context, walletID string) (*Wallet, error) {
	query := `
		SELECT id, user_id, balance, frozen_balance, status, is_withdrawal_enabled,
			   transaction_pin_hash, pin_attempts, pin_locked_until, max_pin_attempts,
			   last_transaction_at, daily_withdrawal_limit, daily_withdrawn_amount,
			   last_withdrawal_reset, withdrawal_count, total_deposited, total_withdrawn,
			   notes, created_at, updated_at
		FROM wallets 
		WHERE id = $1`

	var wallet Wallet
	err := r.db.QueryRowContext(ctx, query, walletID).Scan(
		&wallet.ID, &wallet.UserID, &wallet.Balance, &wallet.FrozenBalance,
		&wallet.Status, &wallet.IsWithdrawalEnabled, &wallet.TransactionPinHash,
		&wallet.PinAttempts, &wallet.PinLockedUntil, &wallet.MaxPinAttempts,
		&wallet.LastTransactionAt, &wallet.DailyWithdrawalLimit, &wallet.DailyWithdrawnAmount,
		&wallet.LastWithdrawalReset, &wallet.WithdrawalCount, &wallet.TotalDeposited,
		&wallet.TotalWithdrawn, &wallet.Notes, &wallet.CreatedAt, &wallet.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("wallet not found with ID %s", walletID)
		}
		r.logger.WithError(err).WithField("wallet_id", walletID).Error("Failed to get wallet by ID")
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return &wallet, nil
}

// UpdateWallet 更新钱包
func (r *repository) UpdateWallet(ctx context.Context, wallet *Wallet) error {
	query := `
		UPDATE wallets SET
			balance = $2, frozen_balance = $3, status = $4, is_withdrawal_enabled = $5,
			transaction_pin_hash = $6, pin_attempts = $7, pin_locked_until = $8,
			last_transaction_at = $9, daily_withdrawal_limit = $10, daily_withdrawn_amount = $11,
			last_withdrawal_reset = $12, withdrawal_count = $13, total_deposited = $14,
			total_withdrawn = $15, notes = $16, updated_at = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		wallet.ID, wallet.Balance, wallet.FrozenBalance, wallet.Status,
		wallet.IsWithdrawalEnabled, wallet.TransactionPinHash, wallet.PinAttempts,
		wallet.PinLockedUntil, wallet.LastTransactionAt, wallet.DailyWithdrawalLimit,
		wallet.DailyWithdrawnAmount, wallet.LastWithdrawalReset, wallet.WithdrawalCount,
		wallet.TotalDeposited, wallet.TotalWithdrawn, wallet.Notes,
	)

	if err != nil {
		r.logger.WithError(err).WithField("wallet_id", wallet.ID).Error("Failed to update wallet")
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	return nil
}

// SetTransactionPin 设置交易密码
func (r *repository) SetTransactionPin(ctx context.Context, userID, pinHash string) error {
	query := `
		UPDATE wallets SET
			transaction_pin_hash = $2, is_withdrawal_enabled = true,
			pin_attempts = 0, pin_locked_until = NULL, updated_at = NOW()
		WHERE user_id = $1`

	result, err := r.db.ExecContext(ctx, query, userID, pinHash)
	if err != nil {
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to set transaction pin")
		return fmt.Errorf("failed to set transaction pin: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("wallet not found for user %s", userID)
	}

	return nil
}

// VerifyTransactionPin 验证交易密码
func (r *repository) VerifyTransactionPin(ctx context.Context, userID, pin string) error {
	query := `
		SELECT transaction_pin_hash, pin_attempts, pin_locked_until, max_pin_attempts
		FROM wallets 
		WHERE user_id = $1`

	var storedHash sql.NullString
	var attempts int
	var lockedUntil sql.NullTime
	var maxAttempts int

	err := r.db.QueryRowContext(ctx, query, userID).Scan(&storedHash, &attempts, &lockedUntil, &maxAttempts)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("wallet not found for user %s", userID)
		}
		r.logger.WithError(err).WithField("user_id", userID).Error("Failed to query transaction pin")
		return fmt.Errorf("failed to verify transaction pin: %w", err)
	}

	// 检查是否设置了交易密码
	if !storedHash.Valid {
		return fmt.Errorf("transaction pin not set")
	}

	// 检查是否被锁定
	if lockedUntil.Valid && lockedUntil.Time.After(time.Now()) {
		return fmt.Errorf("transaction pin is locked until %v", lockedUntil.Time)
	}

	// 创建密码加密器实例来验证密码
	encryptor := cryptoutil.NewPasswordEncryptor("default_key", "bcrypt")
	if err := encryptor.VerifyPassword(pin, storedHash.String); err != nil {
		// 密码错误，增加错误次数
		attempts++
		var newLockedUntil *time.Time

		if attempts >= maxAttempts {
			// 锁定钱包一段时间（比如30分钟）
			lockTime := time.Now().Add(30 * time.Minute)
			newLockedUntil = &lockTime
		}

		updateQuery := `
			UPDATE wallets SET
				pin_attempts = $2, pin_locked_until = $3, updated_at = NOW()
			WHERE user_id = $1`

		_, updateErr := r.db.ExecContext(ctx, updateQuery, userID, attempts, newLockedUntil)
		if updateErr != nil {
			r.logger.WithError(updateErr).WithField("user_id", userID).Error("Failed to update pin attempts")
		}

		if newLockedUntil != nil {
			return fmt.Errorf("transaction pin locked due to too many failed attempts")
		}

		return fmt.Errorf("invalid transaction pin")
	}

	// 密码正确，重置错误次数
	if attempts > 0 {
		resetQuery := `
			UPDATE wallets SET
				pin_attempts = 0, pin_locked_until = NULL, updated_at = NOW()
			WHERE user_id = $1`

		_, err = r.db.ExecContext(ctx, resetQuery, userID)
		if err != nil {
			r.logger.WithError(err).WithField("user_id", userID).Error("Failed to reset pin attempts")
		}
	}

	return nil
}

// === 简化实现其他方法 ===

func (r *repository) GetCurrencies(ctx context.Context, isActive bool) ([]*Currency, error) {
	query := `
		SELECT id, code, name, symbol, is_fiat, is_active, decimal_places, 
			   display_order, description, created_at, updated_at
		FROM currencies 
		WHERE is_active = $1
		ORDER BY display_order ASC, name ASC`

	rows, err := r.db.QueryContext(ctx, query, isActive)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get currencies")
		return nil, fmt.Errorf("failed to get currencies: %w", err)
	}
	defer rows.Close()

	var currencies []*Currency
	for rows.Next() {
		var currency Currency
		err := rows.Scan(
			&currency.ID, &currency.Code, &currency.Name, &currency.Symbol,
			&currency.IsFiat, &currency.IsActive, &currency.DecimalPlaces,
			&currency.DisplayOrder, &currency.Description,
			&currency.CreatedAt, &currency.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan currency row")
			return nil, fmt.Errorf("failed to scan currency: %w", err)
		}
		currencies = append(currencies, &currency)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating currency rows")
		return nil, fmt.Errorf("error iterating currencies: %w", err)
	}

	return currencies, nil
}

func (r *repository) GetCurrencyByCode(ctx context.Context, code string) (*Currency, error) {
	query := `
		SELECT id, code, name, symbol, is_fiat, is_active, decimal_places, 
			   display_order, description, created_at, updated_at
		FROM currencies 
		WHERE code = $1 AND is_active = true`

	var currency Currency
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&currency.ID, &currency.Code, &currency.Name, &currency.Symbol,
		&currency.IsFiat, &currency.IsActive, &currency.DecimalPlaces,
		&currency.DisplayOrder, &currency.Description,
		&currency.CreatedAt, &currency.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("currency not found with code %s", code)
		}
		r.logger.WithError(err).WithField("code", code).Error("Failed to get currency by code")
		return nil, fmt.Errorf("failed to get currency: %w", err)
	}

	return &currency, nil
}

func (r *repository) GetCurrencyByID(ctx context.Context, id string) (*Currency, error) {
	query := `
		SELECT id, code, name, symbol, is_fiat, is_active, decimal_places, 
			   display_order, description, created_at, updated_at
		FROM currencies 
		WHERE id = $1`

	var currency Currency
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&currency.ID, &currency.Code, &currency.Name, &currency.Symbol,
		&currency.IsFiat, &currency.IsActive, &currency.DecimalPlaces,
		&currency.DisplayOrder, &currency.Description,
		&currency.CreatedAt, &currency.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("currency not found with ID %s", id)
		}
		r.logger.WithError(err).WithField("id", id).Error("Failed to get currency by ID")
		return nil, fmt.Errorf("failed to get currency: %w", err)
	}

	return &currency, nil
}

func (r *repository) GetExchangeRate(ctx context.Context, fromCurrencyID, toCurrencyID string) (*ExchangeRate, error) {
	query := `
		SELECT id, from_currency_id, to_currency_id, rate, is_active,
			   effective_from, effective_until, created_by, notes,
			   created_at, updated_at
		FROM exchange_rates 
		WHERE from_currency_id = $1 AND to_currency_id = $2
		  AND is_active = true
		  AND effective_from <= NOW()
		  AND (effective_until IS NULL OR effective_until > NOW())
		ORDER BY effective_from DESC
		LIMIT 1`

	var rate ExchangeRate
	err := r.db.QueryRowContext(ctx, query, fromCurrencyID, toCurrencyID).Scan(
		&rate.ID, &rate.FromCurrencyID, &rate.ToCurrencyID, &rate.Rate, &rate.IsActive,
		&rate.EffectiveFrom, &rate.EffectiveUntil, &rate.CreatedBy, &rate.Notes,
		&rate.CreatedAt, &rate.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("exchange rate not found")
		}
		r.logger.WithError(err).WithFields(logrus.Fields{
			"from_currency_id": fromCurrencyID,
			"to_currency_id":   toCurrencyID,
		}).Error("Failed to get exchange rate")
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	return &rate, nil
}

func (r *repository) GetExchangeRateByCode(ctx context.Context, fromCode, toCode string) (*ExchangeRate, error) {
	fromCurrency, err := r.GetCurrencyByCode(ctx, fromCode)
	if err != nil {
		return nil, fmt.Errorf("from currency not found: %w", err)
	}

	toCurrency, err := r.GetCurrencyByCode(ctx, toCode)
	if err != nil {
		return nil, fmt.Errorf("to currency not found: %w", err)
	}

	rate, err := r.GetExchangeRate(ctx, fromCurrency.ID, toCurrency.ID)
	if err != nil {
		return nil, err
	}

	// 设置关联的货币信息
	rate.FromCurrency = fromCurrency
	rate.ToCurrency = toCurrency

	return rate, nil
}

func (r *repository) CreateExchangeRate(ctx context.Context, rate *ExchangeRate) error {
	return fmt.Errorf("not implemented")
}

func (r *repository) UpdateExchangeRate(ctx context.Context, rate *ExchangeRate) error {
	return fmt.Errorf("not implemented")
}

func (r *repository) GetBanks(ctx context.Context, countryCode string) ([]*Bank, error) {
	query := `
		SELECT b.id, b.name, b.code, b.country_code, b.currency_id, b.swift_code,
			   b.routing_number, b.is_active, b.logo_url, b.website_url, b.support_phone,
			   b.support_email, b.description, b.created_at, b.updated_at,
			   c.id as "currency.id", c.code as "currency.code", c.name as "currency.name",
			   c.symbol as "currency.symbol", c.is_fiat as "currency.is_fiat",
			   c.decimal_places as "currency.decimal_places"
		FROM banks b
		JOIN currencies c ON b.currency_id = c.id
		WHERE b.is_active = true`

	args := []interface{}{}
	if countryCode != "" {
		query += " AND b.country_code = $1"
		args = append(args, countryCode)
	}

	query += " ORDER BY b.name ASC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.WithError(err).Error("Failed to get banks")
		return nil, fmt.Errorf("failed to get banks: %w", err)
	}
	defer rows.Close()

	var banks []*Bank
	for rows.Next() {
		var bank Bank
		var currency Currency

		err := rows.Scan(
			&bank.ID, &bank.Name, &bank.Code, &bank.CountryCode, &bank.CurrencyID,
			&bank.SwiftCode, &bank.RoutingNumber, &bank.IsActive, &bank.LogoURL,
			&bank.WebsiteURL, &bank.SupportPhone, &bank.SupportEmail, &bank.Description,
			&bank.CreatedAt, &bank.UpdatedAt,
			&currency.ID, &currency.Code, &currency.Name, &currency.Symbol,
			&currency.IsFiat, &currency.DecimalPlaces,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan bank row")
			return nil, fmt.Errorf("failed to scan bank: %w", err)
		}

		bank.Currency = &currency
		banks = append(banks, &bank)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating bank rows")
		return nil, fmt.Errorf("error iterating banks: %w", err)
	}

	return banks, nil
}

func (r *repository) GetBankByID(ctx context.Context, bankID string) (*Bank, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *repository) GetUserBankAccounts(ctx context.Context, userID string) ([]*UserBankAccount, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *repository) GetBankAccountByID(ctx context.Context, accountID string) (*UserBankAccount, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *repository) CreateBankAccount(ctx context.Context, account *UserBankAccount) error {
	return fmt.Errorf("not implemented")
}

func (r *repository) UpdateBankAccount(ctx context.Context, account *UserBankAccount) error {
	return fmt.Errorf("not implemented")
}

func (r *repository) DeleteBankAccount(ctx context.Context, accountID string) error {
	return fmt.Errorf("not implemented")
}

func (r *repository) CreateTransaction(ctx context.Context, tx *WalletTransaction) error {
	return fmt.Errorf("not implemented")
}

func (r *repository) GetTransactionByID(ctx context.Context, transactionID string) (*WalletTransaction, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *repository) GetUserTransactions(ctx context.Context, userID string, filter *TransactionFilter) ([]*WalletTransaction, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

func (r *repository) UpdateTransaction(ctx context.Context, tx *WalletTransaction) error {
	return fmt.Errorf("not implemented")
}

func (r *repository) CreateWithdrawalRequest(ctx context.Context, req *WithdrawalRequest) error {
	return fmt.Errorf("not implemented")
}

func (r *repository) GetWithdrawalByID(ctx context.Context, withdrawalID string) (*WithdrawalRequest, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *repository) GetUserWithdrawals(ctx context.Context, userID string, filter *WithdrawalFilter) ([]*WithdrawalRequest, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

func (r *repository) GetPendingWithdrawals(ctx context.Context, filter *WithdrawalFilter) ([]*WithdrawalRequest, int64, error) {
	return nil, 0, fmt.Errorf("not implemented")
}

func (r *repository) UpdateWithdrawalRequest(ctx context.Context, req *WithdrawalRequest) error {
	return fmt.Errorf("not implemented")
}

func (r *repository) GetWalletStatistics(ctx context.Context) (*WalletStatistics, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *repository) GetTransactionStatistics(ctx context.Context) (*TransactionStatistics, error) {
	return nil, fmt.Errorf("not implemented")
}
