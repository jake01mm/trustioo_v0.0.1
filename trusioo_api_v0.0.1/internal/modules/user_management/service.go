package user_management

import (
	"context"
	"fmt"
	"time"

	"trusioo_api_v0.0.1/internal/modules/auth"
	"trusioo_api_v0.0.1/internal/modules/auth/user"
	"trusioo_api_v0.0.1/pkg/cryptoutil"

	"github.com/sirupsen/logrus"
)

// Service 用户管理服务
type Service struct {
	repo      *Repository
	userRepo  *user.Repository // 复用用户仓储
	encryptor *cryptoutil.PasswordEncryptor
	logger    *logrus.Logger
}

// NewService 创建新的用户管理服务
func NewService(repo *Repository, userRepo *user.Repository, encryptor *cryptoutil.PasswordEncryptor, logger *logrus.Logger) *Service {
	return &Service{
		repo:      repo,
		userRepo:  userRepo,
		encryptor: encryptor,
		logger:    logger,
	}
}

// === 用户查询服务 ===

// GetUserByID 根据ID获取用户详细信息
func (s *Service) GetUserByID(ctx context.Context, userID string) (*UserManagementModel, error) {
	userModel, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Error("Failed to get user by ID")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return userModel, nil
}

// GetUsers 获取用户列表（支持分页和过滤）
func (s *Service) GetUsers(ctx context.Context, req *GetUsersRequest) (*UserListResponse, error) {
	// 转换请求参数
	filter := req.ToSearchFilter()
	pagination := req.ToPaginationParams()

	var users []*UserManagementModel
	var total int64
	var err error

	// 如果有搜索关键词，使用搜索方法
	if req.Search != "" {
		users, total, err = s.repo.SearchUsers(ctx, req.Search, pagination)
	} else {
		users, total, err = s.repo.GetUsers(ctx, filter, pagination)
	}

	if err != nil {
		s.logger.WithError(err).Error("Failed to get users")
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	// 转换为响应格式
	userSummaries := make([]UserSummaryResponse, 0, len(users))
	for _, userModel := range users {
		// 获取活跃会话数
		sessions, err := s.repo.GetUserSessions(ctx, userModel.ID)
		if err != nil {
			s.logger.WithError(err).WithField("user_id", userModel.ID).Warn("Failed to get user sessions")
		}

		activeSessions := len(sessions)
		userSummaries = append(userSummaries, *userModel.ToUserSummaryResponse(activeSessions))
	}

	// 创建分页结果
	paginatedResult := NewPaginatedResult(userSummaries, total, pagination)

	response := &UserListResponse{
		Users:      userSummaries,
		Total:      paginatedResult.Total,
		Page:       paginatedResult.Page,
		PageSize:   paginatedResult.PageSize,
		TotalPages: paginatedResult.TotalPages,
		HasNext:    paginatedResult.HasNext,
		HasPrev:    paginatedResult.HasPrev,
	}

	return response, nil
}

// GetUserDetail 获取用户详细信息
func (s *Service) GetUserDetail(ctx context.Context, userID string) (*UserDetailResponse, error) {
	userModel, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 获取活跃会话数
	sessions, err := s.repo.GetUserSessions(ctx, userID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Warn("Failed to get user sessions")
	}

	activeSessions := len(sessions)
	response := userModel.ToUserDetailResponse(activeSessions)

	return response, nil
}

// GetUserActivity 获取用户活动信息
func (s *Service) GetUserActivity(ctx context.Context, userID string) (*UserActivityResponse, error) {
	userModel, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// 获取会话信息
	sessions, err := s.repo.GetUserSessions(ctx, userID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Warn("Failed to get user sessions")
		sessions = []SessionSummary{} // 设置为空数组而不是失败
	}

	// 获取最近活动日志
	logs, err := s.repo.GetUserActivityLogs(ctx, userID, 10)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Warn("Failed to get activity logs")
		logs = []UserActivityLogSummary{} // 设置为空数组而不是失败
	}

	// 计算最后活动时间
	var lastActivityAt *time.Time
	if len(sessions) > 0 {
		lastActivityAt = &sessions[0].LastActivity
	}

	response := &UserActivityResponse{
		UserID:              userModel.ID,
		Email:               userModel.Email,
		Name:                userModel.Name,
		LastLoginAt:         userModel.LastLoginAt,
		TotalLogins:         userModel.LoginCount,
		FailedLoginAttempts: userModel.FailedAttempts,
		ActiveSessions:      sessions,
		LastActivityAt:      lastActivityAt,
		RecentLogs:          logs,
	}

	return response, nil
}

// === 统计服务 ===

