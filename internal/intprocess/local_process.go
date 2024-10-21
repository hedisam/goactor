package intprocess

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync/atomic"
	"time"

	"github.com/hedisam/goactor/internal/mailbox"
	"github.com/hedisam/goactor/sysmsg"
)

type (
	HandlerFunc func(ctx context.Context, msg any) error
	AfterFunc   func(context.Context) error
	InitFunc    func(ctx context.Context) error
)

type systemMessageAction int

const (
	systemMessageActionIgnore = iota
	systemMessageActionPropagate
	systemMessageActionDelegate
)

type localReceiver interface {
	Receive(ctx context.Context) (msg any, sysMsg bool, err error)
	ReceiveTimeout(ctx context.Context, d time.Duration) (msg any, sysMsg bool, err error)
	Close()
}

type LocalProcess struct {
	logger        *slog.Logger
	ref           string
	receiver      localReceiver
	msgDispatcher mailbox.Dispatcher
	relations     *relations

	trapExit     atomic.Bool
	disposedFlag atomic.Bool
}

func newLocalProcess(logger *slog.Logger, ref string, r localReceiver, d mailbox.Dispatcher) *LocalProcess {
	return &LocalProcess{
		logger:        logger,
		ref:           ref,
		receiver:      r,
		msgDispatcher: d,
		relations:     newRelations(),
	}
}

func (p *LocalProcess) Ref() string {
	return p.ref
}

func (p *LocalProcess) Dispatcher() mailbox.Dispatcher {
	return p.msgDispatcher
}

// Disposed reports whether this PID is disposed or not.
// Disposed actors neither can be linked/monitored nor can receive messages.
func (p *LocalProcess) disposed() bool {
	return p == nil || p.disposedFlag.Load()
}

func (p *LocalProcess) SetTrapExit(trapExit bool) {
	p.trapExit.Store(trapExit)
}

// Link creates a bidirectional link between the two actors.
// Linked actors gets notified when the other actor exits. If TrapExit is set (see SetTrapExit), the notification
// message gets delegated to the user defined receive function otherwise the linked actor terminates as well if the
// exit reason is anything other than sysmsg.ReasonNormal.
func (p *LocalProcess) Link(pid PID) error {
	if p.disposed() {
		return ErrDisposed
	}
	if pid.disposed() {
		return ErrTargetDisposed
	}
	pid.addRelation(p, relationLinked)
	p.addRelation(pid, relationLinked)
	return nil
}

// Unlink removes the bidirectional link between the two actors.
func (p *LocalProcess) Unlink(pid PID) error {
	if p.disposed() {
		return ErrDisposed
	}
	if !pid.disposed() {
		pid.remRelation(p.ref, relationLinked)
	}
	p.remRelation(pid.Ref(), relationLinked)
	return nil
}

// Monitor monitors the provided PID.
// The user defined receive function of monitor actors receive a sysmsg.Down message when a monitored actor goes down.
func (p *LocalProcess) Monitor(pid PID) error {
	if p.disposed() {
		return ErrDisposed
	}
	if pid.disposed() {
		return ErrTargetDisposed
	}
	pid.addRelation(p, relationMonitor)
	p.addRelation(pid, relationMonitored)
	return nil
}

// Demonitor de-monitors the provided PID.
func (p *LocalProcess) Demonitor(pid PID) error {
	if p.disposed() {
		return ErrDisposed
	}
	if !pid.disposed() {
		pid.remRelation(p.ref, relationMonitor)
	}
	p.remRelation(pid.Ref(), relationMonitored)
	return nil
}

func (p *LocalProcess) addRelation(pid PID, typ relationType) {
	p.relations.Add(pid, typ)
}

func (p *LocalProcess) remRelation(ref string, typ relationType) {
	p.relations.Remove(ref, typ)
}

