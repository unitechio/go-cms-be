package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/owner/go-cms/internal/core/usecases/auth"
	"github.com/owner/go-cms/internal/http/middleware"
	"github.com/owner/go-cms/pkg/response"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authUseCase auth.UseCase
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authUseCase auth.UseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user account with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.RegisterRequest true "Registration request"
// @Success 201 {object} response.Response{data=auth.AuthResponse}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	result, err := h.authUseCase.Register(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, result)
}

// Login godoc
// @Summary Login
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body auth.LoginRequest true "Login request"
// @Success 200 {object} response.Response{data=auth.AuthResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	result, err := h.authUseCase.Login(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// VerifyEmail godoc
// @Summary Verify email
// @Description Verify email address with OTP code
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{email=string,code=string} true "Verification request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.authUseCase.VerifyEmail(c.Request.Context(), req.Email, req.Code); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Email verified successfully",
	})
}

// ResendOTP godoc
// @Summary Resend OTP
// @Description Resend OTP code to email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{email=string} true "Resend OTP request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/resend-otp [post]
func (h *AuthHandler) ResendOTP(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.authUseCase.ResendOTP(c.Request.Context(), req.Email); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "OTP sent successfully",
	})
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{refresh_token=string} true "Refresh token request"
// @Success 200 {object} response.Response{data=auth.AuthResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	result, err := h.authUseCase.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// ForgotPassword godoc
// @Summary Forgot password
// @Description Request password reset OTP
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{email=string} true "Forgot password request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.authUseCase.ForgotPassword(c.Request.Context(), req.Email); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Password reset OTP sent to your email",
	})
}

// ResetPassword godoc
// @Summary Reset password
// @Description Reset password with OTP code
// @Tags auth
// @Accept json
// @Produce json
// @Param request body object{email=string,code=string,new_password=string} true "Reset password request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/reset-password [post]
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Code        string `json:"code" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.authUseCase.ResetPassword(c.Request.Context(), req.Email, req.Code, req.NewPassword); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Password reset successfully",
	})
}

// Logout godoc
// @Summary Logout
// @Description Logout and revoke refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{refresh_token=string} true "Logout request"
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.authUseCase.Logout(c.Request.Context(), userID, req.RefreshToken); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Logged out successfully",
	})
}

// ChangePassword godoc
// @Summary Change password
// @Description Change user password
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{old_password=string,new_password=string} true "Change password request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.authUseCase.ChangePassword(c.Request.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "Password changed successfully",
	})
}

// GetMe godoc
// @Summary Get current user
// @Description Get current authenticated user information
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=domain.User}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	user, err := h.authUseCase.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// UpdateProfile godoc
// @Summary Update profile
// @Description Update current user profile
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body auth.UpdateProfileRequest true "Update profile request"
// @Success 200 {object} response.Response{data=domain.User}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/me [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req auth.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	user, err := h.authUseCase.UpdateProfile(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// Enable2FA godoc
// @Summary Enable 2FA
// @Description Enable two-factor authentication
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=auth.Enable2FAResponse}
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/2fa/enable [post]
func (h *AuthHandler) Enable2FA(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	result, err := h.authUseCase.Enable2FA(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// Verify2FA godoc
// @Summary Verify 2FA
// @Description Verify and activate two-factor authentication
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{code=string} true "2FA verification request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/2fa/verify [post]
func (h *AuthHandler) Verify2FA(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req struct {
		Code string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.authUseCase.Verify2FA(c.Request.Context(), userID, req.Code); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "2FA enabled successfully",
	})
}

// Disable2FA godoc
// @Summary Disable 2FA
// @Description Disable two-factor authentication
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body object{password=string} true "Disable 2FA request"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /auth/2fa/disable [post]
func (h *AuthHandler) Disable2FA(c *gin.Context) {
	userID := middleware.MustGetUserID(c)

	var req struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	if err := h.authUseCase.Disable2FA(c.Request.Context(), userID, req.Password); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"message": "2FA disabled successfully",
	})
}
