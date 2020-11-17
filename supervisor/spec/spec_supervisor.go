package spec

import (
	"github.com/google/uuid"
	p "github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/supervisor/option"
	"strings"
)

type SupervisorSpec struct {
	Id            string
	Children      []Spec
	StartFn       StartLink
	WhenToRestart int
	SupOptions    option.Options
}

func (s SupervisorSpec) StartLink() (*p.PID, error) {
	return supStartFunc(s.SupOptions, s.Children...)
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

func (s SupervisorSpec) ChildType() ChildType {
	return ChildSupervisor
}

func NewSupervisorSpec(name string, startLink StartLink, restartWhen int, supOptions option.Options, childSpecs ...Spec) SupervisorSpec {
	if strings.TrimSpace(name) == "" {
		name = uuid.New().String()
	}
	s := SupervisorSpec{
		Id:            name,
		Children:      childSpecs,
		StartFn:       startLink,
		WhenToRestart: restartWhen,
		SupOptions:    supOptions,
	}
	return s
}
