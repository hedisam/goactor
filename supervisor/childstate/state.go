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

func (child *ChildState) Name() string {
	return child.spec.Name()
}

func (child *ChildState) RestartWhen() int {
	return child.spec.RestartWhen()
}

func (child *ChildState) PID() *p.PID {
	return child.self
}

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

func (child *ChildState) Start() error {
	// spawn the child
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

	// register the child in the process registry by its name
	goactor.Register(child.Name(), pid)
	child.dead = false
	return nil
}

// hasReachedMaxRestarts returns true if the child has restarted more than it's allowed in the specified Period of time.
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

func (child *ChildState) Shutdown(reason sysmsg.SystemMessage) {
	child.DeclareDead()
	intlpid.Shutdown(child.self.InternalPID(), reason)
}

// DeclareDead removes the child's internal_pid from the children manager index. so if by any chances we got a new message
// from the previous dead internal_pid which has been shutdown by the supervisor, we can treat that message as an invalid one
// and do nothing. that way we show that we're only interested in new internal_pid, or new respawned actors.
func (child *ChildState) DeclareDead() {
	// removing ourself from the index. this is like declaring this child actor as dead.
	child.childrenManager.RemoveIndex(child.self.InternalPID())
}

func NewChildState(spec Spec, supRef supService, manager *ChildrenManager) *ChildState {
	return &ChildState{spec: spec, supService: supRef, childrenManager: manager}
}
