package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hedisam/goactor"
)

func main() {
	fmt.Println("----- Links & Monitors ----")
	ctx := context.Background()

	child := goactor.Spawn(ctx, func(ctx context.Context, msg any) (loop bool, err error) {
		fmt.Printf("[!] ChildActor: %+v; Sleeping a bit then error\n", msg)
		time.Sleep(time.Second)
		return false, errors.New("i'm child actor, got nothing to do so exit with an error")
	})

	parent := goactor.Spawn(ctx, func(ctx context.Context, msg any) (loop bool, err error) {
		_, ok := goactor.IsSystemMessage(msg)
		if ok {
			fmt.Printf("[!] ParentActor received system message: %+v\n", msg)
			return true, nil
		}
		fmt.Printf("[!] ParentActor: %+v\n", msg)
		return true, nil
	})

	parent.Link(child, true)
	err := goactor.Send(ctx, child, "go to sleep")
	if err != nil {
		panic(err)
	}

	fmt.Printf("[!] Sleeping for a bit in the main func\n")
	time.Sleep(5 * time.Second)
	fmt.Printf("[!] Sleeping done, exiting in main func\n")
}
