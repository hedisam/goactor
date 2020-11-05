package pid

type InternalPID interface {
	ID() string

	SendMessage(msg interface{}) error
	SendSystemMessage(msg interface{}) error

	Link(to InternalPID) error
	Unlink(who InternalPID) error
	AddMonitor(parent InternalPID) error
	RemMonitor(parent InternalPID) error
}

type relationManager interface {
	AddLink(pid InternalPID)
	RemoveLink(pid InternalPID)

	AddMonitor(pid InternalPID)
	RemoveMonitor(pid InternalPID)
}

type mailbox interface {
	PushMessage(msg interface{}) error
	PushSystemMessage(msg interface{}) error
}