// GetStatistics 获取用户统计信息
func (s *Service) GetStatistics(ctx context.Context) (*StatisticsResponse, error) {
	stats, err := s.repo.GetUserStatistics(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to get user statistics")
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	// 转换为响应格式
	response := &StatisticsResponse{
		TotalUsers:           stats.TotalUsers,
		NewUsersToday:        stats.NewUsersToday,
		NewUsersThisWeek:     stats.NewUsersThisWeek,
		NewUsersThisMonth:    stats.NewUsersThisMonth,
		VerifiedUsers:        stats.VerifiedUsers,
		UnverifiedUsers:      stats.UnverifiedUsers,
		VerificationRate:     stats.VerificationRate,
		ActiveUsers:          stats.ActiveUsers,
		InactiveUsers:        stats.InactiveUsers,
		SuspendedUsers:       stats.SuspendedUsers,
		DeletedUsers:         stats.DeletedUsers,
		UsersByStatus:        make(map[string]int64),
		ActiveUsersToday:     stats.ActiveUsersToday,
		ActiveUsersThisWeek:  stats.ActiveUsersThisWeek,
		ActiveUsersThisMonth: stats.ActiveUsersThisMonth,
		GeneratedAt:          stats.GeneratedAt,
	}

	// 转换状态映射
	for status, count := range stats.UsersByStatus {
		response.UsersByStatus[status.String()] = count
	}

	// 转换趋势数据
	response.RegistrationTrend = make([]DailyCountResponse, len(stats.RegistrationTrend))
	for i, trend := range stats.RegistrationTrend {
		response.RegistrationTrend[i] = DailyCountResponse{
			Date:  trend.Date,
			Count: trend.Count,
		}
	}

	response.VerificationTrend = make([]DailyCountResponse, len(stats.VerificationTrend))
	for i, trend := range stats.VerificationTrend {
		response.VerificationTrend[i] = DailyCountResponse{
			Date:  trend.Date,
			Count: trend.Count,
		}
	}

	return response, nil
}

// === 用户管理操作服务 ===

// UpdateUserStatus 更新用户状态
func (s *Service) UpdateUserStatus(ctx context.Context, userID, adminID, adminEmail, ipAddress string,
	req *UpdateUserStatusRequest) (*OperationResponse, error) {

	// 验证状态有效性
	status := user.UserStatus(req.Status)
	if !status.IsValid() {
		return nil, fmt.Errorf("invalid user status: %s", req.Status)
	}

	// 获取目标用户信息
	targetUser, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target user: %w", err)
	}

	// 更新用户状态
	err = s.repo.UpdateUserStatus(ctx, userID, status, req.Reason)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":    userID,
			"admin_id":   adminID,
			"new_status": req.Status,
		}).Error("Failed to update user status")
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	// 记录管理操作日志
	action := UserManagementAction("update_status")
	if status == user.UserStatusSuspended {
		action = ActionSuspend
	} else if status == user.UserStatusActive {
		action = ActionActivate
	} else if status == user.UserStatusInactive {
		action = ActionDeactivate
	}

	logEntry := &UserManagementLog{
		AdminID:      adminID,
		AdminEmail:   adminEmail,
		TargetUserID: userID,
		TargetEmail:  targetUser.Email,
		Action:       action,
		Reason:       req.Reason,
		IPAddress:    ipAddress,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateManagementLog(ctx, logEntry); err != nil {
		s.logger.WithError(err).Error("Failed to create management log")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"admin_id":   adminID,
		"old_status": targetUser.Status,
		"new_status": req.Status,
		"reason":     req.Reason,
	}).Info("User status updated by admin")

	return &OperationResponse{
		Success:   true,
		Message:   fmt.Sprintf("User status updated to %s successfully", req.Status),
		Timestamp: time.Now(),
	}, nil
}

// SuspendUser 暂停用户
func (s *Service) SuspendUser(ctx context.Context, userID, adminID, adminEmail, ipAddress string,
	req *SuspendUserRequest) (*OperationResponse, error) {

	// 获取目标用户信息
	targetUser, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target user: %w", err)
	}

	// 暂停用户
	err = s.repo.UpdateUserStatus(ctx, userID, user.UserStatusSuspended, &req.Reason)
	if err != nil {
		return nil, fmt.Errorf("failed to suspend user: %w", err)
	}

	// 强制登出所有会话
	err = s.repo.DeactivateUserSessions(ctx, userID, nil)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Warn("Failed to deactivate user sessions")
	}

	// 记录管理操作日志
	logEntry := &UserManagementLog{
		AdminID:      adminID,
		AdminEmail:   adminEmail,
		TargetUserID: userID,
		TargetEmail:  targetUser.Email,
		Action:       ActionSuspend,
		Reason:       &req.Reason,
		IPAddress:    ipAddress,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateManagementLog(ctx, logEntry); err != nil {
		s.logger.WithError(err).Error("Failed to create management log")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"admin_id": adminID,
		"reason":   req.Reason,
		"duration": req.Duration,
	}).Info("User suspended by admin")

	return &OperationResponse{
		Success:   true,
		Message:   "User suspended successfully",
		Timestamp: time.Now(),
	}, nil
}

