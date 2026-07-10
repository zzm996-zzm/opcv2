package tasks

import (
	"context"
	"time"
)

type ReminderDispatcher interface {
	DispatchDueTaskReminders(context.Context, int) (int, error)
}

func RunReminderWorker(ctx context.Context, dispatcher ReminderDispatcher, interval time.Duration, onError func(error)) {
	if interval <= 0 {
		interval = time.Minute
	}
	dispatch := func() {
		if _, err := dispatcher.DispatchDueTaskReminders(ctx, 100); err != nil && onError != nil {
			onError(err)
		}
	}

	dispatch()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			dispatch()
		}
	}
}
