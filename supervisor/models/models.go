package models

import (
	"github.com/hedisam/goactor/internal/intlpid"
	"github.com/hedisam/goactor/internal/relations"
	"github.com/hedisam/goactor/supervisor/childstate"
	"github.com/hedisam/goactor/sysmsg"
)

type RelationManager interface {
	AddLink(pid intlpid.InternalPID) error
	RemoveLink(pid intlpid.InternalPID) error

	LinkedActors() *relations.RelationIterator
	MonitorActors() *relations.RelationIterator

	Dispose()
}

type Mailbox interface {
	Receive(msgHandler, sysMsgHandler func(interface{}) bool) error
	Dispose()
}

type SupHandler interface {
	Run(update sysmsg.SystemMessage) bool
}

type StrategyHandler interface {
	Apply(*childstate.ChildState) error
}

type InitMsg struct {
	SenderPID intlpid.InternalPID
}

func (m *InitMsg) Reason() interface{} {
	return "init_supervisor"
}

func (m *InitMsg) Sender() intlpid.InternalPID {
	return m.SenderPID
}

func (m *InitMsg) Origin() sysmsg.SystemMessage {
	return nil
}