// ReactivateUser 重新激活用户
func (s *Service) ReactivateUser(ctx context.Context, userID, adminID, adminEmail, ipAddress string) (*OperationResponse, error) {
	// 获取目标用户信息
	targetUser, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target user: %w", err)
	}

	// 激活用户
	err = s.repo.UpdateUserStatus(ctx, userID, user.UserStatusActive, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to reactivate user: %w", err)
	}

	// 记录管理操作日志
	logEntry := &UserManagementLog{
		AdminID:      adminID,
		AdminEmail:   adminEmail,
		TargetUserID: userID,
		TargetEmail:  targetUser.Email,
		Action:       ActionActivate,
		IPAddress:    ipAddress,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateManagementLog(ctx, logEntry); err != nil {
		s.logger.WithError(err).Error("Failed to create management log")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"admin_id": adminID,
	}).Info("User reactivated by admin")

	return &OperationResponse{
		Success:   true,
		Message:   "User reactivated successfully",
		Timestamp: time.Now(),
	}, nil
}

// ResetUserPassword 重置用户密码
func (s *Service) ResetUserPassword(ctx context.Context, userID, adminID, adminEmail, ipAddress string,
	req *ResetPasswordRequest) (*OperationResponse, error) {

	// 获取目标用户信息
	targetUser, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target user: %w", err)
	}

	// 加密新密码
	hashedPassword, err := s.encryptor.HashPassword(req.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新密码
	err = s.userRepo.UpdatePassword(ctx, userID, hashedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to update password: %w", err)
	}

	// 强制登出所有会话
	err = s.repo.DeactivateUserSessions(ctx, userID, nil)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID).Warn("Failed to deactivate user sessions")
	}

	// 记录管理操作日志
	logEntry := &UserManagementLog{
		AdminID:      adminID,
		AdminEmail:   adminEmail,
		TargetUserID: userID,
		TargetEmail:  targetUser.Email,
		Action:       ActionResetPassword,
		IPAddress:    ipAddress,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateManagementLog(ctx, logEntry); err != nil {
		s.logger.WithError(err).Error("Failed to create management log")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"admin_id": adminID,
	}).Info("User password reset by admin")

	// TODO: 如果需要发送通知邮件，在这里实现
	if req.SendNotification {
		s.logger.WithField("user_id", userID).Info("Password reset notification email should be sent")
	}

	return &OperationResponse{
		Success:   true,
		Message:   "User password reset successfully",
		Timestamp: time.Now(),
	}, nil
}

// ForceLogoutUser 强制用户登出
func (s *Service) ForceLogoutUser(ctx context.Context, userID, adminID, adminEmail, ipAddress string,
	req *ForceLogoutRequest) (*OperationResponse, error) {

	// 获取目标用户信息
	targetUser, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target user: %w", err)
	}

	// 强制登出会话
	var sessionID *string
	if !req.LogoutAll {
		// 如果不是登出所有会话，这里需要指定具体的会话ID
		// 当前实现为登出所有会话
	}

	err = s.repo.DeactivateUserSessions(ctx, userID, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to force logout user: %w", err)
	}

	// 记录管理操作日志
	logEntry := &UserManagementLog{
		AdminID:      adminID,
		AdminEmail:   adminEmail,
		TargetUserID: userID,
		TargetEmail:  targetUser.Email,
		Action:       ActionForceLogout,
		Reason:       &req.Reason,
		IPAddress:    ipAddress,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateManagementLog(ctx, logEntry); err != nil {
		s.logger.WithError(err).Error("Failed to create management log")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"admin_id":   adminID,
		"reason":     req.Reason,
		"logout_all": req.LogoutAll,
	}).Info("User forced logout by admin")

	return &OperationResponse{
		Success:   true,
		Message:   "User logged out successfully",
		Timestamp: time.Now(),
	}, nil
}

// VerifyUserEmail 管理员验证用户邮箱
func (s *Service) VerifyUserEmail(ctx context.Context, userID, adminID, adminEmail, ipAddress string) (*OperationResponse, error) {
	// 获取目标用户信息
	targetUser, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target user: %w", err)
	}

	// 如果已经验证，直接返回
	if targetUser.EmailVerified {
		return &OperationResponse{
			Success:   true,
			Message:   "User email is already verified",
			Timestamp: time.Now(),
		}, nil
	}

	// 更新邮箱验证状态
	err = s.userRepo.UpdateEmailVerified(ctx, userID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to verify user email: %w", err)
	}

	// 记录管理操作日志
	logEntry := &UserManagementLog{
		AdminID:      adminID,
		AdminEmail:   adminEmail,
		TargetUserID: userID,
		TargetEmail:  targetUser.Email,
		Action:       ActionVerifyEmail,
		IPAddress:    ipAddress,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateManagementLog(ctx, logEntry); err != nil {
		s.logger.WithError(err).Error("Failed to create management log")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":  userID,
		"admin_id": adminID,
	}).Info("User email verified by admin")

	return &OperationResponse{
		Success:   true,
		Message:   "User email verified successfully",
		Timestamp: time.Now(),
	}, nil
}

// ValidateUserExists 验证用户是否存在
func (s *Service) ValidateUserExists(ctx context.Context, userID string) error {
	_, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return auth.ErrUserNotFound
	}
	return nil
}
