package sysmsg

import (
	"github.com/hedisam/goactor/internal/intlpid"
)

type KillExit struct {
	from   intlpid.InternalPID
	reason interface{}
	origin SystemMessage
}

func NewKillMessage(from intlpid.InternalPID, reason interface{}, origin SystemMessage) KillExit {
	return KillExit{
		from:   from,
		reason: reason,
		origin: origin,
	}
}

func (m KillExit) Sender() intlpid.InternalPID {
	return m.from
}

func (m KillExit) Reason() interface{} {
	return m.reason
}

func (m KillExit) Origin() SystemMessage {
	return m.origin
}
