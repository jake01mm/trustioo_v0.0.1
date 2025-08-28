package wallet

import (
	"context"
	"fmt"

	"trusioo_api_v0.0.1/pkg/cryptoutil"

	"github.com/sirupsen/logrus"
)

// Service 钱包服务接口
type Service interface {
	// 钱包相关
	GetWallet(ctx context.Context, userID string) (*WalletResponse, error)
	SetTransactionPin(ctx context.Context, userID string, req *SetTransactionPinRequest) error
	ChangeTransactionPin(ctx context.Context, userID string, req *ChangeTransactionPinRequest) error

	// 货币和汇率相关
	GetCurrencies(ctx context.Context) (*CurrencyListResponse, error)
	GetExchangeRate(ctx context.Context, fromCode, toCode string) (*ExchangeRateResponse, error)

	// 银行相关
	GetBanks(ctx context.Context, countryCode string) (*BankListResponse, error)

	// 银行账户相关
	GetUserBankAccounts(ctx context.Context, userID string) (*BankAccountListResponse, error)
	AddBankAccount(ctx context.Context, userID string, req *AddBankAccountRequest) (*BankAccountResponse, error)
	UpdateBankAccount(ctx context.Context, userID, accountID string, req *UpdateBankAccountRequest) (*BankAccountResponse, error)
	DeleteBankAccount(ctx context.Context, userID, accountID string) error

	// 提现相关
	CalculateWithdrawal(ctx context.Context, userID string, req *CalculateWithdrawalRequest) (*WithdrawalCalculationResponse, error)
	CreateWithdrawalRequest(ctx context.Context, userID string, req *CreateWithdrawalRequest) (*WithdrawalResponse, error)
	GetUserWithdrawals(ctx context.Context, userID string, req *GetWithdrawalsRequest) (*WithdrawalListResponse, error)
	CancelWithdrawal(ctx context.Context, userID, withdrawalID string) error

	// 交易相关
	GetUserTransactions(ctx context.Context, userID string, req *GetTransactionsRequest) (*TransactionListResponse, error)

	// 管理员功能
	ReviewWithdrawal(ctx context.Context, adminID, withdrawalID string, req *AdminReviewWithdrawalRequest) error
	ProcessWithdrawal(ctx context.Context, adminID, withdrawalID string, req *AdminProcessWithdrawalRequest) error
	GetPendingWithdrawals(ctx context.Context, req *GetWithdrawalsRequest) (*WithdrawalListResponse, error)
	UpdateExchangeRate(ctx context.Context, adminID string, req *AdminUpdateExchangeRateRequest) error
	AdjustWallet(ctx context.Context, adminID string, req *AdminWalletAdjustmentRequest) error
	GetWalletStatistics(ctx context.Context) (*WalletStatisticsResponse, error)
}

// service 钱包服务实现
type service struct {
	repo      Repository
	encryptor *cryptoutil.PasswordEncryptor
	logger    *logrus.Logger
}

// NewService 创建新的钱包服务
func NewService(repo Repository, encryptor *cryptoutil.PasswordEncryptor, logger *logrus.Logger) Service {
	return &service{
		repo:      repo,
		encryptor: encryptor,
		logger:    logger,
	}
}

// === 钱包相关实现 ===

// GetWallet 获取用户钱包信息
func (s *service) GetWallet(ctx context.Context, userID string) (*WalletResponse, error) {
	wallet, err := s.repo.GetWalletByUserID(ctx, userID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to get wallet")
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return wallet.ToWalletResponse(), nil
}

// SetTransactionPin 设置交易密码
func (s *service) SetTransactionPin(ctx context.Context, userID string, req *SetTransactionPinRequest) error {
	// 验证请求
	if err := req.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 检查钱包是否存在
	wallet, err := s.repo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("wallet not found: %w", err)
	}

	// 检查是否已设置交易密码
	if wallet.TransactionPinHash != nil {
		return fmt.Errorf("transaction pin already set")
	}

	// 加密交易密码
	hashedPin, err := s.encryptor.HashPassword(req.Pin)
	if err != nil {
		s.logger.WithError(err).Error("Failed to hash transaction pin")
		return fmt.Errorf("failed to encrypt pin: %w", err)
	}

	// 保存交易密码
	if err := s.repo.SetTransactionPin(ctx, userID, hashedPin); err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to set transaction pin")
		return fmt.Errorf("failed to set transaction pin: %w", err)
	}

	s.logger.WithField("user_id", userID).Info("Transaction pin set successfully")
	return nil
}

