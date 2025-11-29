package authorization

import (
	"context"

	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/errors"
	"github.com/owner/go-cms/pkg/logger"
	"github.com/owner/go-cms/pkg/pagination"
	"go.uber.org/zap"
)

// ModuleUseCase defines the business logic for module management
type ModuleUseCase interface {
	CreateModule(ctx context.Context, req CreateModuleRequest) (*domain.Module, error)
	GetModule(ctx context.Context, id uint) (*domain.Module, error)
	GetModuleByCode(ctx context.Context, code string) (*domain.Module, error)
	ListModules(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Module], error)
	UpdateModule(ctx context.Context, id uint, req UpdateModuleRequest) (*domain.Module, error)
	DeleteModule(ctx context.Context, id uint) error
	ListActiveModules(ctx context.Context) ([]domain.Module, error)
}

type moduleUseCase struct {
	moduleRepo repositories.ModuleRepository
}

// NewModuleUseCase creates a new module use case
func NewModuleUseCase(moduleRepo repositories.ModuleRepository) ModuleUseCase {
	return &moduleUseCase{
		moduleRepo: moduleRepo,
	}
}

// CreateModuleRequest represents a request to create a module
type CreateModuleRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	Order       int    `json:"order"`
	IsActive    bool   `json:"is_active"`
}

// UpdateModuleRequest represents a request to update a module
type UpdateModuleRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	Order       *int   `json:"order"`
	IsActive    *bool  `json:"is_active"`
}

// CreateModule creates a new module
func (uc *moduleUseCase) CreateModule(ctx context.Context, req CreateModuleRequest) (*domain.Module, error) {
	// Check if module with same code already exists
	existing, err := uc.moduleRepo.GetByCode(ctx, req.Code)
	if err == nil && existing != nil {
		return nil, errors.New(errors.ErrCodeDuplicateEntry, "module with this code already exists", 409)
	}

	module := &domain.Module{
		Code:        req.Code,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Icon:        req.Icon,
		Color:       req.Color,
		Order:       req.Order,
		IsActive:    req.IsActive,
		IsSystem:    false, // User-created modules are not system modules
	}

	if err := uc.moduleRepo.Create(ctx, module); err != nil {
		logger.Error("Failed to create module", zap.Error(err))
		return nil, err
	}

	logger.Info("Module created successfully", zap.String("code", module.Code), zap.Uint("id", module.ID))
	return module, nil
}

// GetModule retrieves a module by ID
func (uc *moduleUseCase) GetModule(ctx context.Context, id uint) (*domain.Module, error) {
	module, err := uc.moduleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return module, nil
}

// GetModuleByCode retrieves a module by code
func (uc *moduleUseCase) GetModuleByCode(ctx context.Context, code string) (*domain.Module, error) {
	module, err := uc.moduleRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return module, nil
}

// ListModules lists all modules with pagination
func (uc *moduleUseCase) ListModules(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Module], error) {
	result, err := uc.moduleRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateModule updates a module
func (uc *moduleUseCase) UpdateModule(ctx context.Context, id uint, req UpdateModuleRequest) (*domain.Module, error) {
	// Get existing module
	module, err := uc.moduleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if it's a system module
	if module.IsSystem {
		return nil, errors.New(errors.ErrCodeForbidden, "cannot modify system module", 403)
	}

	// Update fields
	if req.Name != "" {
		module.Name = req.Name
	}
	if req.DisplayName != "" {
		module.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		module.Description = req.Description
	}
	if req.Icon != "" {
		module.Icon = req.Icon
	}
	if req.Color != "" {
		module.Color = req.Color
	}
	if req.Order != nil {
		module.Order = *req.Order
	}
	if req.IsActive != nil {
		module.IsActive = *req.IsActive
	}

	if err := uc.moduleRepo.Update(ctx, module); err != nil {
		logger.Error("Failed to update module", zap.Error(err))
		return nil, err
	}

	logger.Info("Module updated successfully", zap.Uint("id", id))
	return module, nil
}

// DeleteModule deletes a module
func (uc *moduleUseCase) DeleteModule(ctx context.Context, id uint) error {
	// Get module
	module, err := uc.moduleRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if it's a system module
	if module.IsSystem {
		return errors.New(errors.ErrCodeForbidden, "cannot delete system module", 403)
	}

	// TODO: Check if module has dependencies (departments, permissions, etc.)
	// For now, we'll allow deletion

	if err := uc.moduleRepo.Delete(ctx, id); err != nil {
		logger.Error("Failed to delete module", zap.Error(err))
		return err
	}

	logger.Info("Module deleted successfully", zap.Uint("id", id))
	return nil
}

// ListActiveModules lists all active modules
func (uc *moduleUseCase) ListActiveModules(ctx context.Context) ([]domain.Module, error) {
	modules, err := uc.moduleRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	return modules, nil
}
