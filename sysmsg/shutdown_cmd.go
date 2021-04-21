package sysmsg

import (
	"github.com/hedisam/goactor/internal/intlpid"
)

type ShutdownCMD struct {
	from   intlpid.InternalPID
	reason interface{}
	origin SystemMessage
}

func NewShutdownCMD(from intlpid.InternalPID, reason interface{}, origin SystemMessage) ShutdownCMD {
	return ShutdownCMD{
		from:   from,
		reason: reason,
		origin: origin,
	}
}

func (cmd ShutdownCMD) Sender() intlpid.InternalPID {
	return cmd.from
}

func (cmd ShutdownCMD) Reason() interface{} {
	return cmd.reason
}

func (cmd ShutdownCMD) Origin() SystemMessage {
	return cmd.origin
}
