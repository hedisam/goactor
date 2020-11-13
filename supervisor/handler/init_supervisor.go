package handler

import (
	"github.com/hedisam/goactor"
	p "github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

type InitHandler struct {
	service supervisorService
}

func NewInitHandler(s supervisorService) *InitHandler {
	return &InitHandler{service: s}
}

func (h *InitHandler) Run(update sysmsg.SystemMessage) bool {
	initErr := h.service.Init()
	sender := p.ToPID(update.Sender())
	err := goactor.Send(sender, initErr)
	if err != nil {
		log.Println("supervisor failed to send back initMsg: %w", err)
	}
	return initErr == nil && err == nil
}
