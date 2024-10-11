package supervision

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/supervision/strategy"
)

// ChildSpec defines the specifications for the supervisor children which can be either a child actor or
// a child supervisor.
type ChildSpec interface {
	// Name returns the name of this child spec used for registration.
	Name() string
	// StartLink starts the child linked to the supervisor.
	StartLink(ctx context.Context) (*goactor.PID, error)
	// RestartStrategy returns the child's restart strategy.
	RestartStrategy() RestartStrategy
	// validateChildSpec is used both for validating and preventing external packages from implementing this interface.
	validateChildSpec() error
}

// Strategy defines the methods required for representing a supervision Strategy.
type Strategy interface {
	Type() strategy.Type
	MaxRestarts() uint
	Period() time.Duration
}

// StartSupervisor starts a supervisor for the provided specs.
func StartSupervisor(ctx context.Context, strategy Strategy, specs ...ChildSpec) error {
	name := fmt.Sprintf("supervisor:parent:%s", uuid.NewString())
	supervisorSpec := NewSupervisorChildSpec(name, strategy, RestartNever, specs...)
	err := supervisorSpec.validateChildSpec()
	if err != nil {
		return fmt.Errorf("validate supervisor child specs: %w", err)
	}

	_, err = supervisorSpec.StartLink(ctx)
	if err != nil {
		return fmt.Errorf("startlink supervisor: %w", err)
	}

	return nil
}
