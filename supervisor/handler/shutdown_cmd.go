package handler

import "github.com/hedisam/goactor/sysmsg"

type ShutdownCMDHandler struct {
	service supervisorService
}

func NewShutdownCMDHandler(s supervisorService) *ShutdownCMDHandler {
	return &ShutdownCMDHandler{service: s}
}

func (h *ShutdownCMDHandler) Run(update sysmsg.SystemMessage) bool {
	h.service.Shutdown(update)
	return false
}
