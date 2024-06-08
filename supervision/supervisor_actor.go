package supervision

import (
	"context"
	"log"

	"github.com/hedisam/goactor"
)

type Supervisor struct {
	strategy    *Strategy
	nameToChild map[string]ChildSpec
	self        *goactor.PID
}

func (s *Supervisor) start(ctx context.Context) {
	log.Println("Starting supervisor...")
	_ = goactor.Spawn(
		ctx,
		s.Receive,
		goactor.WithInitFunc(s.Init),
	)

	for name, child := range s.nameToChild {
		pid := child.StartLink(ctx)
		log.Printf("Child %q with ID %q started\n", name, pid)
	}
}

func (s *Supervisor) Init(ctx context.Context, pid *goactor.PID) {
	log.Println("Initialising supervisor...")
	s.self = pid
}

func (s *Supervisor) Receive(ctx context.Context, msg any) (loop bool, err error) {
	return true, nil
}
