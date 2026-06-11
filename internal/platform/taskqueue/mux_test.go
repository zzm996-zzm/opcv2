package taskqueue

import (
	"context"
	"testing"

	"github.com/hibiken/asynq"
)

func TestNewMuxRejectsUnregisteredTask(t *testing.T) {
	mux := NewMux()
	task := asynq.NewTask("unknown:task", nil)

	if err := mux.ProcessTask(context.Background(), task); err == nil {
		t.Fatal("ProcessTask() error = nil, want unregistered task error")
	}
}
