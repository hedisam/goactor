package spec

import (
	"github.com/google/uuid"
	"github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/supervisor/internal/intlspec"
	"github.com/hedisam/goactor/supervisor/option"
	"strings"
)

type SupervisorSpec struct {
	Id            string
	Children      []intlspec.Spec
	StartFn       StartLink
	WhenToRestart int
	SupOptions    option.Options
}

func (s SupervisorSpec) StartLink() (*pid.PID, error) {
	return intlspec.DefaultSupervisorStartLink(s.SupOptions, s.Children...)
}

func (s SupervisorSpec) SupervisorOptions() *option.Options {
	return &s.SupOptions
}

func (s SupervisorSpec) RestartWhen() int {
	return s.WhenToRestart
}

func (s SupervisorSpec) Name() string {
	return s.Id
}

func (s SupervisorSpec) SetStartLinkFunc(fn StartLink) SupervisorSpec {
	s.StartFn = fn
	return s
}

func NewSupervisorSpec(name string, restartWhen int, supOptions option.Options, childSpecs ...intlspec.Spec) SupervisorSpec {
	if strings.TrimSpace(name) == "" {
		name = uuid.New().String()
	}
	s := SupervisorSpec{
		Id:            name,
		Children:      childSpecs,
		WhenToRestart: restartWhen,
		SupOptions:    supOptions,
	}
	return s
}
