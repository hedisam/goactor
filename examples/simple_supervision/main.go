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
	"github.com/hedisam/goactor/supervision/strategy"
)

func main() {
	err := supervision.Start(
		context.Background(),
		strategy.NewOneForOne(
			strategy.WithPeriod(time.Millisecond*500),
			strategy.WithMaxRestarts(2),
		),
		[]supervision.ChildSpec{
			supervision.NewWorkerSpec(
				":alice",
				supervision.Permanent,
				func() goactor.Actor {
					log.Println("[!] Alice spawned")
					return goactor.ReceiveFunc(aliceReceiver)
				},
			),
		},
	)
	if err != nil {
		panic(err)
	}

	err = goactor.Send(context.Background(), goactor.NamedPID(":alice"), "hey alice what's up?")
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		t := time.NewTicker(300 * time.Millisecond)
		defer t.Stop()
		defer wg.Done()
		for {
			select {
			case <-t.C:
				err = goactor.Send(context.Background(), goactor.NamedPID(":alice"), ":panic")
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

func aliceReceiver(_ context.Context, msg any) error {
	fmt.Println(":alice received msg:", msg)
	if msg == ":panic" {
		panic(msg)
	}
	return nil
}
