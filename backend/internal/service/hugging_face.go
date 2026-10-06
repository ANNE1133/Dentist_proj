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

// HuggingFaceFaceEmbeddingExtractor calls a Hugging Face Inference API (or a
// dedicated Inference Endpoint) running an image feature-extraction / face
// embedding model, and mean-pools its output into a single vector.
//
// Point apiURL at whichever face-embedding model you're using, e.g.
//
//	https://api-inference.huggingface.co/models/<your-org>/<your-face-model>
//
// or a dedicated Inference Endpoint URL. This assumes the model returns a
// raw image embedding for a single face crop — if your model isn't already
// face-cropped, run face detection/alignment before calling Extract, since
// feeding it a full photo (multiple faces, background) will give useless
// embeddings for face-matching.
//
// Satisfies the FaceEmbeddingExtractor interface defined in face_embedding.go.
type HuggingFaceFaceEmbeddingExtractor struct {
	apiURL     string
	apiToken   string
	modelID    string // stored as FaceEmbedding.ModelVersion, e.g. "your-org/your-face-model"
	httpClient *http.Client
}

func NewHuggingFaceFaceEmbeddingExtractor(apiURL, apiToken, modelID string) *HuggingFaceFaceEmbeddingExtractor {
	return &HuggingFaceFaceEmbeddingExtractor{
		apiURL:     apiURL,
		apiToken:   apiToken,
		modelID:    modelID,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// NewHuggingFaceFaceEmbeddingExtractorFromEnv wires this up the way
// docker-compose.yml passes it in: HF_API_URL is optional and, when empty,
// defaults to the public serverless Inference API for HF_MODEL_ID.
//
//	embedder := service.NewHuggingFaceFaceEmbeddingExtractorFromEnv(
//		os.Getenv("HF_API_URL"), os.Getenv("HF_API_TOKEN"), os.Getenv("HF_MODEL_ID"),
//	)
func NewHuggingFaceFaceEmbeddingExtractorFromEnv(apiURL, apiToken, modelID string) *HuggingFaceFaceEmbeddingExtractor {
	if apiURL == "" {
		apiURL = "https://api-inference.huggingface.co/models/" + modelID
	}
	return NewHuggingFaceFaceEmbeddingExtractor(apiURL, apiToken, modelID)
}

type hfErrorResponse struct {
	Error         string  `json:"error"`
	EstimatedTime float64 `json:"estimated_time"`
}

func (h *HuggingFaceFaceEmbeddingExtractor) Extract(ctx context.Context, imageBytes []byte) ([]float64, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.apiURL, bytes.NewReader(imageBytes))
	if err != nil {
		return nil, "", fmt.Errorf("build hugging face request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+h.apiToken)
	// HF's image pipelines accept the raw image bytes directly with this
	// content type rather than a JSON envelope.
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("call hugging face inference api: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read hugging face response: %w", err)
	}

	// The serverless Inference API cold-starts models: a 503 with
	// estimated_time means "loading, try again shortly" rather than a
	// hard failure. Surface that distinctly so the caller can decide
	// whether to retry.
	if resp.StatusCode == http.StatusServiceUnavailable {
		var hfErr hfErrorResponse
		if json.Unmarshal(body, &hfErr) == nil && hfErr.EstimatedTime > 0 {
			return nil, "", fmt.Errorf("hugging face model is still loading, retry in ~%.0fs", hfErr.EstimatedTime)
		}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("hugging face inference api returned status %d: %s", resp.StatusCode, string(body))
	}

	vector, err := poolHuggingFaceEmbedding(body)
	if err != nil {
		return nil, "", fmt.Errorf("parse hugging face embedding: %w", err)
	}
	if len(vector) == 0 {
		return nil, "", fmt.Errorf("no face embedding returned")
	}

	return vector, h.modelID, nil
}

// poolHuggingFaceEmbedding accepts the raw JSON body of an image
// feature-extraction response and mean-pools it down to a single vector.
// HF image-feature-extraction models commonly return one of these shapes:
//
//	[hidden_size]             already a single vector
//	[seq_len][hidden_size]    one vector per patch/token -> mean-pool over seq_len
//	[1][seq_len][hidden_size] batch of 1 -> unwrap, then same as above
func poolHuggingFaceEmbedding(body []byte) ([]float64, error) {
	var raw interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	// unwrap a leading batch dimension of size 1
	if arr, ok := raw.([]interface{}); ok && len(arr) == 1 {
		if _, isNested := arr[0].([]interface{}); isNested {
			raw = arr[0]
		}
	}

	arr, ok := raw.([]interface{})
	if !ok || len(arr) == 0 {
		return nil, fmt.Errorf("unexpected embedding response shape")
	}

	// [hidden_size] - already flat
	if _, isFlat := arr[0].(float64); isFlat {
		out := make([]float64, 0, len(arr))
		for _, x := range arr {
			f, ok := x.(float64)
			if !ok {
				return nil, fmt.Errorf("non-numeric value in embedding response")
			}
			out = append(out, f)
		}
		return out, nil
	}

	// [seq_len][hidden_size] - mean-pool over seq_len
	var sum []float64
	count := 0
	for _, row := range arr {
		rowArr, ok := row.([]interface{})
		if !ok {
			continue
		}
		if sum == nil {
			sum = make([]float64, len(rowArr))
		}
		for i, x := range rowArr {
			f, ok := x.(float64)
			if !ok || i >= len(sum) {
				continue
			}
			sum[i] += f
		}
		count++
	}
	if count == 0 {
		return nil, fmt.Errorf("could not parse embedding rows")
	}

	out := make([]float64, len(sum))
	for i, s := range sum {
		out[i] = s / float64(count)
	}
	return out, nil
}