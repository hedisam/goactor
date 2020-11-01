package goactor
/*
	Exit messages
	1. Exit messages
;*/

type ExitMessage interface {
	MsgFrom() PID
	ExitReason() interface{}
}

type NormalExit struct {
	From PID
}

func (m NormalExit) MsgFrom() PID {
	return m.From
}

func (m NormalExit) ExitReason() interface{} {
	return "normal_exit"
}

type AbnormalExit struct {
	From PID
	Reason interface{}
}

func (m AbnormalExit) MsgFrom() PID {
	return m.From
}

func (m AbnormalExit) ExitReason() interface{} {
	return m.Reason
}

