package childstate

import "github.com/hedisam/goactor/pid"

type supService interface {
	Link(*pid.PID) error
	RestartsPeriod() int
	MaxRestartsAllowed() int
	MaxRestartsReached()
}

type Spec interface {
	StartLink() (*pid.PID, error)
	RestartWhen() int
	Name() string
}
