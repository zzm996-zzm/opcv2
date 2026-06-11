package postgres

import (
	"context"
	"testing"
)

func TestOpenRejectsInvalidDatabaseURL(t *testing.T) {
	pool, err := Open(context.Background(), "not-a-postgres-url")
	if err == nil {
		if pool != nil {
			pool.Close()
		}
		t.Fatal("Open() error = nil, want invalid URL error")
	}
}
