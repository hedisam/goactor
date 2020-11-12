package supervisor

import (
	"github.com/hedisam/goactor"
	p "github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

type initHandler struct{}

func (h *initHandler) run(service *SupService, update sysmsg.SystemMessage) bool {
	initErr := service.startChildren()
	sender := p.ToPID(update.Sender())
	err := goactor.Send(sender, initErr)
	if err != nil {
		log.Println("supervisor failed to send back initMsg: %w", err)
	}
	return initErr == nil && err == nil
}

type normalExitHandler struct{}

func (h *normalExitHandler) run(service *SupService, update sysmsg.SystemMessage) bool {
	childState, ok := service.childrenManager.GetByPID(update.Sender())
	if !ok {
		// no child found with the given internal_pid
		return true
	}
	// check the child's restart type
	switch childState.RestartWhen() {
	case RestartAlways:
		if err := childState.Restart(); err != nil {
			log.Printf("[!] supervisor, failed restarting child %s from a normal exit, err: %v\n",
				childState.Name(), err)
		}
		break
	case RestartNever, RestartTransient:
		service.disposeChild(childState)
	}
	return true
}

type abnormalExitHandler struct{}

func (h *abnormalExitHandler) run(service *SupService, update sysmsg.SystemMessage) bool {
	childState, ok := service.childrenManager.GetByPID(update.Sender())
	if !ok {
		// no child found with the given internal_pid
		return true
	}
	// check the child's restart type
	switch childState.RestartWhen() {
	case RestartAlways, RestartTransient:
		if err := childState.Restart(); err != nil {
			log.Printf("[!] supervisor, failed restarting child %s from an abnormal exit, err: %v\n",
				childState.Name(), err)
		}
		break
	case RestartNever:
		service.disposeChild(childState)
	}
	return true
}

type killExitHandler struct{}

func (h *killExitHandler) run(service *SupService, update sysmsg.SystemMessage) bool {
	childState, ok := service.childrenManager.GetByPID(update.Sender())
	if !ok {
		// no child found with the given internal_pid
		return true
	}
	// check the child's restart type
	switch childState.RestartWhen() {
	case RestartAlways, RestartTransient:
		if err := childState.Restart(); err != nil {
			log.Printf("[!] supervisor, failed restarting child %s from an kill exit, err: %v\n",
				childState.Name(), err)
		}
		break
	case RestartNever:
		service.disposeChild(childState)
	}
	return true
}

type shutdownCMDHandler struct{}

func (h *shutdownCMDHandler) run(service *SupService, update sysmsg.SystemMessage) bool {
	service.shutdown(update)
	return false
}

type defaultHandler struct{}

func (h *defaultHandler) run(service *SupService, update sysmsg.SystemMessage) bool {
	log.Printf("[!] supervisor %s received an unknown message: %v\n", service.supervisor.Self().ID(), update)
	return true
}
