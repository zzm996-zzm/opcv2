package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newRedisTestClient(t *testing.T) *redis.Client {
	t.Helper()
	server := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: server.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestRedisCodeStoreIssuesVerifiesAndConsumesCode(t *testing.T) {
	ctx := context.Background()
	store := NewRedisCodeStore(newRedisTestClient(t))

	if err := store.Issue(ctx, "13800138000", "246810", 5*time.Minute, time.Minute); err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if err := store.Verify(ctx, "13800138000", "246810"); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if err := store.Verify(ctx, "13800138000", "246810"); !errors.Is(err, ErrInvalidCode) {
		t.Fatalf("second Verify() error = %v, want ErrInvalidCode", err)
	}
}

func TestRedisCodeStoreRateLimitsRepeatedIssue(t *testing.T) {
	ctx := context.Background()
	store := NewRedisCodeStore(newRedisTestClient(t))

	if err := store.Issue(ctx, "13800138000", "246810", 5*time.Minute, time.Minute); err != nil {
		t.Fatalf("first Issue() error = %v", err)
	}
	if err := store.Issue(ctx, "13800138000", "246810", 5*time.Minute, time.Minute); !errors.Is(err, ErrCodeRateLimited) {
		t.Fatalf("second Issue() error = %v, want ErrCodeRateLimited", err)
	}
}

func TestRedisSessionStoreRotatesRefreshToken(t *testing.T) {
	ctx := context.Background()
	store := NewRedisSessionStore(newRedisTestClient(t))

	token, err := store.Create(ctx, 42, time.Hour)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	userID, err := store.Consume(ctx, token)
	if err != nil {
		t.Fatalf("Consume() error = %v", err)
	}
	if userID != 42 {
		t.Fatalf("userID = %d, want 42", userID)
	}
	if _, err := store.Consume(ctx, token); !errors.Is(err, ErrInvalidRefreshToken) {
		t.Fatalf("second Consume() error = %v, want ErrInvalidRefreshToken", err)
	}
}
