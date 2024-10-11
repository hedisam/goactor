package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/supervision"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := supervision.StartSupervisor(ctx,
		supervision.OneForOneStrategy(),
		supervision.NewSupervisorChildSpec(
			"child-supervisor",
			supervision.OneForOneStrategy(),
			supervision.RestartAlways,
			supervision.NewActorChildSpec(
				"alice-1",
				supervision.RestartAlways,
				goactor.NewActor(newAliceActor("alice-1").Receive),
			),
			supervision.NewActorChildSpec(
				"alice-2",
				supervision.RestartAlways,
				goactor.NewActor(newAliceActor("alice-2").Receive),
			),
		),
	)
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()
		defer wg.Done()
		for {
			select {
			case <-t.C:
				err = goactor.Send(context.Background(), goactor.NamedPID("alice-1"), "hi there")
				if err != nil {
					if errors.Is(err, goactor.ErrActorNotFound) {
						log.Println("actor not found")
						return
					}
					panic(err)
				}
			}
		}
	}()

	wg.Wait()
}

type aliceActor struct {
	name string
}

func newAliceActor(name string) *aliceActor {
	return &aliceActor{name: name}
}

func (a *aliceActor) Receive(ctx context.Context, msg any) (loop bool, err error) {
	panic(fmt.Sprintf("alice %q received msg: %v", a.name, msg))
}
