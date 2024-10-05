package goactor

import (
	"context"
	"errors"
	"fmt"

	"github.com/hedisam/goactor/internal/mailbox"
	"github.com/hedisam/goactor/sysmsg"
)

var (
	// ErrActorNotFound is returned when an ActorHandler cannot be found by a given name.
	ErrActorNotFound = errors.New("no actor was found with the given name")
	// ErrNilPID is returned when trying to send a message using a nil PID
	ErrNilPID = errors.New("cannot send message via a nil PID")
)

// ProcessIdentifier defines a Process Identifier aka PID. It is used to communicate with an Actor.
type ProcessIdentifier interface {
	PID() *PID
}

// Spawn spawns a new actor for the provided ActorFunc and returns the corresponding Process Identifier.
func Spawn(ctx context.Context, fn ReceiveFunc, opts ...ActorOption) *PID {
	config := newActorConfig(fn)
	for _, opt := range opts {
		opt(config)
	}

	m := mailbox.NewChanMailbox()
	pid := newPID(m, m)
	config.initFunc(ctx, pid)

	go func() {
		var runErr error
		var sysMsg *sysmsg.Message
		defer func() {
			pid.dispose(ctx, sysMsg, runErr, recover())
		}()

		sysMsg, runErr = pid.run(ctx, config)
	}()

	return pid
}

// Send sends a message to an ActorHandler with the provided PID.
func Send(ctx context.Context, processIdentifier ProcessIdentifier, msg any) error {
	pid := processIdentifier.PID()
	if pid == nil {
		if _, ok := processIdentifier.(namedPID); ok {
			return fmt.Errorf("%w: %q", ErrActorNotFound, pid)
		}
		return ErrNilPID
	}

	err := pid.dispatcher.PushMessage(ctx, msg)
	if err != nil {
		return fmt.Errorf("push message via dispatcher: %w", err)
	}
	return nil
}
