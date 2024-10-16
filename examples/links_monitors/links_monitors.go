package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/sysmsg"
)

func main() {
	fmt.Println("----- Links & Monitors ----")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	child, _ := goactor.Spawn(ctx, goactor.NewActor(func(ctx context.Context, msg any) (loop bool, err error) {
		fmt.Printf("[ChildActor] message: %+v\n", msg)
		fmt.Printf("[ChildActor] sleeping for 1s then will error\n")
		time.Sleep(time.Second)
		return false, errors.New("got nothing to do so exit with an error")
	}))

	parent := goactor.NewActor(func(ctx context.Context, msg any) (loop bool, err error) {
		if processID, reason, ok := sysmsg.LinkedActorDown(msg); ok {
			fmt.Printf("[ParentActor] Linked actor %q terminated with reason %q\n", processID, reason)
			return true, nil
		}
		fmt.Printf("[ParentActor] message: %+v\n", msg)
		return true, nil
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
