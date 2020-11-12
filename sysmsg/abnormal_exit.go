package sysmsg

import "github.com/hedisam/goactor/internal/intlpid"

type AbnormalExit struct {
	from   intlpid.InternalPID
	reason interface{}
	origin SystemMessage
}

func NewAbnormalExitMsg(from intlpid.InternalPID, reason interface{}, origin SystemMessage) *AbnormalExit {
	return &AbnormalExit{
		from:   from,
		reason: reason,
		origin: origin,
	}
}

func (m *AbnormalExit) Sender() intlpid.InternalPID {
	return m.from
}

func (m *AbnormalExit) Reason() interface{} {
	return m.reason
}

func (m *AbnormalExit) Origin() SystemMessage {
	return m.origin
}
