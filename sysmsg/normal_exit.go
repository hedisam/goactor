package sysmsg

import "github.com/hedisam/goactor/internal/intlpid"

type NormalExit struct {
	from   intlpid.InternalPID
	origin SystemMessage
}

func NewNormalExitMsg(from intlpid.InternalPID, originalMsg SystemMessage) *NormalExit {
	return &NormalExit{
		from:   from,
		origin: originalMsg,
	}
}

func (m *NormalExit) Sender() intlpid.InternalPID {
	return m.from
}

func (m *NormalExit) Reason() interface{} {
	return "normal_exit"
}

func (m *NormalExit) Origin() SystemMessage {
	return m.origin
}
