package service

import (
	"backend/internal/database"
	"backend/internal/pythonclient"
)

type Service struct {
	db       *database.DB
	python   *pythonclient.Client
	minio    *MinIOService
	embedder FaceEmbeddingExtractor
}

func New(
	db *database.DB,
	python *pythonclient.Client,
	minio *MinIOService,
	embedder FaceEmbeddingExtractor,
) *Service {
	return &Service{
		db:       db,
		python:   python,
		minio:    minio,
		embedder: embedder,
	}
}