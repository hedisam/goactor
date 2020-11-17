package childstate

import (
	"fmt"
	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/internal/intlpid"
	p "github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/sysmsg"
	"log"
	"time"
)

type ChildState struct {
	supService      supService
	spec            Spec
	childrenManager *ChildrenManager
	self            *p.PID
	dead            bool
	// we save each restart's timestamp to check if they've occurred in the specified spec's period or not.
	restarts []int64
}

// IsSupervisor returns true if the child process is a supervisor.
// In that case we'd have a supervision tree.
func (child *ChildState) IsSupervisor() bool {
	return child.self.IsSupervisor()
}

// Dead returns true if this child has been disposed or declared as dead
func (child *ChildState) Dead() bool {
	return child.dead
}

// Name returns the child's name or id
func (child *ChildState) Name() string {
	return child.spec.Name()
}

// RestartWhen returns the restart type specified for this child, which could one of the three values:
// RestartAlways, RestartTransient, RestartNever
func (child *ChildState) RestartWhen() int {
	return child.spec.RestartWhen()
}

// PID returns the child process's pid. The returned pid could be disposed.
func (child *ChildState) PID() *p.PID {
	return child.self
}

// Restart disposes the old child's pid and re-spawns a new process for the given child spec.
// The supervisor will panic/shutdown if this child has been restarted more than the allowed max-restarts
// specified in the supervisor's option.Options
func (child *ChildState) Restart() error {
	if child.hasReachedMaxRestarts() {
		log.Println("[!] supervisor reached max restarts")
		// time to shutdown this supervisor
		// this method will panic. so no need to return
		child.supService.MaxRestartsReached()
	}

	child.supService.DisposeChild(child)

	err := child.Start()
	if err != nil {
		return fmt.Errorf("failed to Restart the child #%s: %w", child.spec.Name(), err)
	}
	// add a restart timestamp to the list
	child.restarts = append(child.restarts, time.Now().Unix())

	return nil
}

// Start spawns a new child process for the given child spec, which is will be linked to the supervisor.
// The child spec can be a worker actor or a supervisor.
func (child *ChildState) Start() error {
	// invoke the function that spawns the child process
	pid, err := child.spec.StartLink()
	if err != nil {
		return fmt.Errorf("supervisor failed to Start the child #%s: %w", child.spec.Name(), err)
	}
	child.self = pid
	// link the supervisor to the child
	err = child.supService.Link(pid)
	if err != nil {
		return fmt.Errorf("supervisor failed linking to child #%s: %w", child.spec.Name(), err)
	}

	// index the internal_pid
	child.childrenManager.Index(pid.InternalPID(), child.Name())
	child.dead = false

	// register the child in the process registry by its name
	goactor.Register(child.Name(), pid)
	return nil
}

// hasReachedMaxRestarts returns true if the child has restarted more than max-restarts which is specified in the
// supervisor's option.Options.
func (child *ChildState) hasReachedMaxRestarts() bool {
	// restarts that are not expired, meaning they are in the same last Period
	var restartsNotEx []int64

	now := time.Now()
	periodStartTime := now.Add(time.Duration(-1*child.supService.RestartsPeriod()) * time.Second).Unix()
	// check how many restarts we've got in the same Period
	for _, restartTime := range child.restarts {
		if restartTime >= periodStartTime {
			// this restart has occurred in the Period
			restartsNotEx = append(restartsNotEx, restartTime)
		}
	}

	if len(restartsNotEx) >= child.supService.MaxRestartsAllowed() {
		// we've got restarts more than the allowed MaxRestarts in the same Period
		return true
	}

	// get rid of expired timestamps
	child.restarts = restartsNotEx
	return false
}

// Shutdown declares the child as dead and then attempts to shutdown it by triggering the context.Context's cancel func
// of the child actor.
// Note: There's no way to directly terminate a goroutine so worker actors that are doing time intensive tasks
// should pass around and check actor's context.Context to see if it's cancelled or not.
func (child *ChildState) Shutdown(reason sysmsg.SystemMessage) {
	if child.dead {
		return
	}
	child.DeclareDead()
	intlpid.Shutdown(child.self.InternalPID(), reason)
}

// DeclareDead removes the child's pid from the children manager's index and unregisters the process from the
// processes registry.
// So if by any chances we got a new message from the previous dead pid which has been shutdown by the supervisor,
// we can treat that message as an invalid one and do nothing.
// This way we show that we're only interested in the new pid, or new respawned actor.
func (child *ChildState) DeclareDead() {
	child.childrenManager.RemoveIndex(child.self.InternalPID())
	child.dead = true
	goactor.Unregister(child.Name())
}

func NewChildState(spec Spec, supRef supService, manager *ChildrenManager) *ChildState {
	return &ChildState{spec: spec, supService: supRef, childrenManager: manager}
}
