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
	err := supervision.StartSupervisor(
		context.Background(),
		supervision.NewStrategy(
			supervision.StrategyOneForOne,
			supervision.StrategyWithPeriod(time.Millisecond*500),
			supervision.StrategyWithMaxRestarts(2),
		),
		supervision.NewActorChildSpec(
			":alice",
			supervision.RestartAlways,
			goactor.NewActor(actorAlice, goactor.WithInitFunc(func(_ context.Context, _ *goactor.PID) error {
				fmt.Println("[!] Alice initialised")
				return nil
			})),
		),
	)
	if err != nil {
		log.Fatal("Could not start supervisor:", err)
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

func actorAlice(_ context.Context, msg any) (loop bool, err error) {
	fmt.Println(":alice received msg:", msg)
	if msg == ":panic" {
		panic(msg)
	}
	return true, nil
}
