package health

import (
	"context"
	"errors"
	"testing"
)

type fakePinger struct {
	err error
}

func (p fakePinger) Ping(context.Context) error {
	return p.err
}

func TestCheckerReadyWhenAllDependenciesRespond(t *testing.T) {
	checker := NewChecker(fakePinger{}, fakePinger{})

	if !checker.Ready(context.Background()) {
		t.Fatal("Ready() = false, want true")
	}
}

func TestCheckerNotReadyWhenDependencyFails(t *testing.T) {
	checker := NewChecker(fakePinger{}, fakePinger{err: errors.New("redis unavailable")})

	if checker.Ready(context.Background()) {
		t.Fatal("Ready() = true, want false")
	}
}
