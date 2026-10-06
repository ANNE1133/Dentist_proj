package pythonclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// ---------- /detect ----------

type DetectedFace struct {
	BBox      [4]int    `json:"bbox"` // x, y, w, h in pixels
	Embedding []float64 `json:"embedding"`
}

type detectRequest struct {
	ImageBase64 string `json:"imageBase64"`
}

type detectResponse struct {
	Faces []DetectedFace `json:"faces"`
}

// Detect sends one frame to the python service and returns all faces found,
// each with its bounding box and raw face embedding (128-d).
func (c *Client) Detect(imageBase64 string) ([]DetectedFace, error) {
	body, err := json.Marshal(detectRequest{ImageBase64: imageBase64})
	if err != nil {
		return nil, fmt.Errorf("marshal detect request: %w", err)
	}

	resp, err := c.http.Post(c.baseURL+"/detect", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("call python /detect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("python /detect returned %s", resp.Status)
	}

	var out detectResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode detect response: %w", err)
	}
	return out.Faces, nil
}

// ---------- /register ----------

type registerRequest struct {
	ImageBase64 string `json:"imageBase64"`
	DentistName string `json:"dentistName"`
}

type RegisterResult struct {
	FaceEmbedding []float64 `json:"faceEmbedding"`
	ChairHeight   float64   `json:"chairHeight"`
}

// Register sends a registration photo to the python service and returns the
// face embedding + estimated chair height to store alongside the dentist.
func (c *Client) Register(imageBase64, dentistName string) (*RegisterResult, error) {
	body, err := json.Marshal(registerRequest{ImageBase64: imageBase64, DentistName: dentistName})
	if err != nil {
		return nil, fmt.Errorf("marshal register request: %w", err)
	}

	resp, err := c.http.Post(c.baseURL+"/register", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("call python /register: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("python /register returned %s", resp.Status)
	}

	var out RegisterResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode register response: %w", err)
	}
	return &out, nil
}