package supervision

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/hedisam/goactor"
)

// RestartStrategy determines when to restart a child actor if it terminates.
type RestartStrategy string

// Validate validates this RestartStrategy.
func (s RestartStrategy) Validate() error {
	validStrategies := []string{
		string(RestartAlways),
		string(RestartTransient),
		string(RestartNever),
	}
	if !slices.Contains(validStrategies, string(s)) {
		return fmt.Errorf("invalid child restart strategy %q, valid restart strategies are: [%s]", s, validStrategies)
	}
	return nil
}

const (
	// RestartAlways means the child actor is always restarted (the default).
	RestartAlways RestartStrategy = ":permanent"
	// RestartTransient means the child actor is restarted only if it terminates abnormally (with an exit reason
	// other than :normal and :shutdown)
	RestartTransient RestartStrategy = ":transient"
	// RestartNever means the child actor is never restarted, regardless of its termination reason.
	RestartNever RestartStrategy = ":temporary"
)

// ValidateChildSpec validates the provided child spec.
func ValidateChildSpec(spec ChildSpec) error {
	if spec.Name() == "" {
		return errors.New("supervision child spec name cannot be empty")
	}

	err := spec.RestartStrategy().Validate()
	if err != nil {
		return fmt.Errorf("validate restart strategy: %w", err)
	}
	return nil
}

// ActorChildSpec holds the configuration for spawning a child actor.
type ActorChildSpec struct {
	ActorName          string
	ReceiverFunc       goactor.ReceiveFunc
	ActorOpts          []goactor.ActorOption
	RestartingStrategy RestartStrategy
}

func NewActorChildSpec(name string, restartStrategy RestartStrategy, fn goactor.ReceiveFunc, opts ...goactor.ActorOption) *ActorChildSpec {
	return &ActorChildSpec{
		ActorName:          name,
		ReceiverFunc:       fn,
		ActorOpts:          opts,
		RestartingStrategy: restartStrategy,
	}
}

// Name returns the name of the actor.
func (s *ActorChildSpec) Name() string {
	return s.ActorName
}

// StartLink spawns the child actor.
func (s *ActorChildSpec) StartLink(ctx context.Context) *goactor.PID {
	return goactor.Spawn(ctx, s.ReceiverFunc, s.ActorOpts...)
}

// RestartStrategy returns the restart strategy which determines when to restart this child actor if it terminates.
func (s *ActorChildSpec) RestartStrategy() RestartStrategy {
	return s.RestartingStrategy
}
