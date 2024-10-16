package main

import (
	"context"
	"fmt"
	"time"

	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/sysmsg"
)

func main() {
	fmt.Println("----- Links & Monitors ----")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	child, _ := goactor.Spawn(ctx, goactor.NewActor(func(ctx context.Context, msg any) error {
		fmt.Printf("[ChildActor] message: %+v\n", msg)
		fmt.Printf("[ChildActor] sleeping for a bit then will go down\n")
		time.Sleep(time.Millisecond * 100)
		return sysmsg.ReasonShutdown // stop with a shutdown reason
	}))

	parent := goactor.NewActor(func(ctx context.Context, msg any) error {
		if processID, reason, ok := sysmsg.LinkedActorDown(msg); ok {
			fmt.Printf("[ParentActor] Linked actor %q terminated with reason %q\n", processID, reason)
			return nil
		}
		fmt.Printf("[ParentActor] message: %+v\n", msg)
		return nil
	}, goactor.WithInitFunc(func(context.Context, *goactor.PID) error {
		_ = goactor.SetTrapExit(true)
		_ = goactor.Link(child)
		return nil
	}))
	_, err := goactor.Spawn(ctx, parent)
	if err != nil {
		panic(err)
	}
	err = goactor.Send(ctx, child, "go to sleep")
	if err != nil {
		panic(err)
	}

	<-ctx.Done()
	fmt.Println("[!] Sleeping done, exiting in main func")
}
