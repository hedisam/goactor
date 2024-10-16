package goactor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
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

var (
	logger *slog.Logger
)

func init() {
	logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	initRegistry()
}

// SetLogHandler can be used to set a custom (non slog) log handler for entire package.
// You should call this function in the beginning of your program. It is not safe to call this function when you have
// active actors or supervisors. Access to the logger is not guarded by a mutex.
func SetLogHandler(h slog.Handler) {
	logger = slog.New(h)
}

// GetLogger can be used by internal packages to access the logger.
func GetLogger() *slog.Logger {
	return logger
}

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
	initCh := make(chan error)

	go func() {
		var err error
		var toPropagate *sysmsg.Message
		defer func() {
			r := recover()
			pid.dispose(ctx, toPropagate, err, r)

			select {
			case <-initCh:
				// already closed; init was done successfully; error/panic (if any) must've been from running the actor.
			default:
				// not closed yet; init has failed, either with an error or a panic.
				switch {
				case r != nil:
					initCh <- fmt.Errorf("panic during init: %v", r)
				case err != nil:
					initCh <- fmt.Errorf("init failed: %w", err)
				}
				close(initCh)
			}
		}()

		registry.registerSelf(pid)
		defer registry.unregisterSelf()

		err = actor.Init(ctx, pid)
		if err != nil {
			return
		}
		close(initCh)
		toPropagate, err = pid.run(ctx, actor)
	}()

	err := <-initCh
	if err != nil {
		return nil, fmt.Errorf("init actor: %w", err)
	}

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

// Link links self to the provided target PID.
// Link can only be called from the running Actor e.g. from the actor's Init, Receive, or After functions.
func Link(to *PID) error {
	self, err := registry.self()
	if err != nil {
		return fmt.Errorf("get self pid: %w", err)
	}

	err = self.link(to)
	if err != nil {
		return fmt.Errorf("link self to target pid: %w", err)
	}
	return nil
}

// Unlink unlinks self from the linkee.
// Unlink can only be called from the running Actor e.g. from the actor's Init, Receive, or After functions.
func Unlink(linkee *PID) error {
	self, err := registry.self()
	if err != nil {
		return fmt.Errorf("get self pid: %w", err)
	}

	err = self.unlink(linkee)
	if err != nil {
		return fmt.Errorf("unlink self from linkee: %w", err)
	}
	return nil
}

// Monitor monitors the provided PID.
// The user defined receive function of monitor actors receive a sysmsg.Down message when a monitored actor goes down.
// Monitor can only be called from the running Actor e.g. from the actor's Init, Receive, or After functions.
func Monitor(pid *PID) error {
	self, err := registry.self()
	if err != nil {
		return fmt.Errorf("get self pid: %w", err)
	}

	err = self.monitor(pid)
	if err != nil {
		return fmt.Errorf("monitor target pid: %w", err)
	}
	return nil
}

// Demonitor de-monitors the provided PID.
// Demonitor can only be called from the running Actor e.g. from the actor's Init, Receive, or After functions.
func Demonitor(pid *PID) error {
	self, err := registry.self()
	if err != nil {
		return fmt.Errorf("get self pid: %w", err)
	}

	err = self.demonitor(pid)
	if err != nil {
		return fmt.Errorf("demonitor target pid: %w", err)
	}
	return nil
}

// SetTrapExit can be used to trap signals and exit messages from linked actors.
// A direct sysmsg.Signal with a sysmsg.ReasonKill cannot be trapped.
// SetTrapExit can only be called from the running Actor e.g. from the actor's Init, Receive, or After functions.
func SetTrapExit(trapExit bool) error {
	self, err := registry.self()
	if err != nil {
		return fmt.Errorf("get self pid: %w", err)
	}
	self.trapExit.Store(trapExit)
	return nil
}
