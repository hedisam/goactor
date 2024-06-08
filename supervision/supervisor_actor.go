package supervision

import (
	"context"
	"fmt"
	"log"

	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/internal/syspid"
	"github.com/hedisam/goactor/sysmsg"
)

type Supervisor struct {
	strategy    *Strategy
	nameToChild map[string]ChildSpec
	self        *goactor.PID
}

func (s *Supervisor) start(ctx context.Context) (err error) {
	log.Println("Starting supervisor...")
	goactor.InitRegistry(goactor.DefaultRegistrySize)
	_ = goactor.Spawn(
		ctx,
		s.Receive,
		goactor.WithInitFunc(s.Init),
	)

	defer func() {
		s.stop(ctx, err)
	}()

	for name, child := range s.nameToChild {
		pid := child.StartLink(ctx)
		log.Printf("Child %q with ID %q started\n", name, pid)
		s.self.Link(pid, true)
		err := goactor.Register(name, pid)
		if err != nil {
			return fmt.Errorf("register child actor %q: %w", name, err)
		}
	}

	return nil
}

func (s *Supervisor) Init(ctx context.Context, pid *goactor.PID) {
	log.Println("Initialising supervisor...")
	s.self = pid
}

func (s *Supervisor) Receive(ctx context.Context, msg any) (loop bool, err error) {
	fmt.Printf("[!] Supervisor received: %v\n", msg)
	return true, nil
}

func (s *Supervisor) stop(ctx context.Context, err error) {
	reason := any(":normal")
	typ := sysmsg.NormalExit
	if err != nil {
		reason = err
		typ = sysmsg.AbnormalExit
	}
	sendErr := syspid.Send(ctx, s.self.SystemPID, &sysmsg.Message{
		Sender: syspid.NewSystemPID(s.self.ID(), nil), // todo?
		Reason: reason,
		Type:   typ,
	})
	if err != nil {
		log.Println("Failed to send closure system message to supervisor", s.self.ID(), sendErr)
	}
}
