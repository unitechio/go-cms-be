package page_builder

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/pkg/pagination"
)

type PageUseCase struct {
	pageRepo    repositories.PageRepository
	versionRepo repositories.PageVersionRepository
}

func NewPageUseCase(pageRepo repositories.PageRepository, versionRepo repositories.PageVersionRepository) *PageUseCase {
	return &PageUseCase{
		pageRepo:    pageRepo,
		versionRepo: versionRepo,
	}
}

func (uc *PageUseCase) CreatePage(ctx context.Context, page *domain.Page) error {
	// Set defaults
	if page.Status == "" {
		page.Status = domain.PageStatusDraft
	}
	if page.Template == "" {
		page.Template = "default"
	}
	return uc.pageRepo.Create(ctx, page)
}

func (uc *PageUseCase) GetPage(ctx context.Context, id uuid.UUID) (*domain.Page, error) {
	return uc.pageRepo.GetWithBlocks(ctx, id)
}

func (uc *PageUseCase) UpdatePage(ctx context.Context, page *domain.Page) error {
	// 1. Get existing page to check for changes and create version if needed
	existingPage, err := uc.pageRepo.GetWithBlocks(ctx, page.ID)
	if err != nil {
		return err
	}

	// 2. Create version snapshot before update
	// Note: In a real app, we might only version on publish or specific actions,
	// or have a "save draft" vs "publish" distinction.
	// For now, let's version on every update for simplicity, or maybe we should make it explicit.
	// Let's create a version if it was published or if explicitly requested (logic can be refined).
	// For now, we'll skip auto-versioning on every minor save to avoid bloat,
	// but we should probably have a separate "CreateVersion" action or version on Publish.
	// Let's version when status changes to published or if content changed significantly.
	// For this MVP, let's just update the page. Versioning can be handled by a separate action or hook.

	// Update fields
	existingPage.Title = page.Title
	existingPage.Slug = page.Slug
	existingPage.Template = page.Template
	existingPage.Status = page.Status
	existingPage.SeoTitle = page.SeoTitle
	existingPage.SeoDescription = page.SeoDescription
	existingPage.OgImage = page.OgImage
	// Author shouldn't change typically

	return uc.pageRepo.Update(ctx, existingPage)
}

func (uc *PageUseCase) DeletePage(ctx context.Context, id uuid.UUID) error {
	return uc.pageRepo.Delete(ctx, id)
}

func (uc *PageUseCase) ListPages(ctx context.Context, filter repositories.PageFilter, page *pagination.OffsetPagination) ([]*domain.Page, int64, error) {
	return uc.pageRepo.List(ctx, filter, page)
}

func (uc *PageUseCase) DuplicatePage(ctx context.Context, id uuid.UUID) (*domain.Page, error) {
	originalPage, err := uc.pageRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	newTitle := originalPage.Title + " (Copy)"
	newSlug := originalPage.Slug + "-copy-" + uuid.New().String()[:8]

	return uc.pageRepo.Duplicate(ctx, id, newTitle, newSlug)
}

func (uc *PageUseCase) PublishPage(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// 1. Get page with blocks for snapshot
	page, err := uc.pageRepo.GetWithBlocks(ctx, id)
	if err != nil {
		return err
	}

	// 2. Create version snapshot
	latestVersion, err := uc.versionRepo.GetLatestVersionNumber(ctx, id)
	if err != nil {
		return err
	}

	blocksSnapshot, err := json.Marshal(page.Blocks)
	if err != nil {
		return err
	}

	version := &domain.PageVersion{
		PageID:            id,
		VersionNumber:     latestVersion + 1,
		Title:             page.Title,
		Slug:              page.Slug,
		BlocksSnapshot:    blocksSnapshot,
		CreatedBy:         userID,
		ChangeDescription: "Published page",
	}

	if err := uc.versionRepo.CreateVersion(ctx, version); err != nil {
		return err
	}

	// 3. Update status
	return uc.pageRepo.Publish(ctx, id)
}
