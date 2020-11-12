package supervisor

import (
	"fmt"
	p "github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/supervisor/childstate"
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

// SupService is responsible for doing all the low level and management stuff of the supervisor
type SupService struct {
	supervisor      *Supervisor
	specs           map[string]Spec
	options         *Options
	childrenManager *childstate.ChildrenManager
}

func (supService *SupService) startChildren() error {
	for id, spec := range supService.specs {
		// create child's specific state
		childState := childstate.NewChildState(spec, supService, supService.childrenManager)
		// spawn the child for the first time
		err := childState.Start()
		if err != nil {
			return fmt.Errorf("supervisor failed to start the childrenManager: %w", err)
		}

		// keep tracking of alive actors
		supService.childrenManager.Put(id, childState)
	}
	return nil
}

func (supService *SupService) Link(pid *p.PID) error {
	return supService.supervisor.Link(pid)
}

func (supService *SupService) RestartsPeriod() int {
	return supService.options.Period
}

func (supService *SupService) MaxRestartsAllowed() int {
	return supService.options.MaxRestarts
}

func (supService *SupService) MaxRestartsReached() {
	// shutting down this supervisor because a child reached its max allowed restarts in a specified period
	supService.shutdown(sysmsg.NewKillMessage(
		supService.supervisor.Self().InternalPID(),
		"supervisor's child reached its max allowed restarts",
		nil),
	)
}

// shutdown will unlink and shutdown each child and then panics
func (supService *SupService) shutdown(reason sysmsg.SystemMessage) {
	// iterating through all children
	iterator := supService.childrenManager.Iterator()
	for iterator.HasNext() {
		childID := iterator.Value()
		if err := supService.shutdownChild(childID, reason); err != nil {
			log.Printf("[!] shutdown supervisor: error while shutting down a child #%s, err: %v\n", childID.Name(), err)
		}
	}

	panic(reason)
}

func (supService *SupService) shutdownChildByName(name string, reason sysmsg.SystemMessage) error {
	child, ok := supService.childrenManager.Get(name)
	if !ok {
		return fmt.Errorf("failed to shutdown child #%s: child doesn't exist", child.Name())
	}

	return supService.shutdownChild(child, reason)
}

func (supService *SupService) shutdownChild(child *childstate.ChildState, reason sysmsg.SystemMessage) error {
	// unlink supervisor from the child
	err := supService.supervisor.Unlink(child.PID())
	if err != nil {
		return fmt.Errorf("failed to shutdown child #%s - unlink failed: %w", child.Name(), err)
	}

	child.Shutdown(reason)
	return nil
}

// disposeChild unlink the child and declares its internal_pid as a dead one.
func (supService *SupService) disposeChild(child *childstate.ChildState) {
	err := supService.supervisor.Unlink(child.PID())
	if err != nil {
		return
	}
	supService.childrenManager.RemoveIndex(child.PID().InternalPID())
}

func newSupService(supervisor *Supervisor, specs map[string]Spec, options *Options) *SupService {
	return &SupService{
		supervisor:      supervisor,
		specs:           specs,
		options:         options,
		childrenManager: new(childstate.ChildrenManager),
	}
}
