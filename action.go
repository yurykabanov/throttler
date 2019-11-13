package throttler

// Action is any task that can be run and should be throttled
type Action interface {
	// Key to group tasks
	GroupID() string

	// Task implementation
	Run() error
}
