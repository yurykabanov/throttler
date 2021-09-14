package throttler

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrToManyActionsExecuted = errors.New("too many actions executed")
)

// Throttler is a core object that runs actions.
type Throttler struct {
	clock clock

	// maxAllowedActions is how much actions can be executed within time window
	maxAllowedActions int
	// period is time window
	period time.Duration

	// storage is persistent storage to save all successful attempts of running a task
	storage Storage
}

func New(maxAllowedActions int, period time.Duration, storage Storage) *Throttler {
	return &Throttler{
		clock: defaultClockInstance,

		maxAllowedActions: maxAllowedActions,
		period:            period,

		storage: storage,
	}
}

// Execute given action if limit is not reached.
// Only successful attempts are counted (when no error returned upon calling `action.Run()`).
func (t *Throttler) Execute(ctx context.Context, action Action) error {
	actionsCount, err := t.storage.CountLastExecuted(ctx, action, t.clock.Now().Add(-t.period))
	if err != nil {
		return fmt.Errorf("error querying the storage: %w", err)
	}

	if actionsCount >= t.maxAllowedActions {
		return ErrToManyActionsExecuted
	}

	if err := action.Run(); err != nil {
		return err
	}

	err = t.storage.SaveSuccessfulExecution(ctx, action, t.clock.Now(), t.period)
	if err != nil {
		return fmt.Errorf("error while storing successful execution: %w", err)
	}

	return nil
}
