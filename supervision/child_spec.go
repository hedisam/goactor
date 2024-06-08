package supervision

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/hedisam/goactor/sysmsg"
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

type PID interface {
	PushSystemMessage(ctx context.Context, msg *sysmsg.Message) error
}

// ChildSpec defines the specifications for the supervisor children which can be either a child actor or
// a child supervisor.
type ChildSpec interface {
	Name() string
	StartLink(ctx context.Context) PID
	RestartStrategy() RestartStrategy
}

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
