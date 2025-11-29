package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/config"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/internal/infrastructure/cache"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"github.com/owner/go-cms/pkg/utils"
	"github.com/pquerna/otp/totp"
	"go.uber.org/zap"
)

// UseCase defines the authentication use case interface
type UseCase interface {
	// Registration
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)
	VerifyEmail(ctx context.Context, email, code string) error
	ResendOTP(ctx context.Context, email string) error

	// Login
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)
	Logout(ctx context.Context, userID uuid.UUID, token string) error
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error)

	// Password Management
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, email, code, newPassword string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error

	// 2FA
	Enable2FA(ctx context.Context, userID uuid.UUID) (*Enable2FAResponse, error)
	Verify2FA(ctx context.Context, userID uuid.UUID, code string) error
	Disable2FA(ctx context.Context, userID uuid.UUID, password string) error

	// User Info
	GetCurrentUser(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*domain.User, error)
}

// useCase implements the UseCase interface
type useCase struct {
	userRepo         repositories.UserRepository
	otpRepo          repositories.OTPRepository
	refreshTokenRepo repositories.RefreshTokenRepository
	config           *config.Config
	emailService     EmailService
}

// EmailService defines the interface for email operations
type EmailService interface {
	SendVerifyEmailOTP(to, name, otp string, expirySeconds int) error
	SendResetPasswordOTP(to, name, otp string, expirySeconds int) error
	SendWelcomeEmail(to, name string) error
}

// NewUseCase creates a new authentication use case
func NewUseCase(
	userRepo repositories.UserRepository,
	otpRepo repositories.OTPRepository,
	refreshTokenRepo repositories.RefreshTokenRepository,
	config *config.Config,
	emailService EmailService,
) UseCase {
	return &useCase{
		userRepo:         userRepo,
		otpRepo:          otpRepo,
		refreshTokenRepo: refreshTokenRepo,
		config:           config,
		emailService:     emailService,
	}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
	Code2FA  string `json:"code_2fa"` // Optional 2FA code
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"`
	User         *domain.User `json:"user"`
	Requires2FA  bool         `json:"requires_2fa,omitempty"`
}

// Enable2FAResponse represents a 2FA enable response
type Enable2FAResponse struct {
	Secret  string `json:"secret"`
	QRCode  string `json:"qr_code"`
	Message string `json:"message"`
}

// UpdateProfileRequest represents a profile update request
type UpdateProfileRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Avatar    string `json:"avatar"`
}

// Register registers a new user
func (uc *useCase) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Check if user already exists
	existingUser, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.ErrDuplicateEntry
	}

	// Validate password strength
	if err := utils.ValidatePassword(req.Password); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeValidation, err.Error(), 400)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to hash password", 500)
	}

	// Create user
	user := &domain.User{
		Email:     req.Email,
		Password:  hashedPassword,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Status:    domain.UserStatusPending,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	// Generate OTP for email verification
	otpCode, err := utils.GenerateOTP(uc.config.OTP.Length)
	if err != nil {
		logger.Error("Failed to generate OTP", zap.Error(err))
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to generate OTP", 500)
	}

	// Save OTP
	otp := &domain.OTP{
		Email:     req.Email,
		Code:      otpCode,
		Type:      "email_verification",
		ExpiresAt: time.Now().Add(uc.config.OTP.Expire),
	}

	if err := uc.otpRepo.Create(ctx, otp); err != nil {
		logger.Error("Failed to save OTP", zap.Error(err))
	}

	// Cache OTP in Redis
	_ = cache.CacheOTP(ctx, req.Email, otpCode, uc.config.OTP.Expire)

	// Send verification email with OTP
	name := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	if err := uc.emailService.SendVerifyEmailOTP(req.Email, name, otpCode, int(uc.config.OTP.Expire.Seconds())); err != nil {
		logger.Error("Failed to send verification email", zap.Error(err))
	}

	logger.Info("User registered successfully", zap.String("email", req.Email), zap.String("otp", otpCode))

	// Generate tokens
	accessToken, err := utils.GenerateJWT(user.ID, user.Email, &uc.config.JWT)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to generate access token", 500)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email, &uc.config.JWT)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to generate refresh token", 500)
	}

	// Save refresh token
	refreshTokenModel := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(uc.config.JWT.RefreshTokenExpire),
	}

	if err := uc.refreshTokenRepo.Create(ctx, refreshTokenModel); err != nil {
		logger.Error("Failed to save refresh token", zap.Error(err))
	}

	// Hide sensitive data
	user.Password = ""
	user.TwoFactorSecret = ""

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(uc.config.JWT.AccessTokenExpire.Seconds()),
		User:         user,
	}, nil
}

