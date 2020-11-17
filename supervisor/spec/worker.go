package spec

import (
	"github.com/google/uuid"
	"github.com/hedisam/goactor"
	p "github.com/hedisam/goactor/pid"
	"github.com/hedisam/goactor/supervisor/option"
	"strings"
)

type WorkerSpec struct {
	Id             string
	actorFunc      goactor.ActorFunc
	mailboxBuilder goactor.MailboxBuilderFunc
	WhenToRestart  int
}

func (w WorkerSpec) StartLink() (*p.PID, error) {
	pid := goactor.Spawn(w.actorFunc, w.mailboxBuilder)
	return pid, nil
}

func (w WorkerSpec) SupervisorOptions() *option.Options {
	return nil
}

func (w WorkerSpec) RestartWhen() int {
	return w.WhenToRestart
}

func (w WorkerSpec) Name() string {
	return w.Id
}

func (w WorkerSpec) SetMailboxBuilder(fn goactor.MailboxBuilderFunc) WorkerSpec {
	w.mailboxBuilder = fn
	return w
}

func NewWorkerSpec(name string, restartWhen int, fn goactor.ActorFunc) WorkerSpec {
	if strings.TrimSpace(name) == "" {
		name = uuid.New().String()
	}
	w := WorkerSpec{
		Id:            name,
		actorFunc:     fn,
		WhenToRestart: restartWhen,
	}
	return w
}
