package service

import (
	"math"

	"backend/internal/domain"
)

// faceMatchThreshold follows the face_recognition library convention:
// distance <= 0.6 is generally considered the same person.
const faceMatchThreshold = 0.6

func euclideanDistance(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return math.MaxFloat64
	}
	var sum float64
	for i := range a {
		d := a[i] - b[i]
		sum += d * d
	}
	return math.Sqrt(sum)
}

// matchEmbedding finds the closest known dentist to the given embedding.
// ok is false if no dentist is within faceMatchThreshold (i.e. unknown face).
func matchEmbedding(embedding []float64, candidates []domain.Dentist) (match domain.Dentist, distance float64, ok bool) {
	distance = math.MaxFloat64
	for _, d := range candidates {
		if len(d.FaceEmbedding) == 0 {
			continue
		}
		dist := euclideanDistance(embedding, d.FaceEmbedding)
		if dist < distance {
			distance = dist
			match = d
		}
	}
	return match, distance, distance <= faceMatchThreshold
}