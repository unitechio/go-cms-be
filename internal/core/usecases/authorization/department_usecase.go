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

// DepartmentUseCase defines the business logic for department management
type DepartmentUseCase interface {
	CreateDepartment(ctx context.Context, req CreateDepartmentRequest) (*domain.Department, error)
	GetDepartment(ctx context.Context, id uint) (*domain.Department, error)
	GetDepartmentByCode(ctx context.Context, code string) (*domain.Department, error)
	ListDepartments(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Department], error)
	ListDepartmentsByModule(ctx context.Context, moduleID uint) ([]domain.Department, error)
	UpdateDepartment(ctx context.Context, id uint, req UpdateDepartmentRequest) (*domain.Department, error)
	DeleteDepartment(ctx context.Context, id uint) error
	ListActiveDepartments(ctx context.Context) ([]domain.Department, error)
}

type departmentUseCase struct {
	departmentRepo repositories.DepartmentRepository
	moduleRepo     repositories.ModuleRepository
}

// NewDepartmentUseCase creates a new department use case
func NewDepartmentUseCase(
	departmentRepo repositories.DepartmentRepository,
	moduleRepo repositories.ModuleRepository,
) DepartmentUseCase {
	return &departmentUseCase{
		departmentRepo: departmentRepo,
		moduleRepo:     moduleRepo,
	}
}

// CreateDepartmentRequest represents a request to create a department
type CreateDepartmentRequest struct {
	ModuleID    uint   `json:"module_id" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	ParentID    *uint  `json:"parent_id"`
	ManagerID   *uint  `json:"manager_id"`
	IsActive    bool   `json:"is_active"`
}

// UpdateDepartmentRequest represents a request to update a department
type UpdateDepartmentRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	ParentID    *uint  `json:"parent_id"`
	ManagerID   *uint  `json:"manager_id"`
	IsActive    *bool  `json:"is_active"`
}

// CreateDepartment creates a new department
func (uc *departmentUseCase) CreateDepartment(ctx context.Context, req CreateDepartmentRequest) (*domain.Department, error) {
	// Validate module exists
	_, err := uc.moduleRepo.GetByID(ctx, req.ModuleID)
	if err != nil {
		return nil, errors.New(errors.ErrCodeNotFound, "module not found", 404)
	}

	// Check if department with same code already exists
	existing, err := uc.departmentRepo.GetByCode(ctx, req.Code)
	if err == nil && existing != nil {
		return nil, errors.New(errors.ErrCodeDuplicateEntry, "department with this code already exists", 409)
	}

	// Validate parent department if provided
	if req.ParentID != nil {
		parent, err := uc.departmentRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, errors.New(errors.ErrCodeNotFound, "parent department not found", 404)
		}
		// Ensure parent is in the same module
		if parent.ModuleID != req.ModuleID {
			return nil, errors.New(errors.ErrCodeValidation, "parent department must be in the same module", 400)
		}
	}

	department := &domain.Department{
		ModuleID:    req.ModuleID,
		Code:        req.Code,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		ParentID:    req.ParentID,
		ManagerID:   req.ManagerID,
		IsActive:    req.IsActive,
		IsSystem:    false,
	}

	if err := uc.departmentRepo.Create(ctx, department); err != nil {
		logger.Error("Failed to create department", zap.Error(err))
		return nil, err
	}

	logger.Info("Department created successfully", zap.String("code", department.Code), zap.Uint("id", department.ID))
	return department, nil
}

// GetDepartment retrieves a department by ID
func (uc *departmentUseCase) GetDepartment(ctx context.Context, id uint) (*domain.Department, error) {
	department, err := uc.departmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return department, nil
}

// GetDepartmentByCode retrieves a department by code
func (uc *departmentUseCase) GetDepartmentByCode(ctx context.Context, code string) (*domain.Department, error) {
	department, err := uc.departmentRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return department, nil
}

// ListDepartments lists all departments with pagination
func (uc *departmentUseCase) ListDepartments(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Department], error) {
	result, err := uc.departmentRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ListDepartmentsByModule lists departments by module
func (uc *departmentUseCase) ListDepartmentsByModule(ctx context.Context, moduleID uint) ([]domain.Department, error) {
	// Validate module exists
	_, err := uc.moduleRepo.GetByID(ctx, moduleID)
	if err != nil {
		return nil, errors.New(errors.ErrCodeNotFound, "module not found", 404)
	}

	departments, err := uc.departmentRepo.ListByModule(ctx, moduleID)
	if err != nil {
		return nil, err
	}
	return departments, nil
}

// UpdateDepartment updates a department
func (uc *departmentUseCase) UpdateDepartment(ctx context.Context, id uint, req UpdateDepartmentRequest) (*domain.Department, error) {
	// Get existing department
	department, err := uc.departmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if it's a system department
	if department.IsSystem {
		return nil, errors.New(errors.ErrCodeForbidden, "cannot modify system department", 403)
	}

	// Validate parent department if provided
	if req.ParentID != nil {
		// Prevent self-reference
		if *req.ParentID == id {
			return nil, errors.New(errors.ErrCodeValidation, "department cannot be its own parent", 400)
		}

		parent, err := uc.departmentRepo.GetByID(ctx, *req.ParentID)
		if err != nil {
			return nil, errors.New(errors.ErrCodeNotFound, "parent department not found", 404)
		}
		// Ensure parent is in the same module
		if parent.ModuleID != department.ModuleID {
			return nil, errors.New(errors.ErrCodeValidation, "parent department must be in the same module", 400)
		}
	}

	// Update fields
	if req.Name != "" {
		department.Name = req.Name
	}
	if req.DisplayName != "" {
		department.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		department.Description = req.Description
	}
	if req.ParentID != nil {
		department.ParentID = req.ParentID
	}
	if req.ManagerID != nil {
		department.ManagerID = req.ManagerID
	}
	if req.IsActive != nil {
		department.IsActive = *req.IsActive
	}

	if err := uc.departmentRepo.Update(ctx, department); err != nil {
		logger.Error("Failed to update department", zap.Error(err))
		return nil, err
	}

	logger.Info("Department updated successfully", zap.Uint("id", id))
	return department, nil
}

// DeleteDepartment deletes a department
func (uc *departmentUseCase) DeleteDepartment(ctx context.Context, id uint) error {
	// Get department
	department, err := uc.departmentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if it's a system department
	if department.IsSystem {
		return errors.New(errors.ErrCodeForbidden, "cannot delete system department", 403)
	}

	// TODO: Check if department has children or dependencies

	if err := uc.departmentRepo.Delete(ctx, id); err != nil {
		logger.Error("Failed to delete department", zap.Error(err))
		return err
	}

	logger.Info("Department deleted successfully", zap.Uint("id", id))
	return nil
}

// ListActiveDepartments lists all active departments
func (uc *departmentUseCase) ListActiveDepartments(ctx context.Context) ([]domain.Department, error) {
	departments, err := uc.departmentRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	return departments, nil
}
