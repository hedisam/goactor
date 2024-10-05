package supervision

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hedisam/goactor"
)

// ChildSpec defines the specifications for the supervisor children which can be either a child actor or
// a child supervisor.
type ChildSpec interface {
	Name() string
	StartLink(ctx context.Context) *goactor.PID
	RestartStrategy() RestartStrategy
}

// StartSupervisor starts a supervisor for the provided specs.
func StartSupervisor(ctx context.Context, strategy *Strategy, specs ...ChildSpec) error {
	if len(specs) == 0 {
		return errors.New("no child spec provided")
	}

	nameToChild := make(map[string]ChildSpec, len(specs))
	for _, spec := range specs {
		err := validateChildSpec(spec)
		if err != nil {
			return fmt.Errorf("validate child spec: %w", err)
		}
		_, ok := nameToChild[spec.Name()]
		if ok {
			return fmt.Errorf("cannot have dupliate child spec names: %q", spec.Name())
		}
		nameToChild[spec.Name()] = spec
	}

	err := strategy.Validate()
	if err != nil {
		return fmt.Errorf("validate supervision strategy: %w", err)
	}

	s := &Supervisor{
		strategy:    strategy,
		nameToChild: nameToChild,
	}
	log.Println("Starting supervisor...")
	_ = goactor.Spawn(
		ctx,
		s.Receive,
		goactor.WithInitFunc(s.Init),
	)
	if s.initErr != nil {
		return fmt.Errorf("init supervisor actor: %w", err)
	}

	return nil
}