func (p *LocalProcess) run(ctx context.Context, msgHandler HandlerFunc, afterFunc AfterFunc, afterTimeout time.Duration) (*sysmsg.Message, error) {
	for {
		msg, isSysMsg, err := p.receiver.ReceiveTimeout(ctx, afterTimeout)
		if err != nil {
			switch {
			case errors.Is(err, mailbox.ErrReceiveTimeout):
				err = afterFunc(ctx)
				if err != nil {
					return nil, fmt.Errorf("after timeout handler: %w", err)
				}
				return nil, nil
			case errors.Is(err, context.Canceled),
				errors.Is(err, context.DeadlineExceeded),
				errors.Is(err, mailbox.ErrClosedMailbox):
				return nil, nil
			default:
				return nil, fmt.Errorf("receive incoming messages with timeout: %w", err)
			}
		}
		if isSysMsg {
			sysMsg, ok := msg.(*sysmsg.Message)
			if !ok {
				return nil, fmt.Errorf("non system message received to be handled by system message handler: %T, %+v", msg, msg)
			}
			action, err := p.handleSystemMessage(sysMsg)
			if err != nil {
				return nil, fmt.Errorf("handle system message: %w", err)
			}
			switch action {
			case systemMessageActionPropagate:
				return sysMsg, nil
			case systemMessageActionDelegate:
				msg = sysMsg
			case systemMessageActionIgnore:
				continue
			}
		}

		err = msgHandler(ctx, msg)
		if err != nil {
			return nil, fmt.Errorf("user message handler: %w", err)
		}
	}
}

func (p *LocalProcess) handleSystemMessage(msg *sysmsg.Message) (systemMessageAction, error) {
	trapExit := p.trapExit.Load()

	switch msg.Type {
	case sysmsg.Signal:
		// it's a direct termination signal
		if errors.Is(msg.Reason, sysmsg.ReasonKill) || !trapExit {
			// a direct kill signal cannot be delegated to the user even if trap_exit is set
			return systemMessageActionPropagate, nil
		}
		return systemMessageActionDelegate, nil
	case sysmsg.Down:
		// a monitored actor went down
		p.relations.Remove(msg.ProcessID, relationMonitored)
		return systemMessageActionDelegate, nil
	case sysmsg.Exit:
		// a linked actor reported exit for whatever reason
		p.relations.Remove(msg.ProcessID, relationLinked)
		switch {
		case trapExit:
			return systemMessageActionDelegate, nil
		case errors.Is(msg.Reason, sysmsg.ReasonNormal):
			return systemMessageActionIgnore, nil
		default:
			return systemMessageActionPropagate, nil
		}
	default:
		return systemMessageActionIgnore, fmt.Errorf("unknown system message type received: %+v", msg)
	}
}

func (p *LocalProcess) dispose(ctx context.Context, propagate *sysmsg.Message, runErr error, recovered any) {
	p.receiver.Close()
	p.disposedFlag.Store(true)

	relationTypeToPIDs := p.relations.TypeToRelatedPIDs()

	monitoredActors := relationTypeToPIDs[relationMonitored]
	for pid := range slices.Values(monitoredActors) {
		p.relations.Remove(pid.Ref(), relationMonitor)
	}

	var reason sysmsg.Reason
	switch {
	case recovered != nil:
		reason = fmt.Errorf("panic: %v", recovered)
	case runErr != nil:
		reason = fmt.Errorf("actor runtime: %w", runErr)
	case propagate != nil:
		reason = propagate.Reason
	default:
		reason = sysmsg.ReasonNormal
	}

	p.logger.Debug("Actor is getting disposed, notifying related actors",
		slog.String("actor", p.ref),
		slog.String("reason", reason.Error()),
	)

	p.notify(ctx, sysmsg.Exit, reason, relationTypeToPIDs[relationLinked]...)
	p.notify(ctx, sysmsg.Down, reason, relationTypeToPIDs[relationMonitor]...)
}

func (p *LocalProcess) notify(ctx context.Context, msgType sysmsg.Type, reason sysmsg.Reason, pids ...PID) {
	if len(pids) == 0 {
		return
	}

	// the actor may have been terminated due to a canceled context, therefore we need to make sure we have a
	// non canceled context in order to be able to notify the related actors
	select {
	case <-ctx.Done():
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.WithoutCancel(ctx), time.Second*5)
		defer cancel()
	default:
	}

	// todo: notify concurrently via a worker pool?
	notify := func(who PID) error {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		return mailbox.PushSystemMessage(ctx, who.Dispatcher(), &sysmsg.Message{
			Type:      msgType,
			ProcessID: p.ref,
			Reason:    reason,
		})
	}

	for pid := range slices.Values(pids) {
		err := notify(pid)
		if err != nil {
			p.logger.Warn("Could not send termination message to related actor",
				slog.Any("error", err),
				slog.String("actor", p.ref),
				slog.String("related_actor", pid.Ref()),
			)
		}
	}
}
