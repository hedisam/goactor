package handler

import (
	"github.com/hedisam/goactor/internal/intlpid"
	p "github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/supervisor/childstate"
	"github.com/hedisam/goactor/supervisor/models"
	"github.com/hedisam/goactor/sysmsg"
)

type supervisorService interface {
	Init() error
	Strategy() models.StrategyHandler
	GetChildByPID(pid intlpid.InternalPID) (*childstate.ChildState, bool)
	DisposeChild(state *childstate.ChildState)
	Shutdown(reason sysmsg.SystemMessage)
	Self() *p.PID
}

const (
	RestartAlways = iota
	RestartTransient
	RestartNever
)