// ChangeTransactionPin 修改交易密码
func (s *service) ChangeTransactionPin(ctx context.Context, userID string, req *ChangeTransactionPinRequest) error {
	// 验证请求
	if err := req.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// 验证当前密码
	if err := s.repo.VerifyTransactionPin(ctx, userID, req.CurrentPin); err != nil {
		return fmt.Errorf("current pin verification failed: %w", err)
	}

	// 加密新密码
	hashedPin, err := s.encryptor.HashPassword(req.NewPin)
	if err != nil {
		s.logger.WithError(err).Error("Failed to hash new transaction pin")
		return fmt.Errorf("failed to encrypt new pin: %w", err)
	}

	// 保存新密码
	if err := s.repo.SetTransactionPin(ctx, userID, hashedPin); err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to change transaction pin")
		return fmt.Errorf("failed to change transaction pin: %w", err)
	}

	s.logger.WithField("user_id", userID).Info("Transaction pin changed successfully")
	return nil
}

// === 货币和汇率相关实现 ===

// GetCurrencies 获取支持的货币列表
func (s *service) GetCurrencies(ctx context.Context) (*CurrencyListResponse, error) {
	currencies, err := s.repo.GetCurrencies(ctx, true)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get currencies")
		return nil, fmt.Errorf("failed to get currencies: %w", err)
	}

	var currencyResponses []CurrencyResponse
	for _, currency := range currencies {
		currencyResponses = append(currencyResponses, *currency.ToCurrencyResponse())
	}

	return &CurrencyListResponse{
		Currencies: currencyResponses,
		Total:      int64(len(currencyResponses)),
	}, nil
}

// GetExchangeRate 获取汇率
func (s *service) GetExchangeRate(ctx context.Context, fromCode, toCode string) (*ExchangeRateResponse, error) {
	rate, err := s.repo.GetExchangeRateByCode(ctx, fromCode, toCode)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"from_code": fromCode,
			"to_code":   toCode,
		}).Error("Failed to get exchange rate")
		return nil, fmt.Errorf("failed to get exchange rate: %w", err)
	}

	response := &ExchangeRateResponse{
		Rate:           rate.Rate,
		EffectiveFrom:  rate.EffectiveFrom,
		EffectiveUntil: rate.EffectiveUntil,
	}

	if rate.FromCurrency != nil {
		response.FromCurrency = *rate.FromCurrency.ToCurrencyResponse()
	}
	if rate.ToCurrency != nil {
		response.ToCurrency = *rate.ToCurrency.ToCurrencyResponse()
	}

	return response, nil
}

// === 银行相关实现 ===

// GetBanks 获取银行列表
func (s *service) GetBanks(ctx context.Context, countryCode string) (*BankListResponse, error) {
	banks, err := s.repo.GetBanks(ctx, countryCode)
	if err != nil {
		s.logger.WithError(err).WithField("country_code", countryCode).Error("Failed to get banks")
		return nil, fmt.Errorf("failed to get banks: %w", err)
	}

	var bankResponses []BankResponse
	for _, bank := range banks {
		bankResponses = append(bankResponses, *bank.ToBankResponse())
	}

	return &BankListResponse{
		Banks: bankResponses,
		Total: int64(len(bankResponses)),
	}, nil
}

// === 银行账户相关实现 ===

// GetUserBankAccounts 获取用户银行账户列表
func (s *service) GetUserBankAccounts(ctx context.Context, userID string) (*BankAccountListResponse, error) {
	accounts, err := s.repo.GetUserBankAccounts(ctx, userID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user bank accounts")
		return nil, fmt.Errorf("failed to get bank accounts: %w", err)
	}

	var accountResponses []BankAccountResponse
	for _, account := range accounts {
		accountResponses = append(accountResponses, *account.ToBankAccountResponse())
	}

	return &BankAccountListResponse{
		BankAccounts: accountResponses,
		Total:        int64(len(accountResponses)),
	}, nil
}

// AddBankAccount 添加银行账户
func (s *service) AddBankAccount(ctx context.Context, userID string, req *AddBankAccountRequest) (*BankAccountResponse, error) {
	// 验证银行是否存在
	bank, err := s.repo.GetBankByID(ctx, req.BankID)
	if err != nil {
		return nil, fmt.Errorf("bank not found: %w", err)
	}

	// 创建银行账户模型
	account := &UserBankAccount{
		UserID:        userID,
		BankID:        req.BankID,
		AccountNumber: req.AccountNumber,
		AccountName:   req.AccountName,
		AccountType:   req.AccountType,
		SortCode:      req.SortCode,
		Status:        BankAccountStatusPendingVerification,
		IsDefault:     req.IsDefault,
		IsVerified:    false,
		UsageCount:    0,
	}

	// 设置默认账户类型
	if account.AccountType == "" {
		account.AccountType = "savings"
	}

	// 创建银行账户
	if err := s.repo.CreateBankAccount(ctx, account); err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to create bank account")
		return nil, fmt.Errorf("failed to create bank account: %w", err)
	}

	// 设置银行信息
	account.Bank = bank

	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"account_id": account.ID,
		"bank_name":  bank.Name,
	}).Info("Bank account created successfully")

	return account.ToBankAccountResponse(), nil
}

