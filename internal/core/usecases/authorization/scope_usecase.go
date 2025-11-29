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

// ScopeUseCase defines the business logic for scope management
type ScopeUseCase interface {
	CreateScope(ctx context.Context, req CreateScopeRequest) (*domain.Scope, error)
	GetScope(ctx context.Context, id uint) (*domain.Scope, error)
	GetScopeByCode(ctx context.Context, code string) (*domain.Scope, error)
	ListScopes(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Scope], error)
	UpdateScope(ctx context.Context, id uint, req UpdateScopeRequest) (*domain.Scope, error)
	DeleteScope(ctx context.Context, id uint) error
	ListAllScopes(ctx context.Context) ([]domain.Scope, error)
}

type scopeUseCase struct {
	scopeRepo repositories.ScopeRepository
}

// NewScopeUseCase creates a new scope use case
func NewScopeUseCase(scopeRepo repositories.ScopeRepository) ScopeUseCase {
	return &scopeUseCase{
		scopeRepo: scopeRepo,
	}
}

// CreateScopeRequest represents a request to create a scope
type CreateScopeRequest struct {
	Code        string            `json:"code" binding:"required"`
	Name        string            `json:"name" binding:"required"`
	DisplayName string            `json:"display_name"`
	Description string            `json:"description"`
	Level       domain.ScopeLevel `json:"level" binding:"required"`
	Priority    int               `json:"priority" binding:"required"`
}

// UpdateScopeRequest represents a request to update a scope
type UpdateScopeRequest struct {
	Name        string            `json:"name"`
	DisplayName string            `json:"display_name"`
	Description string            `json:"description"`
	Level       domain.ScopeLevel `json:"level"`
	Priority    *int              `json:"priority"`
}

// CreateScope creates a new scope
func (uc *scopeUseCase) CreateScope(ctx context.Context, req CreateScopeRequest) (*domain.Scope, error) {
	// Check if scope with same code already exists
	existing, err := uc.scopeRepo.GetByCode(ctx, req.Code)
	if err == nil && existing != nil {
		return nil, errors.New(errors.ErrCodeDuplicateEntry, "scope with this code already exists", 409)
	}

	// Validate scope level
	validLevels := map[domain.ScopeLevel]bool{
		domain.ScopeLevelOrganization: true,
		domain.ScopeLevelDepartment:   true,
		domain.ScopeLevelTeam:         true,
		domain.ScopeLevelPersonal:     true,
	}
	if !validLevels[req.Level] {
		return nil, errors.New(errors.ErrCodeValidation, "invalid scope level", 400)
	}

	scope := &domain.Scope{
		Code:        req.Code,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Level:       req.Level,
		Priority:    req.Priority,
		IsSystem:    false,
	}

	if err := uc.scopeRepo.Create(ctx, scope); err != nil {
		logger.Error("Failed to create scope", zap.Error(err))
		return nil, err
	}

	logger.Info("Scope created successfully", zap.String("code", scope.Code), zap.Uint("id", scope.ID))
	return scope, nil
}

// GetScope retrieves a scope by ID
func (uc *scopeUseCase) GetScope(ctx context.Context, id uint) (*domain.Scope, error) {
	scope, err := uc.scopeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return scope, nil
}

// GetScopeByCode retrieves a scope by code
func (uc *scopeUseCase) GetScopeByCode(ctx context.Context, code string) (*domain.Scope, error) {
	scope, err := uc.scopeRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return scope, nil
}

// ListScopes lists all scopes with pagination
func (uc *scopeUseCase) ListScopes(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Scope], error) {
	result, err := uc.scopeRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// UpdateScope updates a scope
func (uc *scopeUseCase) UpdateScope(ctx context.Context, id uint, req UpdateScopeRequest) (*domain.Scope, error) {
	// Get existing scope
	scope, err := uc.scopeRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if it's a system scope
	if scope.IsSystem {
		return nil, errors.New(errors.ErrCodeForbidden, "cannot modify system scope", 403)
	}

	// Validate scope level if provided
	if req.Level != "" {
		validLevels := map[domain.ScopeLevel]bool{
			domain.ScopeLevelOrganization: true,
			domain.ScopeLevelDepartment:   true,
			domain.ScopeLevelTeam:         true,
			domain.ScopeLevelPersonal:     true,
		}
		if !validLevels[req.Level] {
			return nil, errors.New(errors.ErrCodeValidation, "invalid scope level", 400)
		}
	}

	// Update fields
	if req.Name != "" {
		scope.Name = req.Name
	}
	if req.DisplayName != "" {
		scope.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		scope.Description = req.Description
	}
	if req.Level != "" {
		scope.Level = req.Level
	}
	if req.Priority != nil {
		scope.Priority = *req.Priority
	}

	if err := uc.scopeRepo.Update(ctx, scope); err != nil {
		logger.Error("Failed to update scope", zap.Error(err))
		return nil, err
	}

	logger.Info("Scope updated successfully", zap.Uint("id", id))
	return scope, nil
}

// DeleteScope deletes a scope
func (uc *scopeUseCase) DeleteScope(ctx context.Context, id uint) error {
	// Get scope
	scope, err := uc.scopeRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if it's a system scope
	if scope.IsSystem {
		return errors.New(errors.ErrCodeForbidden, "cannot delete system scope", 403)
	}

	// TODO: Check if scope has dependencies (permissions using this scope)

	if err := uc.scopeRepo.Delete(ctx, id); err != nil {
		logger.Error("Failed to delete scope", zap.Error(err))
		return err
	}

	logger.Info("Scope deleted successfully", zap.Uint("id", id))
	return nil
}

// ListAllScopes lists all scopes without pagination
func (uc *scopeUseCase) ListAllScopes(ctx context.Context) ([]domain.Scope, error) {
	scopes, err := uc.scopeRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	return scopes, nil
}
