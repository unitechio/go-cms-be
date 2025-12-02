package page_builder

import (
	"context"

	"github.com/google/uuid"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
)

type PageBlockUseCase struct {
	pageBlockRepo repositories.PageBlockRepository
	blockRepo     repositories.BlockRepository
}

func NewPageBlockUseCase(pageBlockRepo repositories.PageBlockRepository, blockRepo repositories.BlockRepository) *PageBlockUseCase {
	return &PageBlockUseCase{
		pageBlockRepo: pageBlockRepo,
		blockRepo:     blockRepo,
	}
}

func (uc *PageBlockUseCase) AddBlockToPage(ctx context.Context, pageBlock *domain.PageBlock) error {
	// Verify block exists
	_, err := uc.blockRepo.GetByID(ctx, pageBlock.BlockID)
	if err != nil {
		return err
	}

	// TODO: Validate config against block schema
	// block.Schema contains JSON schema, we should validate pageBlock.Config against it

	return uc.pageBlockRepo.AddBlockToPage(ctx, pageBlock)
}

func (uc *PageBlockUseCase) UpdatePageBlock(ctx context.Context, pageBlock *domain.PageBlock) error {
	// TODO: Validate config against block schema

	return uc.pageBlockRepo.UpdatePageBlock(ctx, pageBlock)
}

func (uc *PageBlockUseCase) RemoveBlockFromPage(ctx context.Context, id uuid.UUID) error {
	return uc.pageBlockRepo.RemoveBlockFromPage(ctx, id)
}

func (uc *PageBlockUseCase) ReorderBlocks(ctx context.Context, pageID uuid.UUID, blockOrders map[uuid.UUID]int) error {
	return uc.pageBlockRepo.ReorderBlocks(ctx, pageID, blockOrders)
}
