package supervision

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"slices"
	"time"

	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/internal/ringbuffer"
	"github.com/hedisam/goactor/supervision/strategy"
	"github.com/hedisam/goactor/sysmsg"
)

var (
	// ErrReachedMaxRestartIntensity is returned when too many restarts occur within the specified time window.
	ErrReachedMaxRestartIntensity = errors.New("shutdown: reached maximum restart intensity")
)

type activeChildInfo struct {
	pid           *goactor.PID
	spec          ChildSpec
	ctxCancelFunc context.CancelCauseFunc
}

// Supervisor is a supervisor Actor. It implements the goactor.Actor interface.
type Supervisor struct {
	self     *goactor.PID
	name     string
	strategy *strategy.Strategy
	// nameToChild holds the raw ChildSpec information which is immutable.
	nameToChild map[string]ChildSpec
	// children holds the list of ChildSpec in their original order, needed for RestForOne and OneForAllStrategy strategies.
	children []ChildSpec

	idToActiveChild map[string]*activeChildInfo
	nameToID        map[string]string
	restarts        *ringbuffer.RingBuffer[time.Time]
}

// Init initialises the supervisor by spawning all the children.
func (s *Supervisor) Init(ctx context.Context, self *goactor.PID) (err error) {
	goactor.GetLogger().Debug("Initialising supervisor", slog.String("name", s.name))
	s.self = self
	s.idToActiveChild = make(map[string]*activeChildInfo, len(s.nameToChild))
	s.nameToID = make(map[string]string, len(s.nameToChild))
	s.restarts = ringbuffer.New[time.Time](s.strategy.MaxRestarts())
	s.self.SetTrapExit(true)

	defer func() {
		if err != nil {
			s.cancelAll(err)
		}
	}()

	for name := range maps.Keys(s.nameToChild) {
		err = s.startChild(ctx, name)
		if err != nil {
			return fmt.Errorf("start child: %w", err)
		}
	}
	return nil
}

// Receive processes received system messages from its children.
func (s *Supervisor) Receive(ctx context.Context, message any) (loop bool, err error) {
	msg, ok := sysmsg.ToMessage(message)
	if !ok {
		return true, nil
	}

	switch msg.Type {
	case sysmsg.Signal:
		// TODO: support signal messages required for supervision trees
		return true, nil
	case sysmsg.Down:
		// a supervisor doesn't monitor actors, it gets linked to them while trapping exit messages.
		return true, nil
	case sysmsg.Exit:
		// this is what a supervisor handles re its child actors; fallthrough
	}

	childInfo, ok := s.idToActiveChild[msg.ProcessID]
	if !ok {
		goactor.GetLogger().Debug("Supervisor received system message from an unknown actor",
			slog.String("supervisor_name", s.name),
			slog.String("actor_id", msg.ProcessID),
		)
		return true, nil
	}

	s.unregisterChild(childInfo)
	if !s.shouldRestartChild(childInfo.spec, msg.Reason) {
		if len(s.idToActiveChild) == 0 {
			goactor.GetLogger().Debug("Stopping supervisor as no children are active anymore",
				slog.String("supervisor_name", s.name),
			)
			return false, nil
		}
	}

	if s.reachedMaxRestartIntensity() {
		err = ErrReachedMaxRestartIntensity
		s.cancelAll(err)
		return false, err
	}

	// TODO: should we record the exit message timestamp as the restart event timestamp?
	s.restarts.Put(time.Now().UTC())

	toShutdown, toRestart := s.strategy.Evaluate(childInfo.spec.Name(), mapSlice(s.children, func(spec ChildSpec) strategy.ChildInfo {
		_, alive := s.nameToID[spec.Name()]
		return strategy.ChildInfo{
			Name:      spec.Name(),
			Temporary: spec.RestartType() == Temporary,
			Stopped:   !alive,
		}
	}))

	for name := range slices.Values(toShutdown) {
		err = s.stopChild(name)
		if err != nil {
			return false, fmt.Errorf("stop child %q: %w", name, err)
		}
	}
	for name := range slices.Values(toRestart) {
		err = s.startChild(ctx, name)
		if err != nil {
			return false, fmt.Errorf("restart child %q: %w", name, err)
		}
	}

	return true, nil
}

