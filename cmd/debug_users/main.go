package main

import (
	"fmt"
	"log"
	"os"

	"github.com/owner/go-cms/internal/config"
	"github.com/owner/go-cms/internal/core/domain"
	"github.com/owner/go-cms/internal/infrastructure/database"
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

	// Count users
	var total int64
	if err := db.Model(&domain.User{}).Count(&total).Error; err != nil {
		log.Fatalf("Failed to count users: %v", err)
	}
	fmt.Printf("Total users in DB: %d\n", total)

	// List users
	var users []domain.User
	if err := db.Find(&users).Error; err != nil {
		log.Fatalf("Failed to list users: %v", err)
	}

	for _, u := range users {
		fmt.Printf("User: ID=%s, Email=%s, Status=%s, DeletedAt=%v\n", u.ID, u.Email, u.Status, u.DeletedAt)
	}
}
