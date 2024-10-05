package syspid

import "context"

// Dispatcher is a used to dispatch system messages.
type Dispatcher interface {
	PushSystemMessage(ctx context.Context, msg any) error
}

// PID is a system PID.
type PID struct {
	dispatcher Dispatcher
}

// New returns a new system PID.
func New(d Dispatcher) *PID {
	return &PID{
		dispatcher: d,
	}
}

// Send sends a system message to the specified system PID.
func Send(ctx context.Context, pid *PID, msg any) error {
	return pid.dispatcher.PushSystemMessage(ctx, msg)
}
