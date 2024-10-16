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

// supervisor is a supervisor Actor. It implements the goactor.Actor interface.
type supervisor struct {
	self     *goactor.PID
	name     string
	strategy *strategy.Strategy
	// nameToChild holds the raw ChildSpec information which is immutable.
	nameToChild map[string]ChildSpec
	// children holds the list of ChildSpec in their original order, needed for RestForOne and OneForAllStrategy strategies.
	children []ChildSpec

	idToActiveChild map[string]*activeChildInfo
	nameToActivePID map[string]string
	restarts        *ringbuffer.RingBuffer[time.Time]
}

// Init initialises the supervisor by spawning all the children.
func (s *supervisor) Init(ctx context.Context, self *goactor.PID) (err error) {
	goactor.GetLogger().Debug("Initialising supervisor", slog.String("name", s.name))
	s.self = self
	s.idToActiveChild = make(map[string]*activeChildInfo, len(s.nameToChild))
	s.nameToActivePID = make(map[string]string, len(s.nameToChild))
	s.restarts = ringbuffer.New[time.Time](s.strategy.MaxRestarts())
	err = goactor.SetTrapExit(true)
	if err != nil {
		return fmt.Errorf("set supervisor's trap exit: %w", err)
	}

	defer func() {
		if err != nil {
			for child := range maps.Values(s.idToActiveChild) {
				_ = s.stopChild(child.spec.Name())
			}
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
func (s *supervisor) Receive(ctx context.Context, message any) (loop bool, err error) {
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

	defer func() {
		if err == nil && len(s.idToActiveChild) == 0 {
			goactor.GetLogger().Debug("Stopping supervisor as no children are active anymore",
				slog.String("supervisor_name", s.name),
			)
			loop = false
			return
		}
		if err != nil {
			for child := range maps.Values(s.idToActiveChild) {
				_ = s.stopChild(child.spec.Name())
			}
		}
	}()

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
		return true, nil
	}

	if s.reachedMaxRestartIntensity() {
		return false, ErrReachedMaxRestartIntensity // todo: should explicitly shutdown other children with :shutdown
	}

	// TODO: should we record the exit message timestamp as the restart event timestamp?
	s.restarts.Put(time.Now().UTC())

	toShutdown, toRestart := s.strategy.Evaluate(childInfo.spec.Name(), mapSlice(s.children, func(spec ChildSpec) strategy.ChildInfo {
		_, alive := s.nameToActivePID[spec.Name()]
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
func (s *supervisor) AfterFunc() (timeout time.Duration, afterFunc goactor.AfterFunc) {
	return 0, func(ctx context.Context) error {
		return nil
	}
}

func (s *supervisor) shouldRestartChild(spec ChildSpec, reason sysmsg.Reason) bool {
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

func (s *supervisor) reachedMaxRestartIntensity() bool {
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

func (s *supervisor) startChild(ctx context.Context, name string) error {
	child, ok := s.nameToChild[name]
	if !ok {
		return fmt.Errorf("no child spec found with the given name %q", name)
	}

	ctx, cancel := context.WithCancelCause(ctx)
	pid, err := child.StartLink(ctx)
	if err != nil {
		return fmt.Errorf("startlink child %q: %w", child.Name(), err)
	}

	err = s.registerChild(&activeChildInfo{
		pid:           pid,
		spec:          child,
		ctxCancelFunc: cancel,
	})
	if err != nil {
		_ = s.stopChild(name)
		return fmt.Errorf("resgiter child: %w", err)
	}
	return nil
}

func (s *supervisor) registerChild(info *activeChildInfo) error {
	s.nameToActivePID[info.spec.Name()] = info.pid.ID()
	s.idToActiveChild[info.pid.ID()] = info
	err := goactor.Link(info.pid)
	if err != nil {
		return fmt.Errorf("link to child: %w", err)
	}
	err = goactor.Register(info.spec.Name(), info.pid)
	if err != nil {
		return fmt.Errorf("could not register child actor %q: %w", info.spec.Name(), err)
	}
	return nil
}

func (s *supervisor) unregisterChild(activeChild *activeChildInfo) {
	goactor.Unregister(activeChild.spec.Name())
	_ = goactor.Unlink(activeChild.pid)
	delete(s.idToActiveChild, activeChild.pid.ID())
	delete(s.nameToActivePID, activeChild.spec.Name())
}

func (s *supervisor) stopChild(name string) error {
	id, ok := s.nameToActivePID[name]
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

func mapSlice[T any, D any](s []T, fn func(T) D) []D {
	mapped := make([]D, 0, len(s))
	for t := range slices.Values(s) {
		mapped = append(mapped, fn(t))
	}
	return mapped
}
