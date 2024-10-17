package sysmsg

import "errors"

// Type represents the type of system message.
type Type string

const (
	// Down is used when a monitored process exits, regardless of the reason for its termination.
	Down Type = ":DOWN"
	// Exit is used when a linked process exits and trap_exit is enabled.
	Exit Type = ":EXIT"
	// Signal is used to signal an actor for termination.
	Signal Type = ":signal"
)

// Message is sent to the user-defined actor's receiver either when a monitored process exits regardless of
// the reason or when a linked process exits and trap_exit is enabled.
type Message struct {
	// Type is either Down, Exit, or Signal.
	Type Type
	// ProcessID is the process ID of the actor that has terminated. It can be an empty string if it's a direct Signal.
	ProcessID string
	// Reason is the reason of termination.
	Reason Reason
}

// Reason is an internal notifications about process termination or exit reasons.
// A Signal is handled internally by an actor which may result in a Message (e.g. if linked and trapping exit messages)
// You can use any error as Reason which will be considered an abnormal exit causing linked actors to terminate if
// not trapping exit messages.
type Reason error

var (
	// ReasonNormal is used when an actor exits normally. A ReasonNormal does not terminate a linked actor even if
	// it's not trapping exit messages.
	// An actor can trap exit ReasonNormal.
	ReasonNormal Reason = errors.New(":normal")
	// ReasonShutdown can be used to shut down an actor and its linked ones if they're not trapping exit messages.
	// An actor can trap exit ReasonShutdown.
	ReasonShutdown Reason = errors.New(":shutdown")
	// ReasonKill can be used to immediately dispose an actor and its linked ones if they're not trapping exit messages.
	// An actor cannot trap exit ReasonKill.
	ReasonKill Reason = errors.New(":kill")
)

// ToMessage checks whether the message is a system *Message or not.
func ToMessage(message any) (*Message, bool) {
	msg, ok := message.(*Message)
	return msg, ok
}
