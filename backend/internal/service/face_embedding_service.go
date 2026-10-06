package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// FaceEmbeddingExtractor turns a raw image into a face embedding vector.
// This repo doesn't include an actual face model, so implement this against
// whatever you run the inference with (a Python/ONNX microservice, AWS
// Rekognition, etc). HTTPFaceEmbeddingExtractor below assumes a small
// internal HTTP service that accepts image bytes and returns a JSON vector.
type FaceEmbeddingExtractor interface {
	Extract(ctx context.Context, imageBytes []byte) (embedding []float64, modelVersion string, err error)
}

type HTTPFaceEmbeddingExtractor struct {
	endpoint   string // python-service /embed endpoint
	httpClient *http.Client
}

func NewHTTPFaceEmbeddingExtractor(endpoint string) *HTTPFaceEmbeddingExtractor {
	return &HTTPFaceEmbeddingExtractor{
		endpoint:   endpoint,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

type embedResponse struct {
	Embedding    []float64 `json:"embedding"`
	ModelVersion string    `json:"modelVersion"`
}

func (h *HTTPFaceEmbeddingExtractor) Extract(ctx context.Context, imageBytes []byte) ([]float64, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.endpoint, bytes.NewReader(imageBytes))
	if err != nil {
		return nil, "", fmt.Errorf("build embedding request: %w", err)
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("call embedding service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var detail struct {
			Detail string `json:"detail"`
		}
		if json.Unmarshal(body, &detail) == nil && detail.Detail != "" {
			return nil, "", fmt.Errorf("%s", detail.Detail)
		}
		return nil, "", fmt.Errorf("embedding service returned status %d", resp.StatusCode)
	}

	var out embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, "", fmt.Errorf("decode embedding response: %w", err)
	}
	if len(out.Embedding) != 128 {
		return nil, "", fmt.Errorf("invalid embedding dimension %d (expected 128)", len(out.Embedding))
	}
	if out.ModelVersion != "face_recognition_dlib_128d_v1" {
		return nil, "", fmt.Errorf("unsupported embedding model %q", out.ModelVersion)
	}
	return out.Embedding, out.ModelVersion, nil
}
