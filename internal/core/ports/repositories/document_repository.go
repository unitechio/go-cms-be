package repositories

import (
	"context"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/dto"
)

type DocumentRepository interface {
	CreateDocument(ctx context.Context, document *domain.Document) error
	UpdateDocument(ctx context.Context, document *domain.Document) error
	DeleteDocumentByID(ctx context.Context, id uint) error
	GetDocumentByID(ctx context.Context, id uint) (*domain.Document, error)
	GetDocumentByCode(ctx context.Context, code string) (*domain.Document, error)
	GetDocumentByPath(ctx context.Context, path string) (*domain.Document, error)
	GetDocuments(ctx context.Context, filter dto.DocumentFilter) ([]domain.Document, int, int, error)
	GetDocumentsByEntityID(ctx context.Context, entityType string, entityID uint) ([]domain.Document, error)

	// Permission related methods
	CreateDocumentPermission(ctx context.Context, permission *domain.DocumentPermission) error
	UpdateDocumentPermission(ctx context.Context, permission *domain.DocumentPermission) error
	DeleteDocumentPermission(ctx context.Context, id uint) error
	GetDocumentPermissions(ctx context.Context, documentID uint) ([]domain.DocumentPermission, error)
	GetUserDocumentPermission(ctx context.Context, documentID uint, userID uuid.UUID) (*domain.DocumentPermission, error)
	CheckUserPermission(ctx context.Context, documentID uint, userID uuid.UUID, requiredLevel string) (bool, error)

	// Comment related methods
	CreateDocumentComment(ctx context.Context, comment *domain.DocumentComment) error
	UpdateDocumentComment(ctx context.Context, comment *domain.DocumentComment) error
	DeleteDocumentComment(ctx context.Context, id uint) error
	GetDocumentComments(ctx context.Context, documentID uint) ([]domain.DocumentComment, error)

	// Version related methods
	CreateDocumentVersion(ctx context.Context, version *domain.DocumentVersion) error
	GetDocumentVersions(ctx context.Context, documentID uint) ([]domain.DocumentVersion, error)
}
type DocumentUsecase interface {
	UploadDocument(ctx context.Context, file *multipart.FileHeader, uploadRequest dto.DocumentUploadRequest, userID uuid.UUID) (*domain.Document, error)
	UpdateDocument(ctx context.Context, id uint, updateRequest dto.DocumentUpdateRequest, userID uuid.UUID) (*domain.Document, error)
	DeleteDocument(ctx context.Context, id uint, userID uuid.UUID) error
	GetDocumentByID(ctx context.Context, id uint, userID uuid.UUID) (*domain.Document, error)
	GetDocumentByCode(ctx context.Context, code string, userID uuid.UUID) (*domain.Document, error)
	GetDocuments(ctx context.Context, filter dto.DocumentFilter, userID uuid.UUID) (*dto.PaginatedDocumentsResponse, error)
	GetDocumentsByEntityID(ctx context.Context, entityType string, entityID uint, userID uuid.UUID) ([]dto.DocumentResponse, error)
	DownloadDocument(ctx context.Context, id uint, userID uuid.UUID) ([]byte, string, string, error)

	// Permission management
	AddDocumentPermission(ctx context.Context, request dto.DocumentPermissionRequest, userID uuid.UUID) error
	UpdateDocumentPermission(ctx context.Context, id uint, request dto.DocumentPermissionRequest, userID uuid.UUID) error
	RemoveDocumentPermission(ctx context.Context, id uint, userID uuid.UUID) error
	GetDocumentPermissions(ctx context.Context, documentID uint, userID uuid.UUID) ([]domain.DocumentPermission, error)
	CheckUserPermission(ctx context.Context, documentID uint, userID uuid.UUID, requiredLevel string) (bool, error)

	// Document comments
	AddDocumentComment(ctx context.Context, request dto.DocumentCommentRequest, userID uuid.UUID) (*domain.DocumentComment, error)
	UpdateDocumentComment(ctx context.Context, id uint, comment string, userID uuid.UUID) (*domain.DocumentComment, error)
	DeleteDocumentComment(ctx context.Context, id uint, userID uuid.UUID) error
	GetDocumentComments(ctx context.Context, documentID uint, userID uuid.UUID) ([]domain.DocumentComment, error)

	// Document versions
	GetDocumentVersions(ctx context.Context, documentID uint, userID uuid.UUID) ([]domain.DocumentVersion, error)
}
