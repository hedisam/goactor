package intlpid

func SendMessage(pid InternalPID, msg interface{}) error {
	return pid.sendMessage(msg)
}

func SendSystemMessage(pid InternalPID, msg interface{}) error {
	return pid.sendSystemMessage(msg)
}

func Link(from, to InternalPID) error {
	return from.link(to)
}

func Unlink(from, who InternalPID) error {
	return from.unlink(who)
}

func AddMonitor(to, parent InternalPID) error {
	return to.addMonitor(parent)
}

func RemoveMonitor(from, parent InternalPID) error {
	return from.remMonitor(parent)
}

func Shutdown(who InternalPID, reason interface{}) {
	who.shutdown(reason)
}
