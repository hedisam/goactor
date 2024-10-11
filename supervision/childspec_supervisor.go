package supervision

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/hedisam/goactor"
)

var _ ChildSpec = &SupervisorChildSpec{}

// SupervisorChildSpec holds the configuration for spawning a supervisor child spec.
type SupervisorChildSpec struct {
	name            string
	children        []ChildSpec
	restartStrategy RestartStrategy
	strategy        *Strategy
}

// NewSupervisorChildSpec returns a new Supervisor child spec.
func NewSupervisorChildSpec(name string, strategy *Strategy, restartStrategy RestartStrategy, children ...ChildSpec) *SupervisorChildSpec {
	return &SupervisorChildSpec{
		name:            strings.TrimSpace(name),
		children:        children,
		restartStrategy: restartStrategy,
		strategy:        strategy,
	}
}

// Name returns this supervisor's name.
func (s *SupervisorChildSpec) Name() string {
	return s.name
}

// StartLink starts the supervisor child linked to the parent supervisor.
func (s *SupervisorChildSpec) StartLink(ctx context.Context) (*goactor.PID, error) {
	nameToChild := make(map[string]ChildSpec, len(s.children))
	for spec := range slices.Values(s.children) {
		nameToChild[spec.Name()] = spec
	}

	supervisor := &Supervisor{
		name:        s.name,
		strategy:    s.strategy,
		nameToChild: nameToChild,
	}

	goactor.GetLogger().Debug("Starting supervisor", slog.String("supervisor_name", s.name))
	pid, err := goactor.Spawn(ctx, supervisor)
	if err != nil {
		return nil, fmt.Errorf("spawn supervisor actor %q: %w", s.name, err)
	}

	return pid, nil
}

// RestartStrategy returns the restart strategy set for this child.
func (s *SupervisorChildSpec) RestartStrategy() RestartStrategy {
	return s.restartStrategy
}

func (s *SupervisorChildSpec) validateChildSpec() error {
	if s.name == "" {
		return fmt.Errorf("supervisor spec name cannot be empty")
	}
	if len(s.children) == 0 {
		return fmt.Errorf("no child spec provided for supervisor %q", s.name)
	}
	err := validateSupervisionStrategy(s.strategy)
	if err != nil {
		return fmt.Errorf("invalid supervision strategy: %w", err)
	}

	seen := make(map[string]struct{}, len(s.children))
	for child := range slices.Values(s.children) {
		if _, ok := seen[child.Name()]; ok {
			return fmt.Errorf("cannot use duplicate names for child specs: %q", child.Name())
		}
		err := child.validateChildSpec()
		if err != nil {
			return fmt.Errorf("validate child %q: %w", child.Name(), err)
		}
		seen[child.Name()] = struct{}{}
	}
	return nil
}
