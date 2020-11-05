package goactor

import (
	"github.com/hedisam/goactor/internal/pid"
)

type ExitMessage interface {
	MsgFrom() pid.InternalPID
	ExitReason() interface{}
}

type NormalExit struct {
	From pid.InternalPID
}

func (m NormalExit) MsgFrom() pid.InternalPID {
	return m.From
}

func (m NormalExit) ExitReason() interface{} {
	return "normal_exit"
}

type AbnormalExit struct {
	From   pid.InternalPID
	Reason interface{}
}

func (m AbnormalExit) MsgFrom() pid.InternalPID {
	return m.From
}

func (m AbnormalExit) ExitReason() interface{} {
	return m.Reason
}
