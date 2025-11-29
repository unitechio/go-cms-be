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

// ServiceUseCase defines the business logic for service management
type ServiceUseCase interface {
	CreateService(ctx context.Context, req CreateServiceRequest) (*domain.Service, error)
	GetService(ctx context.Context, id uint) (*domain.Service, error)
	GetServiceByCode(ctx context.Context, code string) (*domain.Service, error)
	ListServices(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Service], error)
	ListServicesByDepartment(ctx context.Context, departmentID uint) ([]domain.Service, error)
	UpdateService(ctx context.Context, id uint, req UpdateServiceRequest) (*domain.Service, error)
	DeleteService(ctx context.Context, id uint) error
	ListActiveServices(ctx context.Context) ([]domain.Service, error)
}

type serviceUseCase struct {
	serviceRepo    repositories.ServiceRepository
	departmentRepo repositories.DepartmentRepository
}

// NewServiceUseCase creates a new service use case
func NewServiceUseCase(
	serviceRepo repositories.ServiceRepository,
	departmentRepo repositories.DepartmentRepository,
) ServiceUseCase {
	return &serviceUseCase{
		serviceRepo:    serviceRepo,
		departmentRepo: departmentRepo,
	}
}

// CreateServiceRequest represents a request to create a service
type CreateServiceRequest struct {
	DepartmentID uint   `json:"department_id" binding:"required"`
	Code         string `json:"code" binding:"required"`
	Name         string `json:"name" binding:"required"`
	DisplayName  string `json:"display_name"`
	Description  string `json:"description"`
	Endpoint     string `json:"endpoint"`
	IsActive     bool   `json:"is_active"`
}

// UpdateServiceRequest represents a request to update a service
type UpdateServiceRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
	IsActive    *bool  `json:"is_active"`
}

// CreateService creates a new service
func (uc *serviceUseCase) CreateService(ctx context.Context, req CreateServiceRequest) (*domain.Service, error) {
	// Validate department exists
	_, err := uc.departmentRepo.GetByID(ctx, req.DepartmentID)
	if err != nil {
		return nil, errors.New(errors.ErrCodeNotFound, "department not found", 404)
	}

	// Check if service with same code already exists
	existing, err := uc.serviceRepo.GetByCode(ctx, req.Code)
	if err == nil && existing != nil {
		return nil, errors.New(errors.ErrCodeDuplicateEntry, "service with this code already exists", 409)
	}

	service := &domain.Service{
		DepartmentID: req.DepartmentID,
		Code:         req.Code,
		Name:         req.Name,
		DisplayName:  req.DisplayName,
		Description:  req.Description,
		Endpoint:     req.Endpoint,
		IsActive:     req.IsActive,
		IsSystem:     false,
	}

	if err := uc.serviceRepo.Create(ctx, service); err != nil {
		logger.Error("Failed to create service", zap.Error(err))
		return nil, err
	}

	logger.Info("Service created successfully", zap.String("code", service.Code), zap.Uint("id", service.ID))
	return service, nil
}

// GetService retrieves a service by ID
func (uc *serviceUseCase) GetService(ctx context.Context, id uint) (*domain.Service, error) {
	service, err := uc.serviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// GetServiceByCode retrieves a service by code
func (uc *serviceUseCase) GetServiceByCode(ctx context.Context, code string) (*domain.Service, error) {
	service, err := uc.serviceRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// ListServices lists all services with pagination
func (uc *serviceUseCase) ListServices(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Service], error) {
	result, err := uc.serviceRepo.List(ctx, params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ListServicesByDepartment lists services by department
func (uc *serviceUseCase) ListServicesByDepartment(ctx context.Context, departmentID uint) ([]domain.Service, error) {
	// Validate department exists
	_, err := uc.departmentRepo.GetByID(ctx, departmentID)
	if err != nil {
		return nil, errors.New(errors.ErrCodeNotFound, "department not found", 404)
	}

	services, err := uc.serviceRepo.ListByDepartment(ctx, departmentID)
	if err != nil {
		return nil, err
	}
	return services, nil
}

// UpdateService updates a service
func (uc *serviceUseCase) UpdateService(ctx context.Context, id uint, req UpdateServiceRequest) (*domain.Service, error) {
	// Get existing service
	service, err := uc.serviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if it's a system service
	if service.IsSystem {
		return nil, errors.New(errors.ErrCodeForbidden, "cannot modify system service", 403)
	}

	// Update fields
	if req.Name != "" {
		service.Name = req.Name
	}
	if req.DisplayName != "" {
		service.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		service.Description = req.Description
	}
	if req.Endpoint != "" {
		service.Endpoint = req.Endpoint
	}
	if req.IsActive != nil {
		service.IsActive = *req.IsActive
	}

	if err := uc.serviceRepo.Update(ctx, service); err != nil {
		logger.Error("Failed to update service", zap.Error(err))
		return nil, err
	}

	logger.Info("Service updated successfully", zap.Uint("id", id))
	return service, nil
}

// DeleteService deletes a service
func (uc *serviceUseCase) DeleteService(ctx context.Context, id uint) error {
	// Get service
	service, err := uc.serviceRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if it's a system service
	if service.IsSystem {
		return errors.New(errors.ErrCodeForbidden, "cannot delete system service", 403)
	}

	// TODO: Check if service has dependencies (permissions, etc.)

	if err := uc.serviceRepo.Delete(ctx, id); err != nil {
		logger.Error("Failed to delete service", zap.Error(err))
		return err
	}

	logger.Info("Service deleted successfully", zap.Uint("id", id))
	return nil
}

// ListActiveServices lists all active services
func (uc *serviceUseCase) ListActiveServices(ctx context.Context) ([]domain.Service, error) {
	services, err := uc.serviceRepo.ListActive(ctx)
	if err != nil {
		return nil, err
	}
	return services, nil
}
