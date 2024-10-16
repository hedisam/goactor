package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/supervision"
	"github.com/hedisam/goactor/supervision/strategy"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start a supervision tree
	err := supervision.Start(ctx,
		strategy.NewOneForOne(),
		[]supervision.ChildSpec{
			supervision.NewSupervisorSpec(
				"child-supervisor",
				strategy.NewOneForOne(),
				supervision.Permanent,
				[]supervision.ChildSpec{
					supervision.NewWorkerSpec(
						"Alice",
						supervision.Permanent,
						func() goactor.Actor {
							alice := newPanicActor("Alice")
							return goactor.NewActor(alice.Receive, goactor.WithInitFunc(alice.Init))
						},
					),
					supervision.NewWorkerSpec(
						"Bob",
						supervision.Permanent,
						func() goactor.Actor {
							bob := newPanicActor("Bob")
							return goactor.NewActor(bob.Receive, goactor.WithInitFunc(bob.Init))
						},
					),
				},
			),
		},
	)
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		t := time.NewTicker(100 * time.Millisecond)
		defer t.Stop()
		defer wg.Done()
		for {
			select {
			case <-t.C:
				err = goactor.Send(ctx, goactor.NamedPID("Alice"), "hi there")
				if err != nil {
					if errors.Is(err, goactor.ErrActorNotFound) {
						fmt.Println("[!] Actor 'Alice' not found; parent supervisor exited since child supervisor got restarted too many times")
						return
					}
					panic(err)
				}
			}
		}
	}()

	wg.Wait()

	err = goactor.Send(ctx, goactor.NamedPID("Bob"), "Hey Bob, you there?")
	if err != nil {
		if errors.Is(err, goactor.ErrActorNotFound) {
			fmt.Println("[!] Bob is not available either")
			return
		}
		panic(err)
	}
}

type panicActor struct {
	name string
}

func newPanicActor(name string) *panicActor {
	return &panicActor{name: name}
}

func (a *panicActor) Receive(_ context.Context, msg any) (loop bool, err error) {
	switch a.name {
	case "Alice":
		fmt.Println("Alice received message; PANIC")
		panic(msg)
	case "Bob":
		fmt.Println("Bob received msg:", msg)
	default:
		fmt.Printf("Unknown actor %q received message: %v\n", a.name, msg)
		os.Exit(1)
	}
	return true, nil
}

func (a *panicActor) Init(context.Context, *goactor.PID) error {
	switch a.name {
	case "Alice":
		fmt.Println("Alice initialised")
	case "Bob":
		fmt.Println("Bob initialised. If not first init, then the child supervisor has been restarted!")
	}
	return nil
}
