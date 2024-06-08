package main

import (
	"context"
	"fmt"
	"github.com/hedisam/goactor/supervision"
)

func main() {
	err := supervision.StartSupervisor(
		context.Background(),
		supervision.OneForOneStrategy(),
		supervision.NewActorChildSpec(
			":alice",
			supervision.RestartAlways,
			actorAlice,
		),
	)
	if err != nil {
		panic(err)
	}
}

func actorAlice(ctx context.Context, msg any) (loop bool, err error) {
	fmt.Println("Child actor :alice received msg:", msg)
	return true, nil
}
