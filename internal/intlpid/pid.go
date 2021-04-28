package intlpid

type InternalPID interface {
	ID() string
	IsSupervisor() bool

	sendMessage(msg interface{}) error
	sendSystemMessage(msg interface{}) error

	link(to InternalPID) error
	unlink(who InternalPID) error
	addMonitor(parent InternalPID) error
	remMonitor(parent InternalPID) error

	// Shutdown will shutdown the actor by closing its context's done channel. We're not disposing the mailbox,
	// so we'll be able to receive the system message that's causing the shutdown and notifying related actors with
	// a proper message.
	// specifically used by supervisor
	shutdown(reason interface{})
}

type relationManager interface {
	AddLink(pid InternalPID) error
	RemoveLink(pid InternalPID) error

	AddMonitor(pid InternalPID) error
	RemoveMonitor(pid InternalPID) error
}

type mailbox interface {
	PushMessage(msg interface{}) error
	PushSystemMessage(msg interface{}) error
}
