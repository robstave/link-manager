package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/robstave/link-manager/internal/controllers"
	"github.com/robstave/link-manager/internal/db"
	"github.com/robstave/link-manager/internal/middleware"
	"github.com/robstave/link-manager/internal/platform/logger"
	"github.com/robstave/link-manager/internal/repositories"
	"github.com/robstave/link-manager/internal/services"
)

func main() {
	ctx := context.Background()
	log := logger.New()

	database, err := db.New(ctx)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()
	log.Info("connected to database")

	authSvc := services.NewAuthService(repositories.NewAuthRepository(database.Pool), log)
	if err := authSvc.EnsureAdminUser(ctx); err != nil {
		log.Error("failed to ensure admin user", "error", err)
		os.Exit(1)
	}

	authController := controllers.NewAuthController(authSvc)
	projectController := controllers.NewProjectController(services.NewProjectService(repositories.NewProjectRepository(database.Pool)))
	categoryController := controllers.NewCategoryController(services.NewCategoryService(repositories.NewCategoryRepository(database.Pool)))
	metaSvc := services.NewMetadataService()
	linkController := controllers.NewLinkController(services.NewLinkService(repositories.NewLinkRepository(database.Pool), metaSvc))
	tagController := controllers.NewTagController(services.NewTagService(repositories.NewTagRepository(database.Pool)))
	metadataController := controllers.NewMetadataController(metaSvc)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","db":"connected"}`)
	})
	mux.HandleFunc("POST /api/v1/auth/login", authController.Login)

	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/v1/auth/me", authController.Me)
	protected.HandleFunc("GET /api/v1/projects", projectController.List)
	protected.HandleFunc("POST /api/v1/projects", projectController.Create)
	protected.HandleFunc("DELETE /api/v1/projects/{id}", projectController.Delete)
	protected.HandleFunc("GET /api/v1/projects/{project_id}/categories", categoryController.List)
	protected.HandleFunc("POST /api/v1/projects/{project_id}/categories", categoryController.Create)
	protected.HandleFunc("DELETE /api/v1/categories/{id}", categoryController.Delete)
	protected.HandleFunc("GET /api/v1/links", linkController.List)
	protected.HandleFunc("POST /api/v1/links", linkController.Create)
	protected.HandleFunc("GET /api/v1/links/{id}", linkController.Get)
	protected.HandleFunc("PUT /api/v1/links/{id}", linkController.Update)
	protected.HandleFunc("DELETE /api/v1/links/{id}", linkController.Delete)
	protected.HandleFunc("POST /api/v1/links/{id}/click", linkController.Click)
	protected.HandleFunc("PATCH /api/v1/links/{id}/stars", linkController.UpdateStars)
	protected.HandleFunc("PATCH /api/v1/links/{id}/cart", linkController.ToggleCart)
	protected.HandleFunc("GET /api/v1/export/links.json", linkController.Export)
	protected.HandleFunc("GET /api/v1/tags", tagController.List)
	protected.HandleFunc("GET /api/v1/meta/title", metadataController.FetchTitle)

	mux.Handle("/api/v1/", middleware.AuthMiddleware(protected))
	handler := middleware.CORS(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Info("server starting", "port", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}
