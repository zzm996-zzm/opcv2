package tasks

import (
	"context"
	"testing"
	"time"
)

type fakeReminderDispatcher struct {
	called chan struct{}
}

func (d *fakeReminderDispatcher) DispatchDueTaskReminders(context.Context, int) (int, error) {
	select {
	case d.called <- struct{}{}:
	default:
	}
	return 1, nil
}

func TestReminderWorkerDispatchesImmediatelyAndStopsWithContext(t *testing.T) {
	dispatcher := &fakeReminderDispatcher{called: make(chan struct{}, 1)}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	go func() {
		RunReminderWorker(ctx, dispatcher, time.Hour, nil)
		close(done)
	}()

	select {
	case <-dispatcher.called:
	case <-time.After(time.Second):
		t.Fatal("reminder worker did not dispatch immediately")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("reminder worker did not stop after context cancellation")
	}
}