// VerifyEmail verifies a user's email with OTP
func (uc *useCase) VerifyEmail(ctx context.Context, email, code string) error {
	// Check OTP from cache first
	cachedOTP, err := cache.GetOTP(ctx, email)
	if err == nil && cachedOTP == code {
		// OTP is valid, verify email
		user, err := uc.userRepo.GetByEmail(ctx, email)
		if err != nil {
			return err
		}

		if err := uc.userRepo.VerifyEmail(ctx, user.ID); err != nil {
			return err
		}

		// Update status to active
		if err := uc.userRepo.UpdateStatus(ctx, user.ID, domain.UserStatusActive); err != nil {
			return err
		}

		// Delete OTP from cache
		_ = cache.DeleteOTP(ctx, email)

		return nil
	}

	// Check OTP from database
	otp, err := uc.otpRepo.GetByEmail(ctx, email, "email_verification")
	if err != nil {
		return errors.ErrInvalidOTP
	}

	if otp.Used {
		return errors.ErrInvalidOTP
	}

	if otp.IsExpired() {
		return errors.ErrExpiredOTP
	}

	if otp.Code != code {
		return errors.ErrInvalidOTP
	}

	// Mark OTP as used
	if err := uc.otpRepo.MarkAsUsed(ctx, otp.ID); err != nil {
		logger.Error("Failed to mark OTP as used", zap.Error(err))
	}

	// Verify email
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	if err := uc.userRepo.VerifyEmail(ctx, user.ID); err != nil {
		return err
	}

	// Update status to active
	if err := uc.userRepo.UpdateStatus(ctx, user.ID, domain.UserStatusActive); err != nil {
		return err
	}

	return nil
}

// ResendOTP resends OTP to user's email
func (uc *useCase) ResendOTP(ctx context.Context, email string) error {
	// Check if user exists
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return errors.ErrUserNotFound
	}

	if user.EmailVerified {
		return errors.New("EMAIL_ALREADY_VERIFIED", "Email already verified", 400)
	}

	// Delete old OTP
	_ = uc.otpRepo.DeleteByEmail(ctx, email, "email_verification")
	_ = cache.DeleteOTP(ctx, email)

	// Generate new OTP
	otpCode, err := utils.GenerateOTP(uc.config.OTP.Length)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to generate OTP", 500)
	}

	// Save OTP
	otp := &domain.OTP{
		Email:     email,
		Code:      otpCode,
		Type:      "email_verification",
		ExpiresAt: time.Now().Add(uc.config.OTP.Expire),
	}

	if err := uc.otpRepo.Create(ctx, otp); err != nil {
		return err
	}

	// Cache OTP
	_ = cache.CacheOTP(ctx, email, otpCode, uc.config.OTP.Expire)

	// Send email with OTP
	name := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	if err := uc.emailService.SendVerifyEmailOTP(email, name, otpCode, int(uc.config.OTP.Expire.Seconds())); err != nil {
		logger.Error("Failed to send OTP email", zap.Error(err))
	}

	logger.Info("OTP resent", zap.String("email", email), zap.String("otp", otpCode))

	return nil
}

