package supervisor

import (
	"github.com/hedisam/goactor/supervisor/handler"
	"github.com/hedisam/goactor/supervisor/models"
	"github.com/hedisam/goactor/supervisor/supref"
	"github.com/hedisam/goactor/sysmsg"
)

func (service *SupService) listen(supervisor *Supervisor) {

	supervisor.Receive(func(message interface{}) (loop bool) {
		supHandler, update := service.getHandler(message)
		return supHandler.Run(update)
	})
}

func (service *SupService) getHandler(message interface{}) (models.SupHandler, sysmsg.SystemMessage) {
	switch update := message.(type) {
	case models.InitMsg:
		// spawn the supervisor's childrenManager
		return handler.NewInitHandler(service), &update
	case *sysmsg.NormalExit:
		// some child actor has terminated normally.
		return handler.GetNormalExitHandler(service), update
	case *sysmsg.AbnormalExit:
		// some child actor has exited abnormally.
		return handler.GetAbnormalHandler(service), update
	case *sysmsg.KillExit:
		// some child actor(supervisor) has killed its process.
		return handler.GetKillExitHandler(service), update
	case *sysmsg.ShutdownCMD:
		// the parent supervisor wants us to Shutdown
		return handler.NewShutdownCMDHandler(service), update
	case supRefRequest:
		update.SetSupervisorService(service)
		return update, nil
	default:
		return handler.NewDefaultHandler(service, message), nil
	}
}

type supRefRequest interface {
	SetSupervisorService(service supref.SupervisorService)
	Run(message sysmsg.SystemMessage) bool
}
