package goactor

import (
	"context"
	"errors"
	"fmt"
	"github.com/hedisam/goactor/internal/mailbox"
	"github.com/hedisam/goactor/sysmsg"
	"log/slog"
	"os"
	"sync/atomic"
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

// SetLogHandler can be used to set a custom (non slog) log handler for the entire package.
// You should call this function in the beginning of your program. It is not safe to call it when you have
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

// Spawn spawns the provided Actor and returns the corresponding Process Identifier.
// The provided actor can optionally implement ActorInitializer and ActorAfterFuncProvider interfaces.
func Spawn(ctx context.Context, actor Actor) (*PID, error) {
	m := mailbox.NewChanMailbox()
	pid := newPID(m, m)
	initCond := newChanCondOnce[error]()

	go func() {
		var err error
		var toPropagate *sysmsg.Message
		defer func() {
			r := recover()
			pid.dispose(ctx, toPropagate, err, r)

			if !initCond.Fired() {
				// we must have either an error or recovered panic from the init function.
				if r != nil {
					err = errors.Join(err, fmt.Errorf("init func failed with panic: %v", r))
				}
				initCond.Signal(err)
			}
		}()

		registry.registerSelf(pid)
		defer registry.unregisterSelf()

		if initializer, ok := actor.(ActorInitializer); ok {
			err = initializer.Init(ctx, pid)
			if err != nil {
				return
			}
		}
		initCond.Signal(nil)
		toPropagate, err = pid.run(ctx, actor)
	}()

	err := initCond.Wait()
	if err != nil {
		return nil, fmt.Errorf("init actor: %w", err)
	}

	return pid, nil
}

// Send sends a message to an ActorHandler with the provided PID.
func Send(ctx context.Context, processIdentifier ProcessIdentifier, msg any) error {
	pid := processIdentifier.PID()
	if pid == nil {
		if named, ok := processIdentifier.(NamedPID); ok {
			return fmt.Errorf("%w: %q", ErrActorNotFound, named)
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

// chanCondOnce is similar to sync.Cond but Signal can be fired (only once) with a value of T which can be received by Wait.
type chanCondOnce[T any] struct {
	ch    chan T
	fired atomic.Bool
}

func newChanCondOnce[T any]() *chanCondOnce[T] {
	return &chanCondOnce[T]{
		ch: make(chan T),
	}
}

// Signal signals the goroutine blocked by Wait with the provided value.
// Signal can only be fired once, any further signals will be ignored.
// It blocks until Wait receives the value.
func (c *chanCondOnce[T]) Signal(v T) {
	if c.fired.CompareAndSwap(false, true) {
		c.ch <- v
		close(c.ch)
	}
}

// Wait blocks until Signal is fired. It will return immediately if Signal has already been fired.
func (c *chanCondOnce[T]) Wait() T {
	return <-c.ch
}

// Fired returns true if the cond has already signaled.
func (c *chanCondOnce[T]) Fired() bool {
	return c.fired.Load()
}