// Login authenticates a user
func (uc *useCase) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	// Get user by email
	user, err := uc.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	// Check password
	if !utils.CheckPassword(user.Password, req.Password) {
		return nil, errors.ErrInvalidCredentials
	}

	// Check user status
	if user.Status != domain.UserStatusActive {
		return nil, errors.New("USER_INACTIVE", fmt.Sprintf("User account is %s", user.Status), 403)
	}

	// Check 2FA
	if user.TwoFactorEnabled {
		if req.Code2FA == "" {
			return &AuthResponse{
				Requires2FA: true,
			}, nil
		}

		// Verify 2FA code
		valid := totp.Validate(req.Code2FA, user.TwoFactorSecret)
		if !valid {
			return nil, errors.ErrInvalid2FA
		}
	}

	// Generate tokens
	accessToken, err := utils.GenerateJWT(user.ID, user.Email, &uc.config.JWT)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to generate access token", 500)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.Email, &uc.config.JWT)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to generate refresh token", 500)
	}

	// Save refresh token
	refreshTokenModel := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(uc.config.JWT.RefreshTokenExpire),
	}

	if err := uc.refreshTokenRepo.Create(ctx, refreshTokenModel); err != nil {
		logger.Error("Failed to save refresh token", zap.Error(err))
	}

	// Update last login
	// TODO: Get IP from context
	_ = uc.userRepo.UpdateLastLogin(ctx, user.ID, "")

	// Hide sensitive data
	user.Password = ""
	user.TwoFactorSecret = ""

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(uc.config.JWT.AccessTokenExpire.Seconds()),
		User:         user,
	}, nil
}

// Logout logs out a user
func (uc *useCase) Logout(ctx context.Context, userID uuid.UUID, token string) error {
	// Revoke refresh token
	if err := uc.refreshTokenRepo.Revoke(ctx, token); err != nil {
		logger.Error("Failed to revoke refresh token", zap.Error(err))
	}

	// TODO: Blacklist access token in Redis

	return nil
}

// RefreshToken refreshes an access token
func (uc *useCase) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Get refresh token from database
	tokenModel, err := uc.refreshTokenRepo.GetByToken(ctx, refreshToken)
	if err != nil {
		return nil, errors.ErrInvalidToken
	}

	if tokenModel.Revoked {
		return nil, errors.ErrInvalidToken
	}

	if tokenModel.IsExpired() {
		return nil, errors.ErrExpiredToken
	}

	// Get user
	user, err := uc.userRepo.GetByID(ctx, tokenModel.UserID)
	if err != nil {
		return nil, err
	}

	// Generate new access token
	accessToken, err := utils.GenerateJWT(user.ID, user.Email, &uc.config.JWT)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to generate access token", 500)
	}

	// Hide sensitive data
	user.Password = ""
	user.TwoFactorSecret = ""

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(uc.config.JWT.AccessTokenExpire.Seconds()),
		User:         user,
	}, nil
}

// ForgotPassword initiates password reset
func (uc *useCase) ForgotPassword(ctx context.Context, email string) error {
	// Check if user exists
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not
		return nil
	}

	// Delete old OTP
	_ = uc.otpRepo.DeleteByEmail(ctx, email, "password_reset")

	// Generate OTP
	otpCode, err := utils.GenerateOTP(uc.config.OTP.Length)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to generate OTP", 500)
	}

	// Save OTP
	otp := &domain.OTP{
		Email:     email,
		Code:      otpCode,
		Type:      "password_reset",
		ExpiresAt: time.Now().Add(uc.config.OTP.Expire),
	}

	if err := uc.otpRepo.Create(ctx, otp); err != nil {
		return err
	}

	// Cache OTP
	_ = cache.CacheOTP(ctx, email, otpCode, uc.config.OTP.Expire)

	// Send password reset email
	name := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	if err := uc.emailService.SendResetPasswordOTP(email, name, otpCode, int(uc.config.OTP.Expire.Seconds())); err != nil {
		logger.Error("Failed to send password reset email", zap.Error(err))
	}

	logger.Info("Password reset OTP sent", zap.String("email", email), zap.String("otp", otpCode), zap.String("user_id", user.ID.String()))

	return nil
}

