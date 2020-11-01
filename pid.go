package goactor

type PID interface {
	sendMessage(msg interface{}) error
	sendSystemMessage(msg interface{}) error
	getRelations() *Relations
	ID() string
}
