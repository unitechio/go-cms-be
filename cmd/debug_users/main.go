package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/owner/go-cms/internal/adapters/repositories/postgres"
	"github.com/owner/go-cms/internal/config"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/core/ports/repositories"
	"github.com/owner/go-cms/internal/infrastructure/database"
	"github.com/owner/go-cms/pkg/pagination"
)

func main() {
	// Load config
	os.Setenv("APP_ENV", "development")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect DB
	if err := database.Init(&cfg.Database); err != nil {
		log.Fatalf("Failed to connect DB: %v", err)
	}
	db := database.GetDB()

	fmt.Println("=== DEBUG USER LIST API ===\n")

	// 1. Count total users in DB
	var total int64
	if err := db.Model(&domain.User{}).Count(&total).Error; err != nil {
		log.Fatalf("Failed to count users: %v", err)
	}
	fmt.Printf("1. Total users in DB: %d\n\n", total)

	// 2. List all users directly
	var users []domain.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("Failed to list users: %v", err)
	}
	fmt.Printf("2. Direct query found %d users:\n", len(users))
	for i, u := range users {
		fmt.Printf("   [%d] ID=%s, Email=%s, Status=%s, DeletedAt=%v\n",
			i+1, u.ID, u.Email, u.Status, u.DeletedAt)
	}
	fmt.Println()

	// 3. Test repository List method
	userRepo := postgres.NewUserRepository(db)
	ctx := context.Background()

	filter := repositories.UserFilter{}
	page := &pagination.OffsetPagination{
		Page:  1,
		Limit: 10,
	}

	repoUsers, repoTotal, err := userRepo.List(ctx, filter, page)
	if err != nil {
		log.Fatalf("Repository List failed: %v", err)
	}

	fmt.Printf("3. Repository List method:\n")
	fmt.Printf("   Total: %d\n", repoTotal)
	fmt.Printf("   Found: %d users\n", len(repoUsers))
	for i, u := range repoUsers {
		fmt.Printf("   [%d] ID=%s, Email=%s, Status=%s\n",
			i+1, u.ID, u.Email, u.Status)
	}
	fmt.Println()

	// 4. Test with search filter
	searchFilter := repositories.UserFilter{
		Search: "admin",
	}
	searchUsers, searchTotal, err := userRepo.List(ctx, searchFilter, page)
	if err != nil {
		log.Fatalf("Repository List with search failed: %v", err)
	}

	fmt.Printf("4. Repository List with search='admin':\n")
	fmt.Printf("   Total: %d\n", searchTotal)
	fmt.Printf("   Found: %d users\n", len(searchUsers))
	for i, u := range searchUsers {
		fmt.Printf("   [%d] ID=%s, Email=%s\n", i+1, u.ID, u.Email)
	}
	fmt.Println()

	// 5. Check for soft deletes
	var deletedCount int64
	db.Unscoped().Model(&domain.User{}).Where("deleted_at IS NOT NULL").Count(&deletedCount)
	fmt.Printf("5. Soft deleted users: %d\n", deletedCount)

	fmt.Println("\n=== DEBUG COMPLETE ===")
}
