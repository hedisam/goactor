package supervisor

import (
	"fmt"
	"github.com/hedisam/goactor/internal/intlpid"
	p "github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/supervisor/childstate"
	"github.com/hedisam/goactor/supervisor/models"
	"github.com/hedisam/goactor/supervisor/strategy"
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

// SupService is responsible for doing all the low level and management stuff of the supervisor
type SupService struct {
	supervisor      *Supervisor
	specs           map[string]Spec
	options         *Options
	childrenManager *childstate.ChildrenManager
	strategy        models.StrategyHandler
}

func (service *SupService) Init() error {
	// building the strategy
	switch service.options.Strategy {
	case StrategyOptionOneForOne:
		service.strategy = strategy.NewOneForOneStrategyHandler()
	case StrategyOptionOneForAll:
		service.strategy = strategy.NewOneForAllStrategyHandler(service)
	case StrategyOptionRestForOne:
		service.strategy = strategy.NewRestForOneStrategyHandler(service)
	default:
		return fmt.Errorf("supervisor's service init: invalid strategy type: %v", service.options.Strategy)
	}

	service.childrenManager = childstate.NewChildrenManager()
	return service.startChildren()
}

func (service *SupService) Self() *p.PID {
	return service.supervisor.Self()
}

func (service *SupService) ChildrenIterator() *childstate.ChildrenStateIterator {
	return service.childrenManager.Iterator()
}

func (service *SupService) Strategy() models.StrategyHandler {
	return service.strategy
}

func (service *SupService) startChildren() error {
	for id, spec := range service.specs {
		// create child's specific state
		childState := childstate.NewChildState(spec, service, service.childrenManager)
		// spawn the child for the first time
		err := childState.Start()
		if err != nil {
			return fmt.Errorf("supervisor failed to start the childrenManager: %w", err)
		}

		// keep tracking of alive actors
		service.childrenManager.Put(id, childState)
	}
	return nil
}

func (service *SupService) GetChildByPID(pid intlpid.InternalPID) (*childstate.ChildState, bool) {
	return service.childrenManager.GetByPID(pid)
}

func (service *SupService) Link(pid *p.PID) error {
	return service.supervisor.Link(pid)
}

func (service *SupService) RestartsPeriod() int {
	return service.options.Period
}

func (service *SupService) MaxRestartsAllowed() int {
	return service.options.MaxRestarts
}

func (service *SupService) MaxRestartsReached() {
	// shutting down this supervisor because a child reached its max allowed restarts in a specified period
	service.Shutdown(sysmsg.NewKillMessage(
		service.supervisor.Self().InternalPID(),
		"supervisor's child reached its max allowed restarts",
		nil),
	)
}

// Shutdown will unlink and shutdown each child and then panics
func (service *SupService) Shutdown(reason sysmsg.SystemMessage) {
	// iterating through all children
	iterator := service.childrenManager.Iterator()
	for iterator.HasNext() {
		childID := iterator.Value()
		if err := service.shutdownChild(childID, reason); err != nil {
			log.Printf("[!] Shutdown supervisor: error while shutting down a child #%s, err: %v\n", childID.Name(), err)
		}
	}

	panic(reason)
}

func (service *SupService) shutdownChildByName(name string, reason sysmsg.SystemMessage) error {
	child, ok := service.childrenManager.Get(name)
	if !ok {
		return fmt.Errorf("failed to Shutdown child #%s: child doesn't exist", child.Name())
	}

	return service.shutdownChild(child, reason)
}

func (service *SupService) shutdownChild(child *childstate.ChildState, reason sysmsg.SystemMessage) error {
	// unlink supervisor from the child
	err := service.supervisor.Unlink(child.PID())
	if err != nil {
		return fmt.Errorf("failed to Shutdown child #%s - unlink failed: %w", child.Name(), err)
	}

	child.Shutdown(reason)
	return nil
}

// DisposeChild unlink the child and declares its internal_pid as a dead one.
func (service *SupService) DisposeChild(child *childstate.ChildState) {
	// this will remove the child from children manager's index
	child.DeclareDead()

	err := service.supervisor.Unlink(child.PID())
	if err != nil {
		return
	}
}

func newSupService(supervisor *Supervisor, specs map[string]Spec, options *Options) *SupService {
	return &SupService{
		supervisor:      supervisor,
		specs:           specs,
		options:         options,
		childrenManager: new(childstate.ChildrenManager),
	}
}
