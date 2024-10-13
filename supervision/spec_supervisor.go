package supervision

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/supervision/strategy"
)

var _ ChildSpec = &SupervisorSpec{}

// SupervisorSpec holds the configuration for spawning a supervisor child.
type SupervisorSpec struct {
	name        string
	children    []ChildSpec
	restartType RestartType
	strategy    *strategy.Strategy
}

// NewSupervisorSpec returns a new Supervisor child spec.
func NewSupervisorSpec(name string, strategy *strategy.Strategy, restartType RestartType, children ...ChildSpec) *SupervisorSpec {
	return &SupervisorSpec{
		name:        strings.TrimSpace(name),
		children:    children,
		restartType: restartType,
		strategy:    strategy,
	}
}

// Name returns this supervisor's name.
func (s *SupervisorSpec) Name() string {
	return s.name
}

// StartLink starts the supervisor child linked to the parent supervisor.
func (s *SupervisorSpec) StartLink(ctx context.Context) (*goactor.PID, error) {
	nameToChild := make(map[string]ChildSpec, len(s.children))
	for spec := range slices.Values(s.children) {
		nameToChild[spec.Name()] = spec
	}

	supervisor := &Supervisor{
		name:        s.name,
		strategy:    s.strategy,
		nameToChild: nameToChild,
		children:    s.children,
	}

	goactor.GetLogger().Debug("Starting supervisor", slog.String("supervisor_name", s.name))
	pid, err := goactor.Spawn(ctx, supervisor)
	if err != nil {
		return nil, fmt.Errorf("spawn supervisor actor %q: %w", s.name, err)
	}

	return pid, nil
}

// RestartType returns the restart strategy set for this child.
func (s *SupervisorSpec) RestartType() RestartType {
	return s.restartType
}

func (s *SupervisorSpec) validate() error {
	if s.name == "" {
		return fmt.Errorf("supervisor spec name cannot be empty")
	}
	if len(s.children) == 0 {
		return fmt.Errorf("no child spec provided for supervisor %q", s.name)
	}

	err := validateRestartType(s.restartType)
	if err != nil {
		return fmt.Errorf("validate supevisor restart type: %w", err)
	}

	for child := range slices.Values(s.children) {
		err = child.validate()
		if err != nil {
			return fmt.Errorf("validate child %q: %w", child.Name(), err)
		}
	}

	return nil
}
