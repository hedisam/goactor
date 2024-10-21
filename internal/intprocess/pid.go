package intprocess

import "github.com/hedisam/goactor/internal/mailbox"

type PID interface {
	Ref() string
	Link(pid PID) error
	Unlink(pid PID) error
	Monitor(pid PID) error
	Demonitor(pid PID) error
	SetTrapExit(trapExit bool)
	Dispatcher() mailbox.Dispatcher

	disposed() bool
	addRelation(pid PID, typ relationType)
	remRelation(ref string, typ relationType)
}
