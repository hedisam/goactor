package spec

import (
	"github.com/hedisam/goactor/pid"
)

type StartLink func() (*pid.PID, error)

const (
	RestartAlways = iota
	RestartTransient
	RestartNever
)
