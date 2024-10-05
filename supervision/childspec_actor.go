package supervision

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hedisam/goactor"
)

// ActorChildSpec holds the configuration for spawning a child actor.
type ActorChildSpec struct {
	name               string
	actor              goactor.Actor
	restartingStrategy RestartStrategy
}

// NewActorChildSpec returns a new goactor.Actor child spec.
func NewActorChildSpec(name string, restartStrategy RestartStrategy, actor goactor.Actor) *ActorChildSpec {
	return &ActorChildSpec{
		name:               strings.TrimSpace(name),
		actor:              actor,
		restartingStrategy: restartStrategy,
	}
}

// Name returns the assigned name to this child actor.
func (s *ActorChildSpec) Name() string {
	return s.name
}

// StartLink spawns the child actor linked to the supervisor.
func (s *ActorChildSpec) StartLink(ctx context.Context) (*goactor.PID, error) {
	pid, err := goactor.Spawn(ctx, s.actor)
	if err != nil {
		return nil, fmt.Errorf("spawn actor child: %w", err)
	}
	return pid, nil
}

func (s *ActorChildSpec) validateChildSpec() error {
	if s.name == "" {
		return errors.New("child spec name cannot be empty")
	}

	if s.actor == nil {
		return errors.New("child actor cannot be nil")
	}

	err := validateRestartStrategy(s.restartingStrategy)
	if err != nil {
		return fmt.Errorf("validate child spec restart strategy: %w", err)
	}
	return nil
}
