package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/examples/require"
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
							return newPanicActor("Alice")
						},
					),
					supervision.NewWorkerSpec(
						"Bob",
						supervision.Permanent,
						func() goactor.Actor {
							return newPanicActor("Bob")
						},
					),
				},
			),
		},
	)
	require.NoError(err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		t := time.NewTicker(100 * time.Millisecond)
		defer t.Stop()
		defer wg.Done()
		for {
			select {
			case <-t.C:
				err = goactor.Send(ctx, goactor.Named("Alice"), "hi there")
				if err != nil {
					if errors.Is(err, goactor.ErrNamedActorNotFound) {
						fmt.Println("[!] Actor 'Alice' not found; parent supervisor exited since child supervisor got restarted too many times")
						return
					}
					require.NoError(err)
				}
			}
		}
	}()

	wg.Wait()

	err = goactor.Send(ctx, goactor.Named("Bob"), "Hey Bob, you there?")
	if errors.Is(err, goactor.ErrNamedActorNotFound) {
		fmt.Println("[!] Bob is not available either")
		return
	}
	require.NoError(err)
}

type panicActor struct {
	name string
}

func newPanicActor(name string) *panicActor {
	return &panicActor{name: name}
}

func (a *panicActor) Receive(_ context.Context, msg any) error {
	switch a.name {
	case "Alice":
		fmt.Println("Alice received message; PANIC")
		panic(msg)
	case "Bob":
		fmt.Println("Bob received msg:", msg)
	}
	return fmt.Errorf("unknown actor %q received message: %v\n", a.name, msg)
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
