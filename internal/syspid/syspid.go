package syspid

import "context"

type Dispatcher interface {
	PushSystemMessage(ctx context.Context, msg any) error
}

type PID struct {
	id         string
	dispatcher Dispatcher
}

func NewSystemPID(id string, d Dispatcher) *PID {
	return &PID{
		id:         id,
		dispatcher: d,
	}
}

func Send(ctx context.Context, pid *PID, msg any) error {
	return pid.dispatcher.PushSystemMessage(ctx, msg)
}

func ID(pid *PID) string {
	return pid.id
}
