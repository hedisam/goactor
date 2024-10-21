package intprocess

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/hedisam/goactor/internal/mailbox"
	"github.com/hedisam/goactor/sysmsg"
)

type Dispatcher interface {
	PushMessage(ctx context.Context, msg any) error
	PushSystemMessage(ctx context.Context, msg any) error
}

type PID interface {
	Dispatcher
	Ref() string
	AcceptLink(linker PID) error
	AcceptUnlink(linkerRef string)
	AcceptMonitor(pid PID) error
	AcceptDemonitor(monitorRef string)
}

type Registrar interface {
	RegisterSelf(pid *LocalProcess)
	UnregisterSelf()
}

func SpawnLocal(ctx context.Context, logger *slog.Logger, reg Registrar, initFunc InitFunc, msgHandler HandlerFunc, afterFunc AfterFunc, afterTimeout time.Duration) (PID, error) {
	ref := uuid.NewString()
	m := mailbox.NewChanMailbox()
	process := newLocalProcess(logger, ref, m, m)
	initCond := newChanCondOnce[error]()

	go func() {
		var err error
		var toPropagate *sysmsg.Message
		defer func() {
			r := recover()
			process.dispose(ctx, toPropagate, err, r)

			if !initCond.Fired() {
				// we must have either an error or recovered panic from the init function.
				if r != nil {
					err = errors.Join(err, fmt.Errorf("init func failed with panic: %v", r))
				}
				initCond.Signal(err)
			}
		}()

		reg.RegisterSelf(process)
		defer reg.UnregisterSelf()

		err = initFunc(ctx)
		if err != nil {
			return
		}

		initCond.Signal(nil)
		toPropagate, err = process.run(ctx, msgHandler, afterFunc, afterTimeout)
	}()

	err := initCond.Wait()
	if err != nil {
		return nil, fmt.Errorf("init actor: %w", err)
	}

	return process, nil
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
