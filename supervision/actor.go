package supervision

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/internal/syspid"
	"github.com/hedisam/goactor/sysmsg"
)

// Supervisor is a supervisor Actor. It implements the goactor.Actor interface.
type Supervisor struct {
	strategy    *Strategy
	nameToChild map[string]ChildSpec
	self        *goactor.PID
}

// Init initialises the supervisor by spawning all the children.
func (s *Supervisor) Init(ctx context.Context, self *goactor.PID) (err error) {
	log.Println("Initialising supervisor...")
	s.self = self

	ctx, cancel := context.WithCancelCause(ctx)
	defer func() {
		if err != nil {
			// stop already spawned children
			cancel(err)
		}
	}()

	for name, child := range s.nameToChild {
		pid, err := child.StartLink(ctx)
		if err != nil {
			return fmt.Errorf("startlink %q: %w", name, err)
		}
		log.Printf("Started child %q with ID %q", name, pid)
		s.self.Link(pid, true)
		err = goactor.Register(name, pid)
		if err != nil {
			return fmt.Errorf("register child actor %q: %w", name, err)
		}
	}
	return nil
}

// Receive processes received system messages from its children.
func (s *Supervisor) Receive(ctx context.Context, msg any) (loop bool, err error) {
	fmt.Printf("[!] Supervisor received: %v\n", msg)
	return true, nil
}

// AfterFunc implements goactor.Actor.
func (s *Supervisor) AfterFunc() (timeout time.Duration, afterFunc goactor.AfterFunc) {
	return 0, func(ctx context.Context) error {
		return nil
	}
}

func (s *Supervisor) stop(ctx context.Context, err error) {
	reason := any(":normal")
	typ := sysmsg.NormalExit
	if err != nil {
		reason = err
		typ = sysmsg.AbnormalExit
	}
	sendErr := syspid.Send(ctx, s.self.SystemPID, &sysmsg.Message{
		SenderID: s.self.ID(),
		Reason:   reason,
		Type:     typ,
	})
	if err != nil {
		log.Println("Failed to send closure system message to supervisor", s.self.ID(), sendErr)
	}
}
