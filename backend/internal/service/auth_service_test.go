package service

import (
	"errors"
	"testing"
)

func TestAuthServiceLoginAndLogout(t *testing.T) {
	auth := NewAuthService("clinic-admin", "correct-password")

	if _, err := auth.Login("clinic-admin", "wrong-password"); !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}

	token, err := auth.Login("clinic-admin", "correct-password")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if !auth.ValidateSession(token) {
		t.Fatal("ValidateSession() = false for a new session")
	}

	auth.Logout(token)
	if auth.ValidateSession(token) {
		t.Fatal("ValidateSession() = true after logout")
	}
}

func TestAuthServiceRequiresConfiguredCredentials(t *testing.T) {
	auth := NewAuthService("", "")

	if _, err := auth.Login("clinic-admin", "password"); !errors.Is(err, ErrAuthNotConfigured) {
		t.Fatalf("Login() error = %v, want ErrAuthNotConfigured", err)
	}
}
