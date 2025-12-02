package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/internal/core/usecases/user"
	"github.com/owner/go-cms/pkg/pagination"
	"github.com/owner/go-cms/pkg/response"
)

// UserHandler handles HTTP requests for users
type UserHandler struct {
	useCase *user.UserUseCase
}

// NewUserHandler creates a new user handler
func NewUserHandler(useCase *user.UserUseCase) *UserHandler {
	return &UserHandler{useCase: useCase}
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user in the system
// @Tags users
// @Accept json
// @Produce json
// @Param user body domain.User true "User object"
// @Success 201 {object} response.Response{data=domain.User}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users [post]
// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user in the system
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User creation request"
// @Success 201 {object} response.Response{data=domain.User}
// @Failure 400 {object} response.Response
// @Failure 409 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err.Error())
		return
	}

	// Resolve department name/code to ID if provided
	var departmentID *uint
	if req.Department != "" {
		// User requested to use code for better reliability
		dept, err := h.useCase.GetDepartmentByCode(c.Request.Context(), req.Department)
		if err != nil {
			response.BadRequest(c, "Invalid department code: "+req.Department)
			return
		}
		departmentID = &dept.ID
	}

	// Convert DTO to domain model
	user := req.ToUser(departmentID)

	// Create user
	if err := h.useCase.CreateUser(c.Request.Context(), user); err != nil {
		response.Error(c, err)
		return
	}

	// Assign roles if provided
	if len(req.RoleIDs) > 0 {
		for _, roleID := range req.RoleIDs {
			if err := h.useCase.AssignRole(c.Request.Context(), user.ID, roleID); err != nil {
				// Log error but don't fail the entire request
				// The user is already created
				response.Error(c, err)
				return
			}
		}
	}

	// Reload user with relationships
	createdUser, err := h.useCase.GetUser(c.Request.Context(), user.ID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, createdUser)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Get detailed information about a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Response{data=domain.User}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	user, err := h.useCase.GetUser(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, user)
}

// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Param search query string false "Search query"
// @Param status query string false "Filter by status"
// @Param role_id query int false "Filter by role ID"
// @Success 200 {object} response.Response{data=[]domain.User}
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, err := pagination.ParseOffsetRequest(c.Query("page"), c.Query("limit"), 10, 100)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	filter := repositories.UserFilter{
		Search: c.Query("search"),
		Status: domain.UserStatus(c.Query("status")),
	}

	if roleID := c.Query("role_id"); roleID != "" {
		id, err := strconv.ParseUint(roleID, 10, 32)
		if err == nil {
			uid := uint(id)
			filter.RoleID = &uid
		}
	}

	users, total, err := h.useCase.ListUsers(c.Request.Context(), filter, page)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, users, total, page)
}

// UpdateUser godoc
// @Summary Update a user
// @Description Update an existing user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body domain.User true "User update object"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var updates domain.User
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.useCase.UpdateUser(c.Request.Context(), id, &updates); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "User updated successfully"})
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a user from the system
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := h.useCase.DeleteUser(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "User deleted successfully"})
}

// GetUserRoles godoc
// @Summary Get user roles
// @Description Get all roles assigned to a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} response.Response{data=[]domain.Role}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users/{id}/roles [get]
func (h *UserHandler) GetUserRoles(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	roles, err := h.useCase.GetUserRoles(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, roles)
}

// AssignRole godoc
// @Summary Assign role to user
// @Description Assign a role to a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param request body object{role_id=uint} true "Role assignment"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users/{id}/roles [post]
func (h *UserHandler) AssignRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req struct {
		RoleID uint `json:"role_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.useCase.AssignRole(c.Request.Context(), userID, req.RoleID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Role assigned successfully"})
}

// RemoveRole godoc
// @Summary Remove role from user
// @Description Remove a role from a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param roleId path int true "Role ID"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Security BearerAuth
// @Router /users/{id}/roles/{roleId} [delete]
func (h *UserHandler) RemoveRole(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	roleID, err := strconv.ParseUint(c.Param("roleId"), 10, 32)
	if err != nil {
		response.BadRequest(c, "Invalid role ID")
		return
	}

	if err := h.useCase.RemoveRole(c.Request.Context(), userID, uint(roleID)); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{"message": "Role removed successfully"})
}
