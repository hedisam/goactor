package main

import (
	"context"
	"fmt"
	"time"

	"github.com/hedisam/goactor"
	"github.com/hedisam/goactor/examples/require"
	"github.com/hedisam/goactor/sysmsg"
)

func main() {
	fmt.Println("----- Links & Monitors ----")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	child, _ := goactor.Spawn(ctx, goactor.ReceiveFunc(func(ctx context.Context, msg any) error {
		fmt.Printf("[ChildActor] message: %+v\n", msg)
		fmt.Printf("[ChildActor] sleeping for a bit then will go down\n")
		time.Sleep(time.Millisecond * 100)
		return sysmsg.ReasonShutdown // stop with a shutdown reason
	}))

	_, err := goactor.Spawn(ctx, &Parent{
		child: child,
	})
	require.NoError(err)
	err = goactor.Send(ctx, child, "go to sleep")
	require.NoError(err)

	<-ctx.Done()
	fmt.Println("[!] Sleeping done, exiting main func")
}

type Parent struct {
	child *goactor.PID
}

func (p *Parent) Receive(_ context.Context, msg any) error {
	sysMsg, ok := sysmsg.ToMessage(msg)
	switch {
	case ok && sysMsg.Type == sysmsg.Exit:
		fmt.Printf("[ParentActor] Linked actor %q terminated with reason %q\n", sysMsg.ProcessID, sysMsg.Reason)
		return nil
	default:
		fmt.Printf("[ParentActor] message: %+v\n", msg)
		return nil
	}
}

func (p *Parent) Init(context.Context, *goactor.PID) error {
	_ = goactor.SetTrapExit(true)
	_ = goactor.Link(p.child)
	return nil
}
