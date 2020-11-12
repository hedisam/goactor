package supervisor

import (
	"github.com/google/uuid"
	p "github.com/hedisam/goactor/pid"
	"strings"
)

type SupSpec struct {
	Id            string
	Children      []Spec
	StartFn       StartLink
	WhenToRestart int
	SupOptions    Options
}

func (s SupSpec) StartLink() (*p.PID, error) {
	ref, err := Start(s.SupOptions, s.Children...)
	if err != nil {
		return nil, err
	}
	return ref.pid, err
}

func (s SupSpec) SupervisorOptions() *Options {
	return &s.SupOptions
}

func (s SupSpec) RestartWhen() int {
	return s.WhenToRestart
}

func (s SupSpec) Name() string {
	return s.Id
}

func (s SupSpec) ChildType() ChildType {
	return ChildSupervisor
}

func NewSupervisorSpec(name string, startLink StartLink, restartWhen int, supOptions Options, childSpecs ...Spec) SupSpec {
	if strings.TrimSpace(name) == "" {
		name = uuid.New().String()
	}
	s := SupSpec{
		Id:            name,
		Children:      childSpecs,
		StartFn:       startLink,
		WhenToRestart: restartWhen,
		SupOptions:    supOptions,
	}
	return s
}
