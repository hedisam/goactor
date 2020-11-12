package supervisor

import (
	"github.com/hedisam/goactor/internal/intlpid"
	"github.com/hedisam/goactor/sysmsg"
)

type relationManager interface {
	AddLink(pid intlpid.InternalPID)
	RemoveLink(pid intlpid.InternalPID)
}

type Mailbox interface {
	Receive(msgHandler, sysMsgHandler func(interface{}) bool)
	Dispose()
}

type initMsg struct {
	sender intlpid.InternalPID
}

func (m *initMsg) Reason() interface{} {
	return "init_supervisor"
}

func (m *initMsg) Sender() intlpid.InternalPID {
	return m.sender
}

func (m *initMsg) Origin() sysmsg.SystemMessage {
	return nil
}
