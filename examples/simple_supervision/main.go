package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/supervision"
)

func main() {
	err := supervision.StartSupervisor(
		context.Background(),
		supervision.OneForOneStrategy(),
		supervision.NewActorChildSpec(
			":alice",
			supervision.RestartAlways,
			goactor.NewActor(actorAlice),
		),
	)
	if err != nil {
		log.Fatal("Could not start supervisor:", err)
	}

	err = goactor.Send(context.Background(), goactor.NamedPID(":alice"), "hey alice what's up?")
	if err != nil {
		panic(err)
	}
	err = goactor.Send(context.Background(), goactor.NamedPID(":alice"), ":panic")
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second)
}

func actorAlice(_ context.Context, msg any) (loop bool, err error) {
	fmt.Println(":alice received msg:", msg)
	if msg == ":panic" {
		panic(msg)
	}
	return true, nil
}
