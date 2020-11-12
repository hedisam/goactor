package supervisor

import (
	"github.com/hedisam/goactor/sysmsg"
)

func (supService *SupService) listen(supervisor *Supervisor) {
	normalHandler := &normalExitHandler{}
	abnormalHandler := &abnormalExitHandler{}
	killHandler := &killExitHandler{}
	shutdownHandler := &shutdownCMDHandler{}
	defaultHandler := &defaultHandler{}

	supervisor.Receive(func(message interface{}) (loop bool) {
		switch update := message.(type) {
		case initMsg:
			// spawn the supervisor's childrenManager
			initializer := &initHandler{}
			return initializer.run(supService, &update)
		case sysmsg.NormalExit:
			// some child actor has terminated normally.
			return normalHandler.run(supService, &update)
		case sysmsg.AbnormalExit:
			// some child actor has exited abnormally.
			return abnormalHandler.run(supService, &update)
		case sysmsg.KillExit:
			// some child actor(supervisor) has killed its process.
			return killHandler.run(supService, &update)
		case sysmsg.ShutdownCMD:
			// the parent supervisor wants us to shutdown
			return shutdownHandler.run(supService, &update)
		default:
			return defaultHandler.run(supService, nil)
		}
	})
}
