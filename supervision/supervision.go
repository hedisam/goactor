package supervision

import (
	"context"
	"errors"
	"fmt"
)

type Supervisor struct {
	nameToChild map[string]ChildSpec
}

// StartSupervisor starts a supervisor for the provided specs.
func StartSupervisor(ctx context.Context, specs ...ChildSpec) error {
	if len(specs) == 0 {
		return errors.New("no child spec provided")
	}

	nameToChild := make(map[string]ChildSpec, len(specs))
	for _, spec := range specs {
		err := ValidateChildSpec(spec)
		if err != nil {
			return fmt.Errorf("validate child spec: %w", err)
		}
		_, ok := nameToChild[spec.Name()]
		if ok {
			return fmt.Errorf("cannot have dupliate child spec names: %q", spec.Name())
		}
		nameToChild[spec.Name()] = spec
	}

	supervisor := &Supervisor{
		nameToChild: nameToChild,
	}
	supervisor.start()

	return nil
}

func (s *Supervisor) start() {

}
