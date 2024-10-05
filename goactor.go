package goactor

import (
	"context"
	"errors"
	"fmt"
	"time"

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

// AfterFunc will be called if no messages are received after the specified timeout.
type AfterFunc func(context.Context) error

// Actor defines the methods required by an actor.
type Actor interface {
	// Receive is called when a message is received.
	Receive(ctx context.Context, msg any) (loop bool, err error)
	// Init is called before spawning the Actor when the PID is available.
	Init(ctx context.Context, pid *PID) error
	// AfterFunc specifies a function to be called if no messages are received after the provided timeout.
	AfterFunc() (timeout time.Duration, afterFunc AfterFunc)
}

// Spawn spawns a new actor for the provided ActorFunc and returns the corresponding Process Identifier.
// It returns the InitFunc error, if any. The returned error can be ignored if no InitFunc has been specified.
func Spawn(ctx context.Context, actor Actor) (*PID, error) {
	m := mailbox.NewChanMailbox()
	pid := newPID(m, m)
	err := actor.Init(ctx, pid)
	if err != nil {
		return nil, fmt.Errorf("init actor: %w", err)
	}

	go func() {
		var runErr error
		var sysMsg *sysmsg.Message
		defer func() {
			r := recover()
			pid.dispose(ctx, sysMsg, runErr, r)
		}()

		sysMsg, runErr = pid.run(ctx, actor)
	}()

	return pid, nil
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
