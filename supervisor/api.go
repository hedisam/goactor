package supervisor

import (
	"fmt"
	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/internal/intlpid"
	"github.com/hedisam/goactor/internal/relations"
	"github.com/hedisam/goactor/mailbox"
)

var noShutdown func()

func Start(options Options, specs ...Spec) (*SupRef, error) {
	specsMap, err := specsToMap(specs...)
	if err != nil {
		return nil, fmt.Errorf("invalid specs: %w", err)
	}

	err = options.validate()
	if err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	// creating supervisor actor
	m := mailbox.NewQueueMailbox(1, 100, mailbox.DefaultMailboxTimeout, mailbox.DefaultGoSchedulerInterval)
	relationManager := relations.NewRelation()

	// we don't provide a shutdown func to a supervisor's internal_pid
	// so the only way to shutdown a supervisor is by sending a shutdown Command
	pid := intlpid.NewLocalPID(m, relationManager, true, noShutdown)

	supervisor := newSupervisorActor(m, pid, relationManager)

	supService := newSupService(supervisor, specsMap, &options)
	// spawn our new supervisor
	spawn(supService)

	// sending an init msg so the supervisor starts spawning its childrenManager
	future := goactor.NewFutureActor()
	err = intlpid.SendSystemMessage(future.Self().InternalPID(), initMsg{sender: future.Self().InternalPID()})
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

	supRef, _ := ToSupervisorRef(supervisor.Self())
	return supRef, nil
}

func spawn(supService *SupService) {
	go func() {
		sup := supService.supervisor
		defer sup.dispose()
		supService.listen(sup)
	}()
}
