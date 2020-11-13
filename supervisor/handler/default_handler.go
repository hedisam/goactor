package handler

import (
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

type DefaultHandler struct {
	service supervisorService
	message interface{}
}

func NewDefaultHandler(s supervisorService, message interface{}) *DefaultHandler {
	return &DefaultHandler{
		service: s,
		message: message,
	}
}

func (h *DefaultHandler) Run(_ sysmsg.SystemMessage) bool {
	log.Printf("[!] supervisor %s received an unknown message: %v\n", h.service.Self().ID(), h.message)
	return true
}
