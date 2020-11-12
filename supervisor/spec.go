package supervisor

import (
	"fmt"
	p "github.com/hedisam/goactor/pid"
	"strings"
)

type StartLink func() (*p.PID, error)

type Spec interface {
	StartLink() (*p.PID, error)
	SupervisorOptions() *Options
	RestartWhen() int
	Name() string
	ChildType() ChildType
}

func specsToMap(specs ...Spec) (map[string]Spec, error) {
	if len(specs) == 0 {
		return nil, fmt.Errorf("specs validator: specs list is empty")
	}

	specsMap := make(map[string]Spec)

	for _, spec := range specs {
		if err := validateSpec(spec); err != nil {
			return nil, err
		}
		if _, duplicate := specsMap[spec.Name()]; duplicate {
			return nil, fmt.Errorf("specs validator: duplicate childspec id: %s", spec.Name())
		}
		specsMap[spec.Name()] = spec
	}

	return specsMap, nil
}

func validateSpec(spec Spec) error {
	if strings.TrimSpace(spec.Name()) == "" {
		return fmt.Errorf("childspec validator: childspec's id/name could not be empty")
	} else if spec.RestartWhen() < RestartAlways && spec.RestartWhen() > RestartNever {
		return fmt.Errorf("invalid childspec's restart value: %v", spec.RestartWhen())
	}
	return nil
}

type ChildType uint8

const (
	ChildWorker ChildType = iota
	ChildSupervisor
)

const (
	RestartAlways = iota
	RestartTransient
	RestartNever
)
