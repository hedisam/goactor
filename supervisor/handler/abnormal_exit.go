package handler

import (
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

var abnormalHandler *AbnormalExitHandler

type AbnormalExitHandler struct {
	service supervisorService
}

func GetAbnormalHandler(s supervisorService) *AbnormalExitHandler {
	if abnormalHandler == nil {
		abnormalHandler = &AbnormalExitHandler{service: s}
	}
	return abnormalHandler
}

func (h *AbnormalExitHandler) Run(update sysmsg.SystemMessage) bool {
	childState, ok := h.service.GetChildByPID(update.Sender())
	if !ok {
		// no child found with the given internal_pid
		return true
	}
	// check the child's restart type
	switch childState.RestartWhen() {
	case RestartAlways, RestartTransient:
		strategyHandler := h.service.Strategy()
		err := strategyHandler.Apply(childState)
		if err != nil {
			log.Printf("[!] supervisor, failed restarting child %s from an abnormal exit, err: %v\n",
				childState.Name(), err)
		}
		break
	case RestartNever:
		h.service.DisposeChild(childState)
	}
	return true
}
