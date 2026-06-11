package rediscache

import "testing"

func TestNewClientUsesConfiguredAddress(t *testing.T) {
	client := NewClient("redis.internal:6380")
	t.Cleanup(func() { _ = client.Close() })

	if got := client.Options().Addr; got != "redis.internal:6380" {
		t.Fatalf("Addr = %q, want redis.internal:6380", got)
	}
}
