package supervision

import (
	"context"
	"fmt"
	"slices"

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
	// RestartType returns the child's restart strategy.
	RestartType() RestartType
	// validate is used both for validating and preventing external packages from implementing this interface.
	validate() error
}

// Start starts a supervisor for the provided specs.
func Start(ctx context.Context, strategy *strategy.Strategy, specs ...ChildSpec) error {
	err := ensureUniqueNamesInSupervisionTree(specs)
	if err != nil {
		return fmt.Errorf("ensure unique names in supervision tree: %w", err)
	}

	name := fmt.Sprintf("supervisor:parent:%s", uuid.NewString())
	supervisorSpec := NewSupervisorSpec(name, strategy, Temporary, specs...)
	err = supervisorSpec.validate()
	if err != nil {
		return fmt.Errorf("validate supervisor child specs: %w", err)
	}

	_, err = supervisorSpec.StartLink(ctx)
	if err != nil {
		return fmt.Errorf("startlink supervisor: %w", err)
	}

	return nil
}

func ensureUniqueNamesInSupervisionTree(specs []ChildSpec) error {
	allNames := make([]string, 0, len(specs))
	for spec := range slices.Values(specs) {
		names, err := specNames(spec)
		if err != nil {
			return fmt.Errorf("could not get spec names for child %q: %w", spec.Name(), err)
		}
		allNames = append(allNames, names...)
	}
	allNames = slices.Compact(allNames)
	slices.Sort(allNames)
	if len(slices.Compact(allNames)) != len(allNames) {
		return fmt.Errorf("supervision tree cannot have children with duplicate names: %v", allNames)
	}

	return nil
}

func specNames(spec ChildSpec) ([]string, error) {
	switch s := spec.(type) {
	case *WorkerSpec:
		return []string{s.Name()}, nil
	case *SupervisorSpec:
		names := make([]string, len(s.children)+1)
		names = append(names, s.Name())
		for child := range slices.Values(s.children) {
			childNames, err := specNames(child)
			if err != nil {
				return nil, fmt.Errorf("supervisor spec names: %w", err)
			}
			names = append(names, childNames...)
		}
		return names, nil
	default:
		return nil, fmt.Errorf("unknown chid spec type with name %q: %T", spec.Name(), spec)
	}
}
