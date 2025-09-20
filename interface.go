package launcher

import "context"

// Runnable defines the interface for an application or service that can be launched.
// Any type that implements this interface can be managed by the Launcher.
type Runnable interface {
	// Run starts the execution of the launchable component.
	// The provided context should be used for cancellation. If the context is
	// canceled, the Run method should gracefully shut down and return.
	// It should return an error if the launch fails. The error will be
	// propagated up to the Launcher.
	Run(ctx context.Context) error
}
