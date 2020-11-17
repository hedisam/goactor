package intlspec

import (
	"fmt"
	"github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/supervisor/option"
	"strings"
)

type ChildType uint8

type Spec interface {
	StartLink() (*pid.PID, error)
	SupervisorOptions() *option.Options
	RestartWhen() int
	Name() string
}

var DefaultSupervisorStartLink func(option.Options, ...Spec) (*pid.PID, error)

func SetDefaultSupStartLink(sl func(option.Options, ...Spec) (*pid.PID, error)) {
	DefaultSupervisorStartLink = sl
}

func SpecsToMap(specs ...Spec) (map[string]Spec, error) {
	if len(specs) == 0 {
		return nil, fmt.Errorf("specs validator: specs list is empty")
	}

	specsMap := make(map[string]Spec)

	for _, spec := range specs {
		if err := Validate(spec); err != nil {
			return nil, err
		}
		if _, duplicate := specsMap[spec.Name()]; duplicate {
			return nil, fmt.Errorf("specs validator: duplicate childspec id: %s", spec.Name())
		}
		specsMap[spec.Name()] = spec
	}

	return specsMap, nil
}

func Validate(spec Spec) error {
	if strings.TrimSpace(spec.Name()) == "" {
		return fmt.Errorf("childspec validator: childspec's id/name could not be empty")
	} else if spec.RestartWhen() < 0 && spec.RestartWhen() > 2 {
		return fmt.Errorf("invalid childspec's restart value: %v", spec.RestartWhen())
	}
	return nil
}