// UpdateBankAccount 更新银行账户
func (s *service) UpdateBankAccount(ctx context.Context, userID, accountID string, req *UpdateBankAccountRequest) (*BankAccountResponse, error) {
	// 获取现有账户
	account, err := s.repo.GetBankAccountByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("bank account not found: %w", err)
	}

	// 验证账户所属权
	if account.UserID != userID {
		return nil, fmt.Errorf("account does not belong to user")
	}

	// 更新字段
	if req.AccountName != "" {
		account.AccountName = req.AccountName
	}
	if req.AccountType != "" {
		account.AccountType = req.AccountType
	}
	if req.SortCode != nil {
		account.SortCode = req.SortCode
	}
	if req.IsDefault != nil {
		account.IsDefault = *req.IsDefault
	}

	// 保存更新
	if err := s.repo.UpdateBankAccount(ctx, account); err != nil {
		s.logger.WithError(err).WithField("account_id", accountID).Error("Failed to update bank account")
		return nil, fmt.Errorf("failed to update bank account: %w", err)
	}

	return account.ToBankAccountResponse(), nil
}

// DeleteBankAccount 删除银行账户
func (s *service) DeleteBankAccount(ctx context.Context, userID, accountID string) error {
	// 验证账户存在且属于该用户
	account, err := s.repo.GetBankAccountByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("bank account not found: %w", err)
	}

	if account.UserID != userID {
		return fmt.Errorf("account does not belong to user")
	}

	// 检查是否是默认账户
	if account.IsDefault {
		// 可以考虑是否允许删除默认账户，或者要求先设置其他账户为默认
		return fmt.Errorf("cannot delete default account")
	}

	// 删除账户
	if err := s.repo.DeleteBankAccount(ctx, accountID); err != nil {
		s.logger.WithError(err).WithField("account_id", accountID).Error("Failed to delete bank account")
		return fmt.Errorf("failed to delete bank account: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"account_id": accountID,
	}).Info("Bank account deleted successfully")

	return nil
}

// === 简化实现其他方法 ===

func (s *service) CalculateWithdrawal(ctx context.Context, userID string, req *CalculateWithdrawalRequest) (*WithdrawalCalculationResponse, error) {
	// 简化实现
	return nil, fmt.Errorf("not implemented")
}

func (s *service) CreateWithdrawalRequest(ctx context.Context, userID string, req *CreateWithdrawalRequest) (*WithdrawalResponse, error) {
	// 简化实现
	return nil, fmt.Errorf("not implemented")
}

func (s *service) GetUserWithdrawals(ctx context.Context, userID string, req *GetWithdrawalsRequest) (*WithdrawalListResponse, error) {
	// 简化实现
	return nil, fmt.Errorf("not implemented")
}

func (s *service) CancelWithdrawal(ctx context.Context, userID, withdrawalID string) error {
	// 简化实现
	return fmt.Errorf("not implemented")
}

func (s *service) GetUserTransactions(ctx context.Context, userID string, req *GetTransactionsRequest) (*TransactionListResponse, error) {
	// 简化实现
	return nil, fmt.Errorf("not implemented")
}

func (s *service) ReviewWithdrawal(ctx context.Context, adminID, withdrawalID string, req *AdminReviewWithdrawalRequest) error {
	// 简化实现
	return fmt.Errorf("not implemented")
}

func (s *service) ProcessWithdrawal(ctx context.Context, adminID, withdrawalID string, req *AdminProcessWithdrawalRequest) error {
	// 简化实现
	return fmt.Errorf("not implemented")
}

func (s *service) GetPendingWithdrawals(ctx context.Context, req *GetWithdrawalsRequest) (*WithdrawalListResponse, error) {
	// 简化实现
	return nil, fmt.Errorf("not implemented")
}

func (s *service) UpdateExchangeRate(ctx context.Context, adminID string, req *AdminUpdateExchangeRateRequest) error {
	// 简化实现
	return fmt.Errorf("not implemented")
}

func (s *service) AdjustWallet(ctx context.Context, adminID string, req *AdminWalletAdjustmentRequest) error {
	// 简化实现
	return fmt.Errorf("not implemented")
}

func (s *service) GetWalletStatistics(ctx context.Context) (*WalletStatisticsResponse, error) {
	// 简化实现
	return nil, fmt.Errorf("not implemented")
}
