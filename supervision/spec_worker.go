package supervision

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hedisam/goactor"
)

var _ ChildSpec = &WorkerSpec{}

// ActorFactory is what expected by a WorkerSpec. It's necessary to have an ActorFactory that can create a fresh
// instance of the Actor whenever the supervisor restarts the processes.
type ActorFactory func() goactor.Actor

// WorkerSpec holds the configuration for spawning a worker child.
type WorkerSpec struct {
	name         string
	actorFactory ActorFactory
	restartType  RestartType
}

// NewWorkerSpec returns a new goactor.Actor child spec.
func NewWorkerSpec(name string, restartType RestartType, af ActorFactory) *WorkerSpec {
	return &WorkerSpec{
		name:         strings.TrimSpace(name),
		actorFactory: af,
		restartType:  restartType,
	}
}

// Name returns the assigned name to this child actor.
func (s *WorkerSpec) Name() string {
	return s.name
}

// StartLink spawns the child actor linked to the supervisor.
func (s *WorkerSpec) StartLink(ctx context.Context) (*goactor.PID, error) {
	pid, err := goactor.Spawn(ctx, s.actorFactory())
	if err != nil {
		return nil, fmt.Errorf("spawn actor child: %w", err)
	}
	return pid, nil
}

// RestartType returns the restart strategy set for this child.
func (s *WorkerSpec) RestartType() RestartType {
	return s.restartType
}

func (s *WorkerSpec) validate() error {
	if s.name == "" {
		return errors.New("child spec name cannot be empty")
	}
	if s.actorFactory == nil {
		return fmt.Errorf("nil actor provided for child worker %q", s.name)
	}

	err := validateRestartType(s.restartType)
	if err != nil {
		return fmt.Errorf("validate worker restart type: %w", err)
	}

	return nil
}
