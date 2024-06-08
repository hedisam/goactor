package goactor

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/hedisam/goactor/internal/mailbox"
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

type relationType int

const (
	relNone = iota
	relLinked
	relLinkedTrapExit
	relMonitored
	relMonitor
)

// PID or ProcessID implements methods required to interact with an ActorHandler.
type PID struct {
	id         string
	dispatcher dispatcher
	r          receiver

	relationsMu   sync.RWMutex
	links         map[string]*PID
	trapExitLinks map[string]struct{}
	monitored     map[string]*PID
	monitors      map[string]*PID
}

// PID returns the self PID. It implements the ProcessIdentifier interface.
func (pid *PID) PID() *PID {
	return pid
}

func newPID(r receiver, d dispatcher) *PID {
	return &PID{
		id:         uuid.NewString(),
		dispatcher: d,
		r:          r,
		links:      map[string]*PID{},
		monitors:   map[string]*PID{},
		monitored:  map[string]*PID{},
	}
}

// String returns the ID. It implements the Stringer interface.
func (pid *PID) String() string {
	return fmt.Sprintf("pid@%s", pid.id)
}

// Link creates a bidirectional relationship between the two actors.
func (pid *PID) Link(to *PID, trapExit bool) {
	to.addLink(pid)

	pid.relationsMu.Lock()
	defer pid.relationsMu.Unlock()
	pid.links[to.id] = to
	if trapExit {
		pid.trapExitLinks[to.id] = struct{}{}
	}
}

func (pid *PID) addLink(from *PID) {
	pid.relationsMu.Lock()
	defer pid.relationsMu.Unlock()
	pid.links[from.id] = from
}

func (pid *PID) Unlink(from *PID) {
	from.removeLink(pid)

	pid.relationsMu.Lock()
	defer pid.relationsMu.Unlock()
	delete(pid.links, from.id)
}

func (pid *PID) removeLink(from *PID) {
	pid.relationsMu.Lock()
	defer pid.relationsMu.Unlock()
	delete(pid.links, from.id)
}

func (pid *PID) Monitor(who *PID) {
	who.addMonitor(pid)

	pid.relationsMu.Lock()
	defer pid.relationsMu.Unlock()
	pid.monitored[who.id] = who
}

func (pid *PID) addMonitor(m *PID) {
	pid.relationsMu.Lock()
	defer pid.relationsMu.Unlock()

	pid.monitors[m.id] = m
}

func (pid *PID) Demonitor(who *PID) {
	who.removeMonitor(pid)

	pid.relationsMu.Lock()
	defer pid.relationsMu.Unlock()
	delete(pid.monitored, who.id)
}

func (pid *PID) removeMonitor(m *PID) {
	pid.relationsMu.Lock()
	defer pid.relationsMu.Unlock()

	delete(pid.monitors, m.id)
}

func (pid *PID) relation(who *PID) relationType {
	pid.relationsMu.RLock()
	defer pid.relationsMu.RUnlock()

	id := who.id
	_, ok := pid.links[id]
	if ok {
		_, ok = pid.trapExitLinks[id]
		if ok {
			return relLinkedTrapExit
		}
		return relLinked
	}
	_, ok = pid.monitored[id]
	if ok {
		return relMonitored
	}
	_, ok = pid.monitors[id]
	if ok {
		return relMonitor
	}
	return relNone
}

func (pid *PID) removeRelation(who *PID, rel relationType) {
	if rel == relNone {
		return
	}

	pid.relationsMu.Lock()
	defer pid.relationsMu.Unlock()

	switch rel {
	case relLinked:
		delete(pid.links, who.id)
	case relLinkedTrapExit:
		delete(pid.links, who.id)
		delete(pid.trapExitLinks, who.id)
	case relMonitored:
		delete(pid.monitored, who.id)
	case relMonitor:
		delete(pid.monitors, who.id)
	default:
		return
	}
}

