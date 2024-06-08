package goactor

// SystemMessageType defines the system message type.
type SystemMessageType string

const (
	// SystemMessageDefault is just a normal & default system message that should not affect anything.
	SystemMessageDefault = "system:message:default"
	// SystemMessageAbnormalExit is a type of system message emitted when an ActorHandler exists abnormally.
	SystemMessageAbnormalExit = "system:message:exit:abnormal"
	// SystemMessageNormalExit is a type of system message emitted when an ActorHandler exists normally.
	SystemMessageNormalExit = "system:message:exit:normal"
	// SystemMessageKill todo
	SystemMessageKill = "system:message:kill"
	// SystemMessageShutdown todo
	SystemMessageShutdown = "system:message:shutdown"
)

// SystemMessage holds details about a system message.
type SystemMessage struct {
	Sender *PID
	Reason any
	Type   SystemMessageType
	Origin *SystemMessage
}

// IsSystemMessage is a helper to quickly check if a message is of type *SystemMessage.
func IsSystemMessage(msg any) (sysMsg *SystemMessage, ok bool) {
	m, ok := msg.(*SystemMessage)
	return m, ok
}
