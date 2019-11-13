package throttler

import (
	"time"
)

// Storage stores all successful attempts of executed actions
type Storage interface {
	// Count how many actions of given group were executed since given time
	CountLastExecuted(action Action, after time.Time) (int, error)

	// Save executed action into storage
	SaveSuccessfulExecution(action Action, at time.Time, expiration time.Duration) error
}
