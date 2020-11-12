package pid

import "github.com/hedisam/goactor/internal/intlpid"

type PID struct {
	iPID intlpid.InternalPID
}

func ToPID(iPID intlpid.InternalPID) *PID {
	return &PID{iPID: iPID}
}

func (pid *PID) ID() string {
	return pid.iPID.ID()
}

func (pid *PID) IsSupervisor() bool {
	return pid.iPID.IsSupervisor()
}

func (pid *PID) InternalPID() intlpid.InternalPID {
	return pid.iPID
}
