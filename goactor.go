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
)

// ProcessIdentifier defines a Process Identifier aka PID. It is used to communicate with an Actor.
type ProcessIdentifier interface {
	PID() *PID
}

// Spawn spawns a new actor for the provided ActorFunc and returns the corresponding Process Identifier.
func Spawn(ctx context.Context, fn ReceiveFunc, opts ...ActorOption) *PID {
	a := newActor(fn)
	for _, opt := range opts {
		opt(a)
	}

	m := mailbox.NewChanMailbox()
	pid := newPID(m, m)
	a.initFunc(ctx, pid)

	go func() {
		var runErr error
		var sysMsg *sysmsg.Message
		defer func() {
			pid.dispose(ctx, sysMsg, runErr)
		}()

		sysMsg, runErr = pid.run(ctx, a)
	}()

	return pid
}

// Send sends a message to an ActorHandler with the provided PID.
func Send(ctx context.Context, pid ProcessIdentifier, msg any) error {
	if pid.PID() == nil {
		return errors.New("cannot send message via a nil PID")
	}

	err := pid.PID().dispatcher.PushMessage(ctx, msg)
	if err != nil {
		return fmt.Errorf("push message via dispatcher: %w", err)
	}
	return nil
}

// SendNamed sends a message to an ActorHandler via its associated name.
func SendNamed(ctx context.Context, name string, msg any) error {
	pid, ok := WhereIs(name)
	if !ok {
		return fmt.Errorf("%w %q", ErrActorNotFound, name)
	}

	return Send(ctx, pid, msg)
}
