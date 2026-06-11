package auth

import (
	"errors"
	"testing"
	"time"
)

func TestJWTManagerIssuesAndParsesAccessToken(t *testing.T) {
	manager := NewJWTManager("test-secret", 15*time.Minute)
	user := User{ID: 42, Nickname: "张晨", Phone: "13800138000", Status: "active"}

	token, expiresAt, err := manager.IssueAccess(user)
	if err != nil {
		t.Fatalf("IssueAccess() error = %v", err)
	}
	if token == "" || time.Until(expiresAt) <= 0 {
		t.Fatalf("token/expiresAt = %q/%v", token, expiresAt)
	}

	userID, err := manager.ParseAccess(token)
	if err != nil {
		t.Fatalf("ParseAccess() error = %v", err)
	}
	if userID != user.ID {
		t.Fatalf("userID = %d, want %d", userID, user.ID)
	}
}

func TestJWTManagerRejectsWrongSecret(t *testing.T) {
	issuer := NewJWTManager("issuer-secret", 15*time.Minute)
	parser := NewJWTManager("different-secret", 15*time.Minute)
	token, _, err := issuer.IssueAccess(User{ID: 42, Status: "active"})
	if err != nil {
		t.Fatalf("IssueAccess() error = %v", err)
	}

	if _, err := parser.ParseAccess(token); !errors.Is(err, ErrInvalidAccessToken) {
		t.Fatalf("ParseAccess() error = %v, want ErrInvalidAccessToken", err)
	}
}