func (pid *PID) run(ctx context.Context, a *actor) (*SystemMessage, error) {
	for {
		msg, isSysMsg, err := pid.r.ReceiveTimeout(ctx, a.receiveTimeoutDuration)
		if err != nil {
			if !errors.Is(err, mailbox.ErrReceiveTimeout) {
				return nil, fmt.Errorf("receive incoming messages with timeout: %w", err)
			}
			err = a.afterTimeoutFunc(ctx)
			if err != nil {
				return nil, fmt.Errorf("after timeout handler: %w", err)
			}
			return nil, nil
		}
		if isSysMsg {
			delegate, err := pid.handleSystemMessage(msg)
			if err != nil {
				sysMsg, _ := msg.(*SystemMessage)
				return sysMsg, fmt.Errorf("handle system message: %w", err)
			}
			if !delegate {
				continue
			}
		}
		loop, err := a.receiveFunc(ctx, msg)
		if err != nil {
			sysMsg, _ := msg.(*SystemMessage)
			return sysMsg, fmt.Errorf("msg handler: %w", err)
		}
		if !loop {
			break
		}
	}

	return nil, nil
}

func (pid *PID) handleSystemMessage(message any) (delegate bool, err error) {
	msg, ok := message.(*SystemMessage)
	if !ok {
		return false, fmt.Errorf("non system message received to be handled by system message handler: %T", message)
	}

	rel := pid.relation(msg.Sender)
	if rel == relNone {
		// a message is received and no relation was found with the sender? then why did we get a sys message from them?
		// previously linked, died and restarted by supervisor without refreshing relations for this actor?
		// or used to be related but relation is deleted only on this actor's side? should we delegate?
		return false, nil
	}

	switch {
	case msg.Type == SystemMessageNormalExit && equalsAny(rel, relLinkedTrapExit, relLinked, relMonitored):
		pid.removeRelation(msg.Sender, rel)
		return true, nil
	case msg.Type == SystemMessageAbnormalExit && equalsAny(rel, relLinkedTrapExit, relMonitored):
		pid.removeRelation(msg.Sender, rel)
		return true, nil
	case msg.Type == SystemMessageAbnormalExit && rel == relLinked:
		pid.removeRelation(msg.Sender, rel)
		return false, fmt.Errorf("linked actor %q exited abnormally", msg.Sender.id)
	case rel == relMonitor && msg.Type == SystemMessageNormalExit || msg.Type == SystemMessageAbnormalExit:
		pid.removeRelation(msg.Sender, relMonitor)
		return false, nil
	case msg.Type == SystemMessageKill:
		panic(fmt.Sprintf("kill msg not implemented; received from %q", msg.Sender.id))
	case msg.Type == SystemMessageShutdown:
		panic(fmt.Sprintf("shutdown msg not implemented; received from %q", msg.Sender.id))
	default:
		return false, fmt.Errorf("system message with unknown type %q and/or relation %q received", msg.Type, rel)
	}
}

func (pid *PID) dispose(ctx context.Context, origin *SystemMessage, err error) {
	pid.r.Close()

	reason := any("normal exit")
	typ := SystemMessageNormalExit
	if err != nil {
		reason = err
		typ = SystemMessageAbnormalExit
	}

	if r := recover(); r != nil {
		reason = r
		typ = SystemMessageAbnormalExit
	}

	pid.notifyRelations(ctx, &SystemMessage{
		Sender: pid,
		Reason: reason,
		Type:   typ,
		Origin: origin,
	})
}

func (pid *PID) notifyRelations(ctx context.Context, msg *SystemMessage) {
	log.Printf("%q is getting disposed with system message: %+v\n", pid.id, msg)
	var ctxCancels []func()
	defer func() {
		for _, fn := range ctxCancels {
			fn()
		}
	}()

	// todo: notify concurrently via a worker pool?
	notify := func(who *PID) error {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		ctxCancels = append(ctxCancels, cancel)
		return sendSystemMessage(ctx, who, msg)
	}

	pid.relationsMu.Lock()
	defer pid.relationsMu.Unlock()

	for _, linked := range pid.links {
		err := notify(linked)
		if err != nil {
			log.Printf("%q could not notify linked actor %q of system message: %v: %+v\n", pid, linked, err, msg)
		}
	}
	for _, monitor := range pid.monitors {
		err := notify(monitor)
		if err != nil {
			log.Printf("%q could not notify monitor actor %q of system message: %v: %+v\n", pid, monitor, err, msg)
			continue
		}
	}
	for _, monitored := range pid.monitored {
		err := notify(monitored)
		if err != nil {
			log.Printf("%q could not notify monitored actor %q of system message: %v: %+v\n", pid, monitored, err, msg)
			continue
		}
	}
}

// equalsAny checks if `v` is equal to any of the provided values in the slice `s`.
func equalsAny[T comparable](v T, s ...T) bool {
	return slices.Contains(s, v)
}