// ResetPassword resets user password
func (uc *useCase) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	// Validate password
	if err := utils.ValidatePassword(newPassword); err != nil {
		return errors.Wrap(err, errors.ErrCodeValidation, err.Error(), 400)
	}

	// Verify OTP
	otp, err := uc.otpRepo.GetByEmail(ctx, email, "password_reset")
	if err != nil {
		return errors.ErrInvalidOTP
	}

	if otp.Used || otp.IsExpired() || otp.Code != code {
		return errors.ErrInvalidOTP
	}

	// Get user
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to hash password", 500)
	}

	// Update password
	if err := uc.userRepo.UpdatePassword(ctx, user.ID, hashedPassword); err != nil {
		return err
	}

	// Mark OTP as used
	_ = uc.otpRepo.MarkAsUsed(ctx, otp.ID)

	// Revoke all refresh tokens
	_ = uc.refreshTokenRepo.RevokeAllByUserID(ctx, user.ID)

	return nil
}

// ChangePassword changes user password
func (uc *useCase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	// Get user
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify old password
	if !utils.CheckPassword(user.Password, oldPassword) {
		return errors.New("INVALID_PASSWORD", "Invalid current password", 400)
	}

	// Validate new password
	if err := utils.ValidatePassword(newPassword); err != nil {
		return errors.Wrap(err, errors.ErrCodeValidation, err.Error(), 400)
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to hash password", 500)
	}

	// Update password
	if err := uc.userRepo.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return err
	}

	// Revoke all refresh tokens
	_ = uc.refreshTokenRepo.RevokeAllByUserID(ctx, userID)

	return nil
}

// Enable2FA enables 2FA for a user
func (uc *useCase) Enable2FA(ctx context.Context, userID uuid.UUID) (*Enable2FAResponse, error) {
	// Get user
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.TwoFactorEnabled {
		return nil, errors.New("2FA_ALREADY_ENABLED", "2FA is already enabled", 400)
	}

	// Generate TOTP secret
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      uc.config.TwoFA.Issuer,
		AccountName: user.Email,
	})
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to generate 2FA secret", 500)
	}

	// Save secret (but don't enable yet - user needs to verify first)
	// We'll store it temporarily in cache
	_ = cache.Set(ctx, fmt.Sprintf("2fa_setup:%s", userID), key.Secret(), 10*time.Minute)

	return &Enable2FAResponse{
		Secret:  key.Secret(),
		QRCode:  key.URL(),
		Message: "Scan the QR code with your authenticator app and verify with a code",
	}, nil
}

// Verify2FA verifies and enables 2FA
func (uc *useCase) Verify2FA(ctx context.Context, userID uuid.UUID, code string) error {
	// Get temporary secret from cache
	secret, err := cache.Get(ctx, fmt.Sprintf("2fa_setup:%s", userID))
	if err != nil {
		return errors.New("2FA_SETUP_NOT_FOUND", "2FA setup not found or expired", 400)
	}

	// Verify code
	valid := totp.Validate(code, secret)
	if !valid {
		return errors.ErrInvalid2FA
	}

	// Enable 2FA
	if err := uc.userRepo.Enable2FA(ctx, userID, secret); err != nil {
		return err
	}

	// Delete temporary secret
	_ = cache.Delete(ctx, fmt.Sprintf("2fa_setup:%s", userID))

	return nil
}

// Disable2FA disables 2FA for a user
func (uc *useCase) Disable2FA(ctx context.Context, userID uuid.UUID, password string) error {
	// Get user
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if !user.TwoFactorEnabled {
		return errors.New("2FA_NOT_ENABLED", "2FA is not enabled", 400)
	}

	// Verify password
	if !utils.CheckPassword(user.Password, password) {
		return errors.New("INVALID_PASSWORD", "Invalid password", 400)
	}

	// Disable 2FA
	if err := uc.userRepo.Disable2FA(ctx, userID); err != nil {
		return err
	}

	return nil
}

// GetCurrentUser retrieves the current user
func (uc *useCase) GetCurrentUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Hide sensitive data
	user.Password = ""
	user.TwoFactorSecret = ""

	return user, nil
}

// UpdateProfile updates user profile
func (uc *useCase) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*domain.User, error) {
	// Get user
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	// Save
	if err := uc.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Hide sensitive data
	user.Password = ""
	user.TwoFactorSecret = ""

	return user, nil
}
