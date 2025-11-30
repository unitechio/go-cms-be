package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"github.com/owner/go-cms/pkg/pagination"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// UserUseCase handles user business logic
type UserUseCase struct {
	userRepo       repositories.UserRepository
	roleRepo       repositories.RoleRepository
	departmentRepo repositories.DepartmentRepository
}

// NewUserUseCase creates a new user use case
func NewUserUseCase(
	userRepo repositories.UserRepository,
	roleRepo repositories.RoleRepository,
	departmentRepo repositories.DepartmentRepository,
) *UserUseCase {
	return &UserUseCase{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		departmentRepo: departmentRepo,
	}
}

// GetDepartmentByCode retrieves a department by its code
func (uc *UserUseCase) GetDepartmentByCode(ctx context.Context, code string) (*domain.Department, error) {
	return uc.departmentRepo.GetByCode(ctx, code)
}

// CreateUser creates a new user
func (uc *UserUseCase) CreateUser(ctx context.Context, user *domain.User) error {
	// Validate user
	if user.Email == "" {
		return errors.New(errors.ErrCodeValidation, "email is required", 400)
	}
	if user.Password == "" {
		return errors.New(errors.ErrCodeValidation, "password is required", 400)
	}

	// Check if user with same email already exists
	existing, err := uc.userRepo.GetByEmail(ctx, user.Email)
	if err == nil && existing != nil {
		return errors.New(errors.ErrCodeConflict, "user with this email already exists", 409)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.Wrap(err, errors.ErrCodeInternal, "failed to hash password", 500)
	}
	user.Password = string(hashedPassword)

	// Set default status if not set
	if user.Status == "" {
		user.Status = domain.UserStatusActive
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		logger.Error("Failed to create user", zap.Error(err))
		return err
	}

	logger.Info("User created successfully", zap.String("email", user.Email), zap.String("id", user.ID.String()))
	return nil
}

// GetUser retrieves a user by ID
func (uc *UserUseCase) GetUser(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (uc *UserUseCase) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	user, err := uc.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// ListUsers retrieves all users with optional filters
func (uc *UserUseCase) ListUsers(ctx context.Context, filter repositories.UserFilter, page *pagination.OffsetPagination) ([]*domain.User, int64, error) {
	users, total, err := uc.userRepo.List(ctx, filter, page)
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// UpdateUser updates an existing user
func (uc *UserUseCase) UpdateUser(ctx context.Context, id uuid.UUID, updates *domain.User) error {
	// Get existing user
	existing, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Update fields
	if updates.FirstName != "" {
		existing.FirstName = updates.FirstName
	}
	if updates.LastName != "" {
		existing.LastName = updates.LastName
	}
	if updates.Phone != "" {
		existing.Phone = updates.Phone
	}
	if updates.Avatar != "" {
		existing.Avatar = updates.Avatar
	}
	if updates.Status != "" {
		existing.Status = updates.Status
	}
	if updates.DepartmentID != nil {
		existing.DepartmentID = updates.DepartmentID
	}
	if updates.Position != "" {
		existing.Position = updates.Position
	}

	if err := uc.userRepo.Update(ctx, existing); err != nil {
		logger.Error("Failed to update user", zap.Error(err))
		return err
	}

	logger.Info("User updated successfully", zap.String("id", id.String()))
	return nil
}

// DeleteUser deletes a user
func (uc *UserUseCase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	// Get existing user
	_, err := uc.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.userRepo.Delete(ctx, id); err != nil {
		logger.Error("Failed to delete user", zap.Error(err))
		return err
	}

	logger.Info("User deleted successfully", zap.String("id", id.String()))
	return nil
}

// AssignRole assigns a role to a user
func (uc *UserUseCase) AssignRole(ctx context.Context, userID uuid.UUID, roleID uint) error {
	// Verify user exists
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify role exists
	_, err = uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return err
	}

	if err := uc.userRepo.AssignRole(ctx, userID, roleID); err != nil {
		logger.Error("Failed to assign role to user", zap.Error(err))
		return err
	}

	logger.Info("Role assigned to user", zap.String("userID", userID.String()), zap.Uint("roleID", roleID))
	return nil
}

// RemoveRole removes a role from a user
func (uc *UserUseCase) RemoveRole(ctx context.Context, userID uuid.UUID, roleID uint) error {
	if err := uc.userRepo.RemoveRole(ctx, userID, roleID); err != nil {
		logger.Error("Failed to remove role from user", zap.Error(err))
		return err
	}

	logger.Info("Role removed from user", zap.String("userID", userID.String()), zap.Uint("roleID", roleID))
	return nil
}

// GetUserRoles retrieves all roles for a user
func (uc *UserUseCase) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	roles, err := uc.userRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	return roles, nil
}
