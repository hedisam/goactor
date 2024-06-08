package goactor

// SystemMessageType defines the system message type.
type SystemMessageType string

const (
	// SystemMessageNormalExit is a type of system message emitted when an ActorHandler exists normally.
	SystemMessageNormalExit SystemMessageType = "system:message:exit:normal"
	// SystemMessageAbnormalExit is a type of system message emitted when an ActorHandler exists abnormally.
	SystemMessageAbnormalExit SystemMessageType = "system:message:exit:abnormal"
	// SystemMessageKill todo
	SystemMessageKill SystemMessageType = "system:message:kill"
	// SystemMessageShutdown todo
	SystemMessageShutdown SystemMessageType = "system:message:shutdown"
)

// SystemMessage holds details about a system message.
type SystemMessage struct {
	Sender *PID
	Reason any
	Type   SystemMessageType
	Origin *SystemMessage
}

// ToSystemMessage is a helper to quickly check if a message is of type *SystemMessage.
func ToSystemMessage(msg any) (sysMsg *SystemMessage, ok bool) {
	m, ok := msg.(*SystemMessage)
	return m, ok
}
