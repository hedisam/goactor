package goactor

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/hedisam/goactor/internal/mailbox"
	"github.com/hedisam/goactor/internal/syspid"
	"github.com/hedisam/goactor/sysmsg"
)

type systemMessageAction int

const (
	systemMessageActionIgnore = iota
	systemMessageActionPropagate
	systemMessageActionDelegate
)

var (
	// ErrDisposed is returned when this actor is disposed
	ErrDisposed = errors.New("actor is disposed")
	// ErrTargetDisposed is returned when target actor is disposed
	ErrTargetDisposed = errors.New("target actor is disposed")
)

// dispatcher dispatches messages to the actor.
type dispatcher interface {
	PushMessage(ctx context.Context, msg any) error
	PushSystemMessage(ctx context.Context, msg any) error
}

// receiver defines methods required for receiving new messages.
type receiver interface {
	Receive(ctx context.Context) (msg any, sysMsg bool, err error)
	ReceiveTimeout(ctx context.Context, d time.Duration) (msg any, sysMsg bool, err error)
	Close()
}

// PID or ProcessID implements methods required to interact with an ActorHandler.
type PID struct {
	id         string
	dispatcher dispatcher
	receiver   receiver
	relations  *relationsManager
	trapExit   atomic.Bool
	disposed   atomic.Bool
	systemPID  *syspid.PID
}

func newPID(r receiver, d dispatcher) *PID {
	return &PID{
		id:         uuid.NewString(),
		dispatcher: d,
		receiver:   r,
		systemPID:  syspid.New(d),
		relations:  newRelationsManager(),
	}
}

// PID returns the self PID. It implements the ProcessIdentifier interface.
func (p *PID) PID() *PID {
	return p
}

// ID returns the ID of this PID.
func (p *PID) ID() string {
	return p.id
}

// String returns the ID. It implements the Stringer interface.
func (p *PID) String() string {
	return fmt.Sprintf("pid@%s", p.id)
}

// Disposed reports whether this PID is disposed or not.
// Disposed actors neither can be linked/monitored nor can receive messages.
func (p *PID) Disposed() bool {
	return p == nil || p.disposed.Load()
}

// link creates a bidirectional link between the two actors.
// Linked actors gets notified when the other actor exits. If TrapExit is set (see SetTrapExit), the notification
// message gets delegated to the user defined receive function otherwise the linked actor terminates as well if the
// exit reason is anything other than sysmsg.ReasonNormal.
func (p *PID) link(pid *PID) error {
	if p.Disposed() {
		return ErrDisposed
	}
	if pid.Disposed() {
		return ErrTargetDisposed
	}
	pid.relations.Add(p, relationLinked)
	p.relations.Add(pid, relationLinked)
	return nil
}

// unlink removes the bidirectional link between the two actors.
func (p *PID) unlink(pid *PID) error {
	if p.Disposed() {
		return ErrDisposed
	}
	if !pid.Disposed() {
		pid.relations.Remove(p.ID(), relationLinked)
	}
	p.relations.Remove(pid.ID(), relationLinked)
	return nil
}

// monitor monitors the provided PID.
// The user defined receive function of monitor actors receive a sysmsg.Down message when a monitored actor goes down.
func (p *PID) monitor(pid *PID) error {
	if p.Disposed() {
		return ErrDisposed
	}
	if pid.Disposed() {
		return ErrTargetDisposed
	}
	pid.relations.Add(p, relationMonitor)
	p.relations.Add(pid, relationMonitored)
	return nil
}

// demonitor de-monitors the provided PID.
func (p *PID) demonitor(pid *PID) error {
	if p.Disposed() {
		return ErrDisposed
	}
	if !pid.Disposed() {
		pid.relations.Remove(p.ID(), relationMonitor)
	}
	p.relations.Remove(pid.ID(), relationMonitored)
	return nil
}

func (p *PID) run(ctx context.Context, actor Actor) (*sysmsg.Message, error) {
	afterTimeout, afterFunc := actor.AfterFunc()
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

		err = actor.Receive(ctx, msg)
		if err != nil {
			return nil, fmt.Errorf("actor receive: %w", err)
		}
	}
}

func (p *PID) handleSystemMessage(msg *sysmsg.Message) (systemMessageAction, error) {
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

func (p *PID) dispose(ctx context.Context, propagate *sysmsg.Message, runErr error, recovered any) {
	p.receiver.Close()
	p.disposed.Store(true)

	relationTypeToPIDs := p.relations.TypeToRelatedPIDs()

	monitoredActors := relationTypeToPIDs[relationMonitored]
	for pid := range slices.Values(monitoredActors) {
		pid.relations.Remove(p.ID(), relationMonitor)
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

	logger.Debug("Actor is getting disposed, notifying related actors",
		slog.String("actor", p.ID()),
		slog.String("reason", reason.Error()),
	)

	p.notify(ctx, sysmsg.Exit, reason, relationTypeToPIDs[relationLinked]...)
	p.notify(ctx, sysmsg.Down, reason, relationTypeToPIDs[relationMonitor]...)
}

func (p *PID) notify(ctx context.Context, msgType sysmsg.Type, reason sysmsg.Reason, pids ...*PID) {
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
	notify := func(who *PID) error {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		return syspid.Send(ctx, who.systemPID, &sysmsg.Message{
			Type:      msgType,
			ProcessID: p.ID(),
			Reason:    reason,
		})
	}

	for pid := range slices.Values(pids) {
		err := notify(pid)
		if err != nil {
			logger.Warn("Could not send termination message to related actor",
				slog.Any("error", err),
				slog.String("actor", p.ID()),
				slog.String("related_actor", pid.ID()),
			)
		}
	}
}
