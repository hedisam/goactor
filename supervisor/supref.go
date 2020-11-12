package supervisor

import (
	"fmt"
	p "github.com/hedisam/goactor/pid"
)

type SupRef struct {
	pid *p.PID
}

func ToSupervisorRef(pid *p.PID) (*SupRef, error) {
	if pid.IsSupervisor() {
		return &SupRef{pid: pid}, nil
	}
	return nil, fmt.Errorf("can not convert a worker pid to supervisor reference")
}
