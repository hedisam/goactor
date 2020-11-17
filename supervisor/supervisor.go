package supervisor

import (
	"fmt"
	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/internal/intlpid"
	"github.com/hedisam/goactor/internal/relations"
	"github.com/hedisam/goactor/mailbox"
	p "github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/supervisor/models"
	"github.com/hedisam/goactor/supervisor/option"
	"github.com/hedisam/goactor/supervisor/spec"
	"github.com/hedisam/goactor/supervisor/supref"
)

var noShutdown func()

// Start a new supervisor for the given children specifications. It returns a supervisor reference that can be used
// to interact with the supervisor.
// An error is returned if the supervisor's options or any of children specs are invalid.
func Start(options option.Options, specs ...spec.Spec) (*supref.SupRef, error) {
	specsMap, err := spec.SpecsToMap(specs...)
	if err != nil {
		return nil, fmt.Errorf("invalid specs: %w", err)
	}

	err = options.Validate()
	if err != nil {
		return nil, fmt.Errorf("invalid option: %w", err)
	}

	// creating a supervisor actor
	m := mailbox.NewQueueMailbox(1, 100, mailbox.DefaultMailboxTimeout, mailbox.DefaultGoSchedulerInterval)
	relationManager := relations.NewRelation()

	// we don't provide a Shutdown func to a supervisor's internal_pid
	// so the only way to Shutdown a supervisor is by sending a Shutdown Command
	pid := intlpid.NewLocalPID(m, relationManager, true, noShutdown)

	supervisor := newSupervisorActor(m, pid, relationManager)

	supService := newSupService(supervisor, specsMap, &options)
	// spawn our new supervisor
	spawn(supService)

	// sending an Init msg so the supervisor starts spawning its childrenManager
	future := goactor.NewFutureActor()
	err = intlpid.SendMessage(supervisor.Self().InternalPID(), models.InitMsg{SenderPID: future.Self().InternalPID()})
	if err != nil {
		return nil, fmt.Errorf("could not initialize supervisor: %w", err)
	}
	// wait for childrenManager to be spawned
	future.Receive(func(message interface{}) (loop bool) {
		switch msg := message.(type) {
		case error:
			err = msg
		}
		return false
	})
	if err != nil {
		return nil, fmt.Errorf("supervisor's initialization failed: %w", err)
	}

	supRef, _ := supref.ToSupervisorRef(supervisor.Self())
	return supRef, nil
}

func spawn(supService *SupService) {
	go func() {
		sup := supService.supervisor
		defer sup.dispose()
		supService.listen(sup)
	}()
}

// start is assigned to spec.SupervisorSpec's StartLink function to start a new supervisor child process.
// Basically the goal of this function is to decouple the spec.SupervisorSpec from the Start function when spawning a
// supervisor child process. So the spec package would not depend on its root package (this package).
func start(options option.Options, specs ...spec.Spec) (*p.PID, error) {
	supRef, err := Start(options, specs...)
	if err != nil {
		return nil, err
	}
	return supRef.PID(), nil
}

func init() {
	// here we are assigning the start function
	spec.SetStartLink(start)
}
