package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/robstave/link-manager/internal/auth"
	"github.com/robstave/link-manager/internal/db"
	"github.com/robstave/link-manager/internal/handlers"
	"github.com/robstave/link-manager/internal/middleware"
)

func main() {
	ctx := context.Background()

	// Connect to database
	database, err := db.New(ctx)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	log.Println("Connected to database")

	// Ensure admin user exists
	if err := ensureAdminUser(ctx, database); err != nil {
		log.Fatal("Failed to ensure admin user:", err)
	}

	// Initialize handlers
	authHandler := &handlers.AuthHandler{DB: database}
	projectHandler := &handlers.ProjectHandler{DB: database}
	categoryHandler := &handlers.CategoryHandler{DB: database}
	linkHandler := &handlers.LinkHandler{DB: database}
	tagHandler := &handlers.TagHandler{DB: database}

	// Setup routes
	mux := http.NewServeMux()

	// Health check (no auth)
	mux.HandleFunc("GET /api/v1/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","db":"connected"}`)
	})

	// Auth routes (no auth required)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)

	// Protected routes
	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/v1/auth/me", authHandler.Me)

	// Projects
	protected.HandleFunc("GET /api/v1/projects", projectHandler.List)
	protected.HandleFunc("POST /api/v1/projects", projectHandler.Create)
	protected.HandleFunc("DELETE /api/v1/projects/{id}", projectHandler.Delete)

	// Categories
	protected.HandleFunc("GET /api/v1/projects/{project_id}/categories", categoryHandler.List)
	protected.HandleFunc("POST /api/v1/projects/{project_id}/categories", categoryHandler.Create)
	protected.HandleFunc("DELETE /api/v1/categories/{id}", categoryHandler.Delete)

	// Links
	protected.HandleFunc("GET /api/v1/links", linkHandler.List)
	protected.HandleFunc("POST /api/v1/links", linkHandler.Create)
	protected.HandleFunc("GET /api/v1/links/{id}", linkHandler.Get)
	protected.HandleFunc("DELETE /api/v1/links/{id}", linkHandler.Delete)
	protected.HandleFunc("POST /api/v1/links/{id}/click", linkHandler.Click)
	protected.HandleFunc("PATCH /api/v1/links/{id}/stars", linkHandler.UpdateStars)
	protected.HandleFunc("PATCH /api/v1/links/{id}/cart", linkHandler.ToggleCart)
	protected.HandleFunc("GET /api/v1/export/links.json", linkHandler.Export)

	// Tags
	protected.HandleFunc("GET /api/v1/tags", tagHandler.List)

	// Apply middleware
	mux.Handle("/api/v1/", middleware.AuthMiddleware(protected))

	// Wrap with CORS
	handler := middleware.CORS(mux)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal("Server failed:", err)
	}
}

func ensureAdminUser(ctx context.Context, database *db.DB) error {
	adminUsername := os.Getenv("ADMIN_USERNAME")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if adminUsername == "" {
		adminUsername = "admin"
	}
	if adminPassword == "" {
		adminPassword = "admin"
	}

	// Check if admin user exists
	var exists bool
	err := database.Pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)
	`, adminUsername).Scan(&exists)

	if err != nil {
		return fmt.Errorf("failed to check for admin user: %w", err)
	}

	if !exists {
		log.Printf("Creating admin user: %s", adminUsername)
		user, err := auth.CreateUser(ctx, database.Pool, adminUsername, adminPassword, "admin")
		if err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}

		// Seed with sample links
		err = seedSampleLinks(ctx, database, user.ID)
		if err != nil {
			log.Printf("Warning: failed to seed sample links: %v", err)
		}

		log.Printf("Admin user created successfully")
	} else {
		log.Printf("Admin user already exists")
	}

	return nil
}

func seedSampleLinks(ctx context.Context, database *db.DB, userID string) error {
	// Get default project and category
	var projectID, categoryID string
	err := database.Pool.QueryRow(ctx, `
		SELECT id FROM projects WHERE owner_id = $1 AND is_default = true
	`, userID).Scan(&projectID)
	if err != nil {
		return err
	}

	err = database.Pool.QueryRow(ctx, `
		SELECT id FROM categories WHERE project_id = $1 AND is_default = true
	`, projectID).Scan(&categoryID)
	if err != nil {
		return err
	}

	// Insert sample links
	sampleLinks := []struct {
		url   string
		title string
	}{
		{"https://www.google.com", "Google"},
		{"https://www.github.com", "GitHub"},
		{"https://www.msn.com", "MSN"},
	}

	for _, link := range sampleLinks {
		_, err := database.Pool.Exec(ctx, `
			INSERT INTO links (owner_id, project_id, category_id, url, title)
			VALUES ($1, $2, $3, $4, $5)
		`, userID, projectID, categoryID, link.url, link.title)
		if err != nil {
			return err
		}
	}

	log.Printf("Seeded %d sample links", len(sampleLinks))
	return nil
}
