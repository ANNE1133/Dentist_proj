package main

import (
	"log"
	"net/http"
	"os"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/handlers"
	"backend/internal/pythonclient"
	"backend/internal/service"

	"github.com/rs/cors"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	pyClient := pythonclient.New(cfg.PythonServiceURL)

	// MinIO: stores the dentist photo. Reading straight from env here since
	// config.Config doesn't have MinIO fields yet — move these into cfg
	// (with a cfg.DSN()-style helper) if you'd rather keep all config in
	// one place.
	minioSvc, err := service.NewMinIOService(
		os.Getenv("MINIO_ENDPOINT"),
		os.Getenv("MINIO_ACCESS_KEY"),
		os.Getenv("MINIO_SECRET_KEY"),
		os.Getenv("MINIO_BUCKET"),
		os.Getenv("MINIO_USE_SSL") == "true",
	)
	if err != nil {
		log.Fatalf("failed to connect to minio: %v", err)
	}

	// Registration and live frames use the same face_recognition implementation.
	embedder := service.NewHTTPFaceEmbeddingExtractor(cfg.PythonServiceURL + "/embed")

	svc := service.New(db, pyClient, minioSvc, embedder)
	h := handlers.NewHandler(svc)

	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins: []string{
			"http://localhost:5173",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Content-Type",
		},
		AllowCredentials: true,
	})

	handler := c.Handler(mux)

	addr := ":8080"

	log.Printf("server listening on %s", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