// AfterFunc implements goactor.Actor.
func (s *Supervisor) AfterFunc() (timeout time.Duration, afterFunc goactor.AfterFunc) {
	return 0, func(ctx context.Context) error {
		return nil
	}
}

func (s *Supervisor) shouldRestartChild(spec ChildSpec, reason sysmsg.Reason) bool {
	switch spec.RestartType() {
	case Permanent:
		return true
	case Transient:
		return !errors.Is(reason, sysmsg.ReasonNormal) && !errors.Is(reason, sysmsg.ReasonShutdown)
	default:
		// temporary child
		return false
	}
}

func (s *Supervisor) reachedMaxRestartIntensity() bool {
	oldestRestart, _ := s.restarts.Get()
	switch {
	case s.strategy.MaxRestarts() == 0:
		return true
	case time.Since(oldestRestart) <= s.strategy.Period() && s.restarts.Size() == s.strategy.MaxRestarts():
		return true
	default:
		return false
	}
}

func (s *Supervisor) startChild(ctx context.Context, name string) error {
	child, ok := s.nameToChild[name]
	if !ok {
		return fmt.Errorf("no child spec found with the given name %q", name)
	}

	ctx, cancel := context.WithCancelCause(ctx)
	pid, err := child.StartLink(ctx)
	if err != nil {
		return fmt.Errorf("startlink child %q: %w", child.Name(), err)
	}
	info := &activeChildInfo{
		pid:           pid,
		spec:          child,
		ctxCancelFunc: cancel,
	}
	defer func() {
		if err != nil {
			cancel(err)
			s.unregisterChild(info)
		}
	}()
	err = s.registerChild(info)
	if err != nil {
		return fmt.Errorf("mark child as started: %w", err)
	}
	return nil
}

func (s *Supervisor) registerChild(info *activeChildInfo) error {
	s.idToActiveChild[info.pid.ID()] = info
	s.nameToID[info.spec.Name()] = info.pid.ID()
	err := s.self.Link(info.pid)
	if err != nil {
		return fmt.Errorf("link to child: %w", err)
	}
	err = goactor.Register(info.spec.Name(), info.pid)
	if err != nil {
		return fmt.Errorf("could not register child actor %q: %w", info.spec.Name(), err)
	}
	return nil
}

func (s *Supervisor) unregisterChild(activeChild *activeChildInfo) {
	goactor.Unregister(activeChild.spec.Name())
	_ = s.self.Unlink(activeChild.pid)
	delete(s.idToActiveChild, activeChild.pid.ID())
	delete(s.nameToID, activeChild.spec.Name())
}

func (s *Supervisor) cancelAll(err error) {
	for info := range maps.Values(s.idToActiveChild) {
		s.unregisterChild(info)
		info.ctxCancelFunc(err)
	}
}

func (s *Supervisor) stopChild(name string) error {
	id, ok := s.nameToID[name]
	if !ok {
		return fmt.Errorf("could not find child ID with name %q", name)
	}
	info, ok := s.idToActiveChild[id]
	if !ok {
		return fmt.Errorf("could not find child info with name %q and ID %q", name, id)
	}
	s.unregisterChild(info)
	// TODO: attempt to shutdown then :brutal_kill with context cancellation if had to; requires child shutdown behaviour to be implemented
	info.ctxCancelFunc(sysmsg.ReasonKill)
	return nil
}

func (s *Supervisor) strategyChildrenInfo() []strategy.ChildInfo {
	children := make([]strategy.ChildInfo, 0, len(s.children))
	for child := range slices.Values(s.children) {
		_, alive := s.nameToID[child.Name()]
		children = append(children, strategy.ChildInfo{
			Name:      child.Name(),
			Temporary: child.RestartType() == Temporary,
			Stopped:   !alive,
		})
	}
	return children
}

func mapSlice[T any, D any](s []T, fn func(T) D) []D {
	mapped := make([]D, 0, len(s))
	for t := range slices.Values(s) {
		mapped = append(mapped, fn(t))
	}
	return mapped
}
