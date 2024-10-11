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
	// SystemPID has to be exported to be accessible from the supervision package
	SystemPID *syspid.PID
}

func newPID(r receiver, d dispatcher) *PID {
	return &PID{
		id:         uuid.NewString(),
		dispatcher: d,
		receiver:   r,
		SystemPID:  syspid.New(d),
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

// SetTrapExit sets trap exit. If set to true, terminating linked actors won't cascade to this actor and only
// the exit message will be received and processed.
func (p *PID) SetTrapExit(trapExit bool) {
	p.trapExit.Store(trapExit)
}

// String returns the ID. It implements the Stringer interface.
func (p *PID) String() string {
	return fmt.Sprintf("pid@%s", p.id)
}

// Link creates a bidirectional relationship between the two actors.
func (p *PID) Link(pid *PID) {
	pid.relations.Add(p, relationLinked)
	p.relations.Add(pid, relationLinked)
}

func (p *PID) Unlink(pid *PID) {
	pid.relations.Remove(p.ID(), relationLinked)
	p.relations.Remove(pid.ID(), relationLinked)
}

func (p *PID) Monitor(pid *PID) {
	pid.relations.Add(p, relationMonitor)
	p.relations.Add(pid, relationMonitored)
}

func (p *PID) Demonitor(pid *PID) {
	pid.relations.Remove(p.ID(), relationMonitor)
	p.relations.Remove(pid.ID(), relationMonitored)
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
		var sysMsg *sysmsg.Message
		if isSysMsg {
			var ok bool
			sysMsg, ok = sysmsg.ToSystemMessage(msg)
			if !ok {
				return nil, fmt.Errorf("non system message received to be handled by system message handler: %T", msg)
			}
			delegate, err := p.handleSystemMessage(sysMsg)
			if err != nil {
				return sysMsg, fmt.Errorf("handle system message: %w", err)
			}
			if !delegate {
				continue
			}
		}
		loop, err := actor.Receive(ctx, msg)
		if err != nil {
			if isSysMsg {
				return sysMsg, fmt.Errorf("msg handler with sys message: %w", err)
			}
			return nil, fmt.Errorf("msg handler: %w", err)
		}
		if !loop {
			break
		}
	}

	return nil, nil
}

func (p *PID) handleSystemMessage(msg *sysmsg.Message) (delegate bool, err error) {
	relations := p.relations.Relations(msg.SenderID)
	if len(relations) == 0 {
		return false, nil
	}

	p.relations.Purge(msg.SenderID)
	trapExit := p.trapExit.Load()

	switch {
	case slices.Contains(relations, relationLinked):
		switch {
		case trapExit:
			return true, nil
		case msg.Type == sysmsg.NormalExit:
			return false, nil
		default:
			return false, fmt.Errorf("linked actor %q received %q message", msg.SenderID, msg.Type)
		}
	case slices.Contains(relations, relationMonitored):
		return true, nil
	default:
		return false, nil
	}
}

func (p *PID) dispose(ctx context.Context, origin *sysmsg.Message, err error, recovered any) {
	p.receiver.Close()

	reason := any("normal exit")
	typ := sysmsg.NormalExit
	if err != nil {
		reason = err
		typ = sysmsg.AbnormalExit
	}

	if recovered != nil {
		reason = recovered
		typ = sysmsg.AbnormalExit
	}

	logger.Debug("Actor is getting disposed, notifying related actors",
		slog.String("actor", p.ID()),
		slog.String("exit_type", string(typ)),
		"reason", reason,
	)
	p.notifyRelations(ctx, &sysmsg.Message{
		SenderID: p.ID(),
		Reason:   reason,
		Type:     typ,
		Origin:   origin,
	})
}

func (p *PID) notifyRelations(ctx context.Context, msg *sysmsg.Message) {
	var ctxCancels []func()
	defer func() {
		for _, fn := range ctxCancels {
			fn()
		}
	}()

	// todo: notify concurrently via a worker pool? also, the ctx could be canceled
	notify := func(who *PID) error {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		ctxCancels = append(ctxCancels, cancel)
		return syspid.Send(ctx, who.SystemPID, msg)
	}

	p.relations.mu.Lock()
	defer p.relations.mu.Unlock()

	for _, related := range p.relations.idToPID {
		err := notify(related)
		if err != nil {
			logger.Warn("Could not notify related actor about exit status",
				"error", err,
				"actor", p.ID(),
				"related_actor", related.ID(),
			)
		}
	}
}
