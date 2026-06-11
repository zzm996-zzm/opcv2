package health

import (
	"context"
	"time"
)

type Pinger interface {
	Ping(context.Context) error
}

type Checker struct {
	dependencies []Pinger
}

func NewChecker(dependencies ...Pinger) *Checker {
	return &Checker{dependencies: dependencies}
}

func (c *Checker) Ready(ctx context.Context) bool {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	for _, dependency := range c.dependencies {
		if dependency == nil || dependency.Ping(ctx) != nil {
			return false
		}
	}
	return true
}
