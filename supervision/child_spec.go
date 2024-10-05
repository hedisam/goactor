package supervision

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/hedisam/goactor"
)

const (
	// RestartAlways means the child actor is always restarted (the default).
	RestartAlways RestartStrategy = ":permanent"
	// RestartTransient means the child actor is restarted only if it terminates abnormally (with an exit reason
	// other than :normal and :shutdown)
	RestartTransient RestartStrategy = ":transient"
	// RestartNever means the child actor is never restarted, regardless of its termination reason.
	RestartNever RestartStrategy = ":temporary"
)

// RestartStrategy determines when to restart a child actor if it terminates.
type RestartStrategy string

func validateRestartStrategy(s RestartStrategy) error {
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

func validateChildSpec(spec ChildSpec) error {
	if spec.Name() == "" {
		return errors.New("supervision child spec name cannot be empty")
	}

	err := validateRestartStrategy(spec.RestartStrategy())
	if err != nil {
		return fmt.Errorf("validate restart strategy: %w", err)
	}
	return nil
}

// ActorChildSpec holds the configuration for spawning a child actor.
type ActorChildSpec struct {
	name               string
	receiverFunc       goactor.ReceiveFunc
	opts               []goactor.ActorOption
	restartingStrategy RestartStrategy
}

func NewActorChildSpec(name string, restartStrategy RestartStrategy, fn goactor.ReceiveFunc, opts ...goactor.ActorOption) *ActorChildSpec {
	return &ActorChildSpec{
		name:               name,
		receiverFunc:       fn,
		opts:               opts,
		restartingStrategy: restartStrategy,
	}
}

// Name returns the name of the actor.
func (s *ActorChildSpec) Name() string {
	return s.name
}

// StartLink spawns the child actor.
func (s *ActorChildSpec) StartLink(ctx context.Context) *goactor.PID {
	return goactor.Spawn(ctx, s.receiverFunc, s.opts...)
}

// RestartStrategy returns the restart strategy which determines when to restart this child actor if it terminates.
func (s *ActorChildSpec) RestartStrategy() RestartStrategy {
	return s.restartingStrategy
}
