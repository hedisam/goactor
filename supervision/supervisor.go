package supervision

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"time"

	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/internal/ringbuffer"
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
	strategy *Strategy
	// nameToChild holds the raw ChildSpec information which is immutable.
	nameToChild map[string]ChildSpec

	idToActiveChild map[string]*activeChildInfo
	restarts        *ringbuffer.RingBuffer[time.Time]
}

// Init initialises the supervisor by spawning all the children.
func (s *Supervisor) Init(ctx context.Context, self *goactor.PID) (err error) {
	goactor.GetLogger().Debug("Initialising supervisor", slog.String("name", s.name))
	s.self = self
	s.idToActiveChild = make(map[string]*activeChildInfo, len(s.nameToChild))
	s.restarts = ringbuffer.New[time.Time](s.strategy.maxRestarts)

	defer func() {
		if err != nil {
			s.cancelAll(err)
		}
	}()

	for child := range maps.Values(s.nameToChild) {
		err = s.startChild(ctx, child)
		if err != nil {
			return fmt.Errorf("start child: %w", err)
		}
	}
	return nil
}

// Receive processes received system messages from its children.
func (s *Supervisor) Receive(ctx context.Context, message any) (loop bool, err error) {
	msg, ok := sysmsg.ToSystemMessage(message)
	if !ok {
		return true, nil
	}

	info, ok := s.idToActiveChild[msg.SenderID]
	if !ok {
		goactor.GetLogger().Debug("Supervisor received system message from an unknown actor",
			slog.String("supervisor_name", s.name),
			slog.String("actor_id", msg.SenderID),
		)
		return true, nil
	}

	s.unregisterChild(info)
	if !s.shouldRestartChild(info.spec, msg.Type) {
		if len(s.idToActiveChild) == 0 {
			goactor.GetLogger().Debug("Shutting down supervisor as no children are active anymore",
				slog.String("supervisor_name", s.name),
			)
			return false, nil
		}
	}

	if !s.canRestartChild() {
		err = ErrReachedMaxRestartIntensity
		s.cancelAll(err)
		return false, err
	}

	switch s.strategy.typ {
	case StrategyOneForOne:
		err = s.restartChild(ctx, info.spec)
		if err != nil {
			return false, fmt.Errorf("restart child: %w", err)
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

func (s *Supervisor) shouldRestartChild(spec ChildSpec, msgType sysmsg.MessageType) bool {
	switch spec.RestartStrategy() {
	case RestartNever:
		return false
	case RestartTransient:
		return msgType != sysmsg.NormalExit && msgType != sysmsg.Shutdown
	default:
		return true
	}
}

func (s *Supervisor) restartChild(ctx context.Context, child ChildSpec) error {
	s.restarts.Put(time.Now().UTC())
	err := s.startChild(ctx, child)
	if err != nil {
		return fmt.Errorf("start child: %w", err)
	}
	return nil
}

func (s *Supervisor) canRestartChild() bool {
	if s.strategy.maxRestarts == 0 {
		return false
	}

	firstRestart, ok := s.restarts.Get()
	if !ok {
		return true
	}
	if time.Since(firstRestart) <= s.strategy.period && s.restarts.Size() == s.strategy.maxRestarts {
		return false
	}

	return true
}

func (s *Supervisor) startChild(ctx context.Context, child ChildSpec) error {
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
	s.self.Link(info.pid, true)
	err := goactor.Register(info.spec.Name(), info.pid)
	if err != nil {
		return fmt.Errorf("could not register child actor %q: %w", info.spec.Name(), err)
	}
	return nil
}

func (s *Supervisor) unregisterChild(activeChild *activeChildInfo) {
	goactor.Unregister(activeChild.spec.Name())
	s.self.Unlink(activeChild.pid)
	delete(s.idToActiveChild, activeChild.pid.ID())
}

func (s *Supervisor) cancelAll(err error) {
	for info := range maps.Values(s.idToActiveChild) {
		s.unregisterChild(info)
		info.ctxCancelFunc(err)
	}
}
