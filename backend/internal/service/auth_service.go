package service

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"sync"
	"time"
)

var (
	ErrAuthNotConfigured  = errors.New("login credentials are not configured")
	ErrInvalidCredentials = errors.New("invalid username or password")
)

const SessionDuration = 12 * time.Hour

type AuthService struct {
	username string
	password string
	mu       sync.Mutex
	sessions map[string]time.Time
}

func NewAuthService(username, password string) *AuthService {
	return &AuthService{
		username: username,
		password: password,
		sessions: make(map[string]time.Time),
	}
}

func (s *AuthService) Login(username, password string) (string, error) {
	if s.username == "" || s.password == "" {
		return "", ErrAuthNotConfigured
	}

	configuredPasswordHash := sha256.Sum256([]byte(s.password))
	providedPasswordHash := sha256.Sum256([]byte(password))
	if username != s.username || subtle.ConstantTimeCompare(providedPasswordHash[:], configuredPasswordHash[:]) != 1 {
		return "", ErrInvalidCredentials
	}

	randomToken := make([]byte, 32)
	if _, err := rand.Read(randomToken); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(randomToken)

	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.removeExpiredSessions(now)
	s.sessions[token] = now.Add(SessionDuration)
	return token, nil
}

func (s *AuthService) ValidateSession(token string) bool {
	if token == "" {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	expiresAt, exists := s.sessions[token]
	if !exists {
		return false
	}
	if time.Now().After(expiresAt) {
		delete(s.sessions, token)
		return false
	}
	return true
}

func (s *AuthService) Logout(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, token)
}

func (s *AuthService) removeExpiredSessions(now time.Time) {
	for token, expiresAt := range s.sessions {
		if now.After(expiresAt) {
			delete(s.sessions, token)
		}
	}
}
