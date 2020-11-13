package handler

import (
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

var normalExitHandler *NormalExitHandler

type NormalExitHandler struct {
	service supervisorService
}

func GetNormalExitHandler(s supervisorService) *NormalExitHandler {
	if normalExitHandler == nil {
		normalExitHandler = &NormalExitHandler{service: s}
	}
	return normalExitHandler
}

func (h *NormalExitHandler) Run(update sysmsg.SystemMessage) bool {
	childState, ok := h.service.GetChildByPID(update.Sender())
	if !ok {
		// no child found with the given internal_pid
		return true
	}
	// check the child's restart type
	switch childState.RestartWhen() {
	case RestartAlways:
		strategyHandler := h.service.Strategy()
		err := strategyHandler.Apply(childState)
		if err != nil {
			log.Printf("[!] supervisor, failed restarting child %s from a normal exit, err: %v\n",
				childState.Name(), err)
		}
		break
	case RestartNever, RestartTransient:
		h.service.DisposeChild(childState)
	}
	return true
}
