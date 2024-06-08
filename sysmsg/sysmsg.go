package sysmsg

import "context"

// MessageType defines the system message type.
type MessageType string

const (
	// NormalExit is a type of system message emitted when an ActorHandler exists normally.
	NormalExit MessageType = "system:message:exit:normal"
	// AbnormalExit is a type of system message emitted when an ActorHandler exists abnormally.
	AbnormalExit MessageType = "system:message:exit:abnormal"
	// Kill todo
	Kill MessageType = "system:message:kill"
	// Shutdown todo
	Shutdown MessageType = "system:message:shutdown"
)

type systemPID interface {
	ID() string
	PushSystemMessage(ctx context.Context, msg *Message) error
}

// Message holds details about a system message.
type Message struct {
	Sender systemPID
	Reason any
	Type   MessageType
	Origin *Message
}

// ToSystemMessage is a helper to quickly check if a message is of type *Message.
func ToSystemMessage(msg any) (sysMsg *Message, ok bool) {
	m, ok := msg.(*Message)
	return m, ok
}
